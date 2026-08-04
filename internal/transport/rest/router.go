package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/itGeek-rus/smart-grid.git/internal/pkg/metrics"
)

type Router struct {
	mux      *http.ServeMux
	started  time.Time
	appName  string
	appEnv   string
	checkers []ReadyChecker
}

func NewRouter(appName, appEnv string, api *APIHandler, checkers ...ReadyChecker) *Router {
	r := &Router{
		mux:      http.NewServeMux(),
		started:  time.Now().UTC(),
		appName:  appName,
		appEnv:   appEnv,
		checkers: checkers,
	}
	r.registerRoutes()
	if api != nil {
		api.Register(r.mux)
	}
	return r
}

func (r *Router) Handler() http.Handler {
	return MetricsMiddleware(r.appName, r.mux)
}

func (r *Router) registerRoutes() {
	r.mux.HandleFunc("GET /healthz", r.healthz)
	r.mux.HandleFunc("GET /readyz", r.readyz)
	r.mux.Handle("GET /metrics", metrics.Handler())
}

func (r *Router) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": r.appName,
		"env":     r.appEnv,
		"uptime":  time.Since(r.started).String(),
	})
}

func (r *Router) readyz(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	for _, c := range r.checkers {
		if err := c.Ready(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "not_ready",
				"error":  err.Error(),
			})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
