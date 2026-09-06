// Package config loads runtime settings from the environment.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config is the settings every function reads.
type Config struct {
	Stage            string `env:"STAGE" envDefault:"local"`
	LogLevel         string `env:"LOG_LEVEL" envDefault:"info"`
	ItemsTable       string `env:"ITEMS_TABLE,required,notEmpty"`
	APIKey           string `env:"API_KEY"`
	DynamoDBEndpoint string `env:"DYNAMODB_ENDPOINT"`
	ListenAddr       string `env:"LISTEN_ADDR" envDefault:":8080"`
	Version          string `env:"VERSION" envDefault:"dev"`
}

// Load reads the environment.
func Load() (*Config, error) {
	var c Config
	if err := env.Parse(&c); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return &c, nil
}

// SetupLogging installs a JSON slog handler. Decorators wrap whatever we
// just built, so the request id can land on every record without this
// package knowing about HTTP.
func (c *Config) SetupLogging(decorators ...func(slog.Handler) slog.Handler) {
	var handler slog.Handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: c.slogLevel(),
	})
	for _, decorate := range decorators {
		handler = decorate(handler)
	}
	slog.SetDefault(slog.New(handler))
}

func (c *Config) slogLevel() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
