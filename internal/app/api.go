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
	"github.com/itGeek-rus/smart-grid.git/internal/repository/postgres"
	redisrepo "github.com/itGeek-rus/smart-grid.git/internal/repository/redis"
	"github.com/itGeek-rus/smart-grid.git/internal/repository/timescaledb"
	"github.com/itGeek-rus/smart-grid.git/internal/transport/rest"
	"github.com/itGeek-rus/smart-grid.git/internal/usecase"
)

type API struct {
	cfg      config.Config
	log      *slog.Logger
	server   *http.Server
	pool     interface{ Close() }
	cache    *redisrepo.Cache
	producer *kafkarepo.Producer
}

func NewAPI() (*API, error) {
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

	uc := usecase.NewAPIUseCase(
		postgres.NewDeviceRepo(pool),
		timescaledb.NewTelemetryRepo(pool),
		postgres.NewAlertRepo(pool),
		cache,
		producer,
	)
	apiHandler := rest.NewAPIHandler(uc)

	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           rest.NewRouter(cfg.App.Name, cfg.App.Env, apiHandler).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &API{
		cfg:      cfg,
		log:      log,
		server:   server,
		pool:     pool,
		cache:    cache,
		producer: producer,
	}, nil
}

func (a *API) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		a.log.Info("api http starting", slog.String("addr", a.cfg.HTTP.Addr))
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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.cfg.App.ShutdownTimeout)
	defer shutdownCancel()

	_ = a.server.Shutdown(shutdownCtx)
	_ = a.producer.Close()
	_ = a.cache.Close()
	a.pool.Close()
	a.log.Info("api shutdown completed")
	return nil
}
