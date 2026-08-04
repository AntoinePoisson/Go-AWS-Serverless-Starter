package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/antoinepoisson/bootstrap-go-aws/internal/httpx/middleware"
)

func TestAPIKey(t *testing.T) {
	cases := []struct {
		name       string
		header     string
		wantStatus int
		wantCalled bool
	}{
		{name: "valid key", header: "secret", wantStatus: http.StatusOK, wantCalled: true},
		{name: "wrong key", header: "nope", wantStatus: http.StatusUnauthorized},
		{name: "missing key", header: "", wantStatus: http.StatusUnauthorized},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			called := false
			handler := middleware.APIKey("secret")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				called = true
			}))

			req := httptest.NewRequest(http.MethodGet, "/items", nil)
			if c.header != "" {
				req.Header.Set(middleware.APIKeyHeader, c.header)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, c.wantStatus, rec.Code)
			assert.Equal(t, c.wantCalled, called)
		})
	}
}

func TestAPIKeyRejectsEveryRequestWhenKeyIsEmpty(t *testing.T) {
	handler := middleware.APIKey("")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler must not be reached")
	}))

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	req.Header.Set(middleware.APIKeyHeader, "anything")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
