package config_test

import (
	"strings"
	"testing"

	"github.com/itGeek-rus/smart-grid.git/internal/config"
)

func TestLoad_RequiredPostgresDSN(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://smartgrid:smartgrid@localhost:5432/smartgrid?sslmode=disable")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.App.Name == "" {
		t.Fatal("expected non-empty app name")
	}
	if cfg.HTTP.Addr == "" {
		t.Fatal("expected non-empty http addr")
	}
}

func TestLoad_AndDSN(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "smartgrid")
	t.Setenv("DB_PASSWORD", "smartgrid")
	t.Setenv("DB_NAME", "smartgrid")
	t.Setenv("DB_SSLMODE", "disable")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.App.Name == "" {
		t.Fatal("expected non-empty app name")
	}
	dsn := cfg.Postgres.DSN()
	if !strings.Contains(dsn, "smartgrid") {
		t.Fatalf("unexpected dsn: %s", dsn)
	}
	if !strings.Contains(dsn, "sslmode=disable") {
		t.Fatalf("expected sslmode in dsn: %s", dsn)
	}
}
