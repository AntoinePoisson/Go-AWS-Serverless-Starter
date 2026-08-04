// Command public serves the unauthenticated endpoints.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/antoinepoisson/bootstrap-go-aws/internal/config"
	"github.com/antoinepoisson/bootstrap-go-aws/internal/httpx"
	"github.com/antoinepoisson/bootstrap-go-aws/internal/httpx/middleware"
	"github.com/antoinepoisson/bootstrap-go-aws/lambda/public/internal/handler/health"
	"github.com/antoinepoisson/bootstrap-go-aws/lambda/public/internal/handler/items"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load configuration", "error", err)
		os.Exit(1)
	}
	cfg.SetupLogging(middleware.LogRequestID)

	handler, err := inject(cfg)
	if err != nil {
		slog.Error("build application", "error", err)
		os.Exit(1)
	}

	if err := httpx.Serve(handler, cfg.ListenAddr); err != nil {
		slog.Error("serve", "error", err)
		os.Exit(1)
	}
}

// newHandler assembles the router and the middleware chain.
func newHandler(healthHandler *health.Handler, itemsHandler *items.Handler) http.Handler {
	// RequestID comes first so everything logged under it carries the
	// identifier, a recovered panic included.
	return httpx.Chain(
		httpx.NewRouter(healthHandler, itemsHandler),
		middleware.RequestID,
		middleware.Recover,
		middleware.Logger,
	)
}
