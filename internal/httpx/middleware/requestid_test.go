package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/antoinepoisson/bootstrap-go-aws/internal/httpx/middleware"
)

func TestRequestIDGeneratesIdentifier(t *testing.T) {
	var fromContext string

	handler := middleware.RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		fromContext = middleware.RequestIDFrom(r.Context())
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items", nil))

	assert.NotEmpty(t, fromContext)
	assert.Equal(t, fromContext, rec.Header().Get(middleware.RequestIDHeader))
}

func TestRequestIDReusesIncomingIdentifier(t *testing.T) {
	var fromContext string

	handler := middleware.RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		fromContext = middleware.RequestIDFrom(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	req.Header.Set(middleware.RequestIDHeader, "known-id")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "known-id", fromContext)
	assert.Equal(t, "known-id", rec.Header().Get(middleware.RequestIDHeader))
}

func TestRequestIDFromEmptyContext(t *testing.T) {
	assert.Empty(t, middleware.RequestIDFrom(t.Context()))
}
