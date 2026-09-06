package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx"
)

// Recover turns a panic into a 500 instead of taking the process down.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := wrapResponseWriter(w)

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

			// already writing, don't append more garbage on top
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
