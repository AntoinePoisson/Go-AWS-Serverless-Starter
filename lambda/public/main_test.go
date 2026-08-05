package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/config"
	"github.com/AntoinePoisson/go-aws-serverless-starter/lambda/public/internal/handler/health"
	"github.com/AntoinePoisson/go-aws-serverless-starter/lambda/public/internal/handler/items"
)

func TestNewHandlerServesHealthWithoutAKey(t *testing.T) {
	handler := newHandler(health.New(&config.Config{Stage: "alpha"}), items.New(nil))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"stage":"alpha"`)
}

func TestNewHandlerDoesNotExposeTheAuthenticatedRoutes(t *testing.T) {
	handler := newHandler(health.New(&config.Config{}), items.New(nil))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items", nil))

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
