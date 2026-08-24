// Package middleware is the HTTP middleware the functions share.
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDHeader is how the request id travels in and out.
const RequestIDHeader = "X-Request-Id"

type requestIDKey struct{}

// RequestID assigns an id, reuses the incoming one if present, and echoes it.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set(RequestIDHeader, id)

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

// RequestIDFrom returns the id on ctx, or "".
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}
