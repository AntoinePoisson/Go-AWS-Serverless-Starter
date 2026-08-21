package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx/middleware"
)

// captureContextLog installs the decorated handler, runs h behind the RequestID
// middleware and returns the single record it wrote.
func captureContextLog(t *testing.T, h http.Handler) map[string]any {
	t.Helper()

	var buf bytes.Buffer
	previous := slog.Default()
	base := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(middleware.LogRequestID(base)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	req.Header.Set(middleware.RequestIDHeader, "req-42")
	middleware.RequestID(h).ServeHTTP(httptest.NewRecorder(), req)

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	return entry
}

func TestLogRequestIDReachesAHandlerRecord(t *testing.T) {
	entry := captureContextLog(t, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		slog.ErrorContext(r.Context(), "unhandled error", "error", "boom")
	}))

	assert.Equal(t, "unhandled error", entry["msg"])
	assert.Equal(t, "req-42", entry["request_id"])
}

func TestLogRequestIDReachesARecoveredPanic(t *testing.T) {
	entry := captureContextLog(t, middleware.Recover(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	})))

	assert.Equal(t, "panic recovered", entry["msg"])
	assert.Equal(t, "req-42", entry["request_id"], "Recover must sit under RequestID")
}

func TestLogRequestIDSkipsARecordWithoutARequest(t *testing.T) {
	var buf bytes.Buffer
	base := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(middleware.LogRequestID(base)).With("stage", "local")

	logger.Info("listening", "addr", ":8080")

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))

	assert.Equal(t, "local", entry["stage"], "With must keep the decoration in place")
	assert.NotContains(t, entry, "request_id")
}
