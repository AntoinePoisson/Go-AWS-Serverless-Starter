package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/config"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx/middleware"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/item"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/item/mock_item"
	"github.com/AntoinePoisson/go-aws-serverless-starter/lambda/api/internal/handler/items"
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

// the one request worth reading is the one that crashed. we lose it if
// Recover wraps Logger
func TestNewHandlerLogsARequestThatPanics(t *testing.T) {
	service := mock_item.NewMockServiceAPI(gomock.NewController(t))
	service.EXPECT().List(gomock.Any(), gomock.Any()).DoAndReturn(
		func(context.Context, int32) ([]item.Item, error) { panic("boom") })

	// same decorator main() installs via SetupLogging, that's how the
	// request id lands on the records
	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(middleware.LogRequestID(slog.NewJSONHandler(&buf, nil))))
	t.Cleanup(func() { slog.SetDefault(previous) })

	handler, err := newHandler(&config.Config{APIKey: "secret"}, items.New(service))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	req.Header.Set(middleware.APIKeyHeader, "secret")

	rec := httptest.NewRecorder()
	assert.NotPanics(t, func() { handler.ServeHTTP(rec, req) })
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var request map[string]any
	for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		var entry map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		if entry["msg"] == "request" {
			request = entry
		}
	}

	require.NotNil(t, request, "a request that panicked must still be logged")
	assert.EqualValues(t, http.StatusInternalServerError, request["status"])
	assert.NotEmpty(t, request["request_id"], "with its request id")
}
