package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// Logger writes one line per request, after the handler returns.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := wrapResponseWriter(w)

		next.ServeHTTP(rw, r)

		level := slog.LevelInfo
		if rw.status >= http.StatusInternalServerError {
			level = slog.LevelError
		}

		// Leave request_id to LogRequestID. Adding it here prints the key twice.
		slog.Log(r.Context(), level, "request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
