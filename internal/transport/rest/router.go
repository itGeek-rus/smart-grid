package rest

import (
	"encoding/json"
	"net/http"
	"time"
)

type Router struct {
	mux     *http.ServeMux
	started time.Time
	appName string
	appEnv  string
}

func NewRouter(appName string, appEnv string) *Router {
	r := &Router{
		mux:     http.NewServeMux(),
		started: time.Now().UTC(),
		appName: appName,
		appEnv:  appEnv,
	}
	r.registerRoutes()
	return r
}

func (r *Router) Handler() http.Handler {
	return r.mux
}

func (r *Router) registerRoutes() {
	r.mux.HandleFunc("GET /healthz", r.healthz)
	r.mux.HandleFunc("GET /readyz", r.readyz)
}

func (r *Router) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"started": "ok",
		"service": r.appName,
		"env":     r.appEnv,
		"uptime":  time.Since(r.started).String(),
	})
}

func (r *Router) readyz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
