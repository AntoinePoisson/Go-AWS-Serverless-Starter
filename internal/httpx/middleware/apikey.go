package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx"
)

// APIKeyHeader is the header APIKey reads.
//
//nolint:gosec // G101 hits the header name, not a real secret.
const APIKeyHeader = "X-Api-Key"

// APIKey rejects a request when the key is wrong.
func APIKey(key string) httpx.Middleware {
	expected := []byte(key)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := []byte(r.Header.Get(APIKeyHeader))
			if len(got) == 0 || subtle.ConstantTimeCompare(got, expected) != 1 {
				httpx.WriteError(w, r, httpx.Errorf(http.StatusUnauthorized, "unauthorized", "missing or invalid API key"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
