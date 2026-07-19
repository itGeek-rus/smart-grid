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
	"github.com/itGeek-rus/smart-grid.git/internal/transport/rest"
)

type App struct {
	cfg    config.Config
	log    *slog.Logger
	server *http.Server
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

	router := rest.NewRouter(cfg.App.Name, cfg.App.Env)

	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           router.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		cfg:    cfg,
		log:    log,
		server: server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		a.log.Info("http server starting", slog.String("addr", a.cfg.HTTP.Addr))
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

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	a.log.Info("shutdown completed")
	return nil
}
