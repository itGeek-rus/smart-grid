package rest_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itGeek-rus/smart-grid.git/internal/transport/rest"
)

func TestHealthz(t *testing.T) {
	router := rest.NewRouter("smart-grid-processor", "local", nil)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestReadyz_OK(t *testing.T) {
	router := rest.NewRouter("smart-grid-api", "local", nil,
		rest.ReadyFunc(func(context.Context) error { return nil }),
	)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	router.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestReadyz_NotReady(t *testing.T) {
	router := rest.NewRouter("smart-grid-api", "local", nil,
		rest.ReadyFunc(func(context.Context) error {
			return errors.New("db down")
		}),
	)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	router.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}
