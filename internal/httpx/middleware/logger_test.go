package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx/middleware"
)

func captureLog(t *testing.T, status int) map[string]any {
	t.Helper()

	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previous) })

	handler := middleware.Logger(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/items", nil))

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	return entry
}

func TestLoggerRecordsTheRequest(t *testing.T) {
	entry := captureLog(t, http.StatusOK)

	assert.Equal(t, "request", entry["msg"])
	assert.Equal(t, "INFO", entry["level"])
	assert.Equal(t, http.MethodGet, entry["method"])
	assert.Equal(t, "/items", entry["path"])
	assert.EqualValues(t, http.StatusOK, entry["status"])
}

func TestLoggerRaisesTheLevelOnServerErrors(t *testing.T) {
	assert.Equal(t, "ERROR", captureLog(t, http.StatusInternalServerError)["level"])
	assert.Equal(t, "INFO", captureLog(t, http.StatusNotFound)["level"])
}

// LogRequestID already puts the id on every record. If Logger adds it too
// the key shows up twice. Still valid JSON, annoying to query.
func TestLoggerLeavesTheRequestIDToTheDecorator(t *testing.T) {
	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(middleware.LogRequestID(slog.NewJSONHandler(&buf, nil))))
	t.Cleanup(func() { slog.SetDefault(previous) })

	handler := middleware.RequestID(middleware.Logger(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) },
	)))

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	req.Header.Set(middleware.RequestIDHeader, "req-42")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	line := buf.String()
	assert.Equal(t, 1, strings.Count(line, `"request_id"`), "the key must be written once")
	assert.Contains(t, line, `"request_id":"req-42"`)
}
