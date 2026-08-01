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

	kafkarepo "github.com/itGeek-rus/smart-grid.git/internal/repository/kafka"
	"github.com/itGeek-rus/smart-grid.git/internal/repository/postgres"
	redisrepo "github.com/itGeek-rus/smart-grid.git/internal/repository/redis"
	"github.com/itGeek-rus/smart-grid.git/internal/repository/timescaledb"
	"github.com/itGeek-rus/smart-grid.git/internal/service/processor"

	"github.com/itGeek-rus/smart-grid.git/internal/config"
	"github.com/itGeek-rus/smart-grid.git/internal/pkg/logger"
	"github.com/itGeek-rus/smart-grid.git/internal/transport/rest"
)

type App struct {
	cfg      config.Config
	log      *slog.Logger
	server   *http.Server
	pool     interface{ Close() }
	cache    *redisrepo.Cache
	producer *kafkarepo.Producer
	consumer *kafkarepo.Consumer
	svc      *processor.Service
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log := logger.New(cfg.App.LogLevel).With(
		slog.String("service", cfg.App.Name),
		slog.String("env", cfg.App.Env),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := timescaledb.NewPool(ctx, cfg.Postgres.DSN())
	if err != nil {
		return nil, err
	}
	cache := redisrepo.NewCache(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err := cache.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("redis: %w", err)
	}

	producer := kafkarepo.NewProducer(cfg.Kafka)
	telemetryRepo := timescaledb.NewTelemetryRepo(pool)
	alertRepo := postgres.NewAlertRepo(pool)
	svc := processor.NewService(telemetryRepo, alertRepo, cache, producer, log)
	consumer := kafkarepo.NewConsumer(cfg.Kafka, cfg.Kafka.TopicRawTelemetry, log)

	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           rest.NewRouter(cfg.App.Name, cfg.App.Env).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		cfg:      cfg,
		log:      log,
		server:   server,
		pool:     pool,
		cache:    cache,
		producer: producer,
		consumer: consumer,
		svc:      svc,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 2)

	go func() {
		a.log.Info("processor http starting", slog.String("addr", a.cfg.HTTP.Addr))
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	go func() {
		a.log.Info("processor kafka consumer starting", slog.String("topic", a.cfg.Kafka.TopicRawTelemetry))
		if err := a.consumer.Subscribe(ctx, nil, "", a.svc.HandleRawMessage); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		a.log.Info("shutdown signal received")
	case err := <-errCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.App.ShutdownTimeout)
	defer cancel()

	_ = a.server.Shutdown(shutdownCtx)
	_ = a.consumer.Close()
	_ = a.producer.Close()
	_ = a.cache.Close()
	a.pool.Close()
	return nil
}
