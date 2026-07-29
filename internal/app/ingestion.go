package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/itGeek-rus/smart-grid.git/internal/config"
	"github.com/itGeek-rus/smart-grid.git/internal/pkg/logger"
	kafkarepo "github.com/itGeek-rus/smart-grid.git/internal/repository/kafka"
	mqttrepo "github.com/itGeek-rus/smart-grid.git/internal/repository/mqtt"
	"github.com/itGeek-rus/smart-grid.git/internal/service/ingestion"
	"github.com/itGeek-rus/smart-grid.git/internal/transport/rest"
)

type IngestionApp struct {
	cfg      config.Config
	log      *slog.Logger
	server   *http.Server
	mqtt     *mqttrepo.Client
	producer *kafkarepo.Producer
	svc      *ingestion.Service
}

func NewIngestion() (*IngestionApp, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log := logger.New(cfg.App.LogLevel).With(
		slog.String("service", cfg.App.Name),
		slog.String("env", cfg.App.Env),
	)

	producer := kafkarepo.NewProducer(cfg.Kafka)
	svc := ingestion.NewService(producer, cfg.MQTT.TopicTelemetry, log)

	mqttClient, err := mqttrepo.NewClient(cfg.MQTT.BrokerURL, cfg.MQTT.ClientID, log)
	if err != nil {
		_ = producer.Close()
		return nil, err
	}

	router := rest.NewRouter(cfg.App.Name, cfg.App.Env)
	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           router.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &IngestionApp{
		cfg:      cfg,
		log:      log,
		server:   server,
		mqtt:     mqttClient,
		producer: producer,
		svc:      svc,
	}, nil
}

func (a *IngestionApp) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := a.mqtt.Subscribe(ctx, a.cfg.MQTT.TopicTelemetry, a.cfg.MQTT.Qos, a.svc.HandleMQTT); err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		a.log.Info("ingestion http starting", slog.String("addr", a.cfg.HTTP.Addr))
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		a.log.Info("shutdown signal received")
	case err := <-errCh:
		return fmt.Errorf("http server failed: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.App.ShutdownTimeout)
	defer cancel()

	_ = a.server.Shutdown(shutdownCtx)
	_ = a.mqtt.Close()
	_ = a.producer.Close()
	a.log.Info("ingestion shutdown completed")
	return nil
}
