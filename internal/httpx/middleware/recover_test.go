package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx/middleware"
)

func TestRecoverTurnsPanicIntoInternalError(t *testing.T) {
	handler := middleware.Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items", nil))
	})

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, `{"code":"internal_error","message":"internal server error"}`, rec.Body.String())
	assert.NotContains(t, rec.Body.String(), "boom")
}

func TestRecoverLeavesAStartedResponseAlone(t *testing.T) {
	handler := middleware.Recover(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[`))
		panic("boom")
	}))

	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items", nil))
	})

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, `{"items":[`, rec.Body.String(), "an error document must not be appended to a started response")
}

func TestRecoverLeavesSuccessfulResponseUntouched(t *testing.T) {
	handler := middleware.Recover(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items", nil))

	assert.Equal(t, http.StatusAccepted, rec.Code)
}
