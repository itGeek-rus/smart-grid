package rest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itGeek-rus/smart-grid.git/internal/transport/rest"
)

func TestHealthz(t *testing.T) {
	router := rest.NewRouter("smart-grid-processor", "local")
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
