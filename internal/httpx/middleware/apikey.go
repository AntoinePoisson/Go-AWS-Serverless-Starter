package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/antoinepoisson/bootstrap-go-aws/internal/httpx"
)

// APIKeyHeader carries the shared secret expected by APIKey.
//
//nolint:gosec // G101 matches the name of the header, not a credential.
const APIKeyHeader = "X-Api-Key"

// APIKey rejects requests whose API key header does not match key.
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
