package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/itGeek-rus/smart-grid.git/internal/pkg/metrics"
)

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(service string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{w, http.StatusOK}
		next.ServeHTTP(sw, r)

		path := r.Pattern
		if path == "" {
			path = r.URL.Path
		}
		status := strconv.Itoa(sw.code)
		metrics.HTTPRequests.WithLabelValues(service, r.Method, path, status).Inc()
		metrics.HTTPDuration.WithLabelValues(service, r.Method, path).Observe(time.Since(start).Seconds())
	})
}
