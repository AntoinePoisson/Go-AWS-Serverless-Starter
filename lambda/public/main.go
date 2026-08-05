// Command public serves the unauthenticated endpoints.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/config"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx/middleware"
	"github.com/AntoinePoisson/go-aws-serverless-starter/lambda/public/internal/handler/health"
	"github.com/AntoinePoisson/go-aws-serverless-starter/lambda/public/internal/handler/items"
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
	// identifier. Logger wraps Recover, not the reverse: a panic unwinding
	// through Logger skips the line it writes after the handler returns.
	return httpx.Chain(
		httpx.NewRouter(healthHandler, itemsHandler),
		middleware.RequestID,
		middleware.Logger,
		middleware.Recover,
	)
}
