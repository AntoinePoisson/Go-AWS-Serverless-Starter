package config_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/config"
)

func TestLoadAppliesDefaults(t *testing.T) {
	t.Setenv("ITEMS_TABLE", "items")

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "items", cfg.ItemsTable)
	assert.Equal(t, "local", cfg.Stage)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, ":8080", cfg.ListenAddr)
	assert.Empty(t, cfg.DynamoDBEndpoint)
}

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv("ITEMS_TABLE", "starter-prod-items")
	t.Setenv("STAGE", "prod")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("API_KEY", "secret")
	t.Setenv("DYNAMODB_ENDPOINT", "http://localhost:8000")

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "starter-prod-items", cfg.ItemsTable)
	assert.Equal(t, "prod", cfg.Stage)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "secret", cfg.APIKey)
	assert.Equal(t, "http://localhost:8000", cfg.DynamoDBEndpoint)
}

func TestLoadRequiresItemsTable(t *testing.T) {
	t.Setenv("ITEMS_TABLE", "")

	_, err := config.Load()
	assert.Error(t, err)
}

func TestSetupLoggingAppliesTheConfiguredLevel(t *testing.T) {
	cases := map[string]struct {
		level      string
		wantDebug  bool
		wantErrors bool
	}{
		"debug":       {level: "debug", wantDebug: true, wantErrors: true},
		"info":        {level: "info", wantDebug: false, wantErrors: true},
		"warning":     {level: "warning", wantDebug: false, wantErrors: true},
		"error":       {level: "error", wantDebug: false, wantErrors: true},
		"mixed case":  {level: "DEBUG", wantDebug: true, wantErrors: true},
		"unknown one": {level: "chatty", wantDebug: false, wantErrors: true},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			restoreDefaultLogger(t)

			cfg := &config.Config{LogLevel: c.level}
			cfg.SetupLogging()

			assert.Equal(t, c.wantDebug, slog.Default().Enabled(t.Context(), slog.LevelDebug))
			assert.Equal(t, c.wantErrors, slog.Default().Enabled(t.Context(), slog.LevelError))
		})
	}
}

// The decorators are how the request id reaches every record, see
// middleware.LogRequestID.
func TestSetupLoggingAppliesTheDecoratorsOutwards(t *testing.T) {
	restoreDefaultLogger(t)

	var applied []string
	decorate := func(name string) func(slog.Handler) slog.Handler {
		return func(h slog.Handler) slog.Handler {
			applied = append(applied, name)
			return h
		}
	}

	cfg := &config.Config{}
	cfg.SetupLogging(decorate("first"), decorate("second"))

	assert.Equal(t, []string{"first", "second"}, applied)
}

func restoreDefaultLogger(t *testing.T) {
	t.Helper()
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
}
