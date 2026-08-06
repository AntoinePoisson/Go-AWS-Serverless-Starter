package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/antoinepoisson/bootstrap-go-aws/internal/config"
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
