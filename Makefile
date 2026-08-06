TASK ?= $(shell command -v task 2>/dev/null || echo 'go run github.com/go-task/task/v3/cmd/task@v3.52.0')

.DEFAULT_GOAL := help
.PHONY: help test fmt lint tidy

help: ## Show the available targets
	@grep -hE '^[a-z0-9-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

test: ## Run the unit tests with coverage
	@$(TASK) test

fmt: ## Format the Go sources
	@$(TASK) fmt

lint: ## Run golangci-lint
	@$(TASK) lint

tidy: ## Sync go.mod and go.sum
	@$(TASK) tidy
