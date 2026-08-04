package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/antoinepoisson/bootstrap-go-aws/internal/config"
	"github.com/antoinepoisson/bootstrap-go-aws/internal/httpx/middleware"
	"github.com/antoinepoisson/bootstrap-go-aws/internal/item/mock_item"
	"github.com/antoinepoisson/bootstrap-go-aws/lambda/api/internal/handler/items"
)

func TestNewHandlerRequiresAnAPIKey(t *testing.T) {
	_, err := newHandler(&config.Config{}, items.New(nil))

	assert.Error(t, err, "starting without a key would expose the API")
}

func TestNewHandlerRejectsAnUnauthenticatedRequest(t *testing.T) {
	handler, err := newHandler(&config.Config{APIKey: "secret"}, items.New(nil))
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items", nil))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.NotEmpty(t, rec.Header().Get(middleware.RequestIDHeader), "the request id middleware must run")
}

func TestNewHandlerServesAnAuthenticatedRequest(t *testing.T) {
	service := mock_item.NewMockServiceAPI(gomock.NewController(t))
	service.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, nil)

	handler, err := newHandler(&config.Config{APIKey: "secret"}, items.New(service))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	req.Header.Set(middleware.APIKeyHeader, "secret")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
