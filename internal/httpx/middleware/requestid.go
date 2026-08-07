// Package middleware holds the HTTP middlewares applied to the functions.
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDHeader carries the request identifier in and out.
const RequestIDHeader = "X-Request-Id"

type requestIDKey struct{}

// RequestID assigns an identifier to the request, reusing the incoming header
// when present, and echoes it in the response.
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

// RequestIDFrom returns the identifier carried by ctx, or an empty string.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}
