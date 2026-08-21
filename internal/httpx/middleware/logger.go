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
		rw := wrapResponseWriter(w)

		next.ServeHTTP(rw, r)

		level := slog.LevelInfo
		if rw.status >= http.StatusInternalServerError {
			level = slog.LevelError
		}

		// No request_id here: the record carries the context, and LogRequestID
		// puts the id on every record that does. Adding it twice writes the key
		// twice in the same JSON object.
		slog.Log(r.Context(), level, "request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
