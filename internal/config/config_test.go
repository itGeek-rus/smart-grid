package config_test

import (
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
