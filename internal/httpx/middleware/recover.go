package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx"
)

// Recover turns a panic into a 500 response instead of killing the process.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := newResponseWriter(w)

		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			slog.ErrorContext(r.Context(), "panic recovered",
				"panic", recovered,
				"method", r.Method,
				"path", r.URL.Path,
				"stack", string(debug.Stack()),
			)

			// Appending an error document to a started response would only
			// corrupt what the client is reading.
			if rw.written {
				return
			}

			httpx.WriteJSON(rw, http.StatusInternalServerError, &httpx.Error{
				Code:    "internal_error",
				Message: "internal server error",
			})
		}()

		next.ServeHTTP(rw, r)
	})
}
