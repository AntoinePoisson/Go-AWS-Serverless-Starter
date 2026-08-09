package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// Logger writes one structured line per request, after the response.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		next.ServeHTTP(rw, r)

		level := slog.LevelInfo
		if rw.status >= http.StatusInternalServerError {
			level = slog.LevelError
		}

		slog.Log(r.Context(), level, "request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", RequestIDFrom(r.Context()),
		)
	})
}
