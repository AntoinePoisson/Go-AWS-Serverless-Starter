TASK ?= $(shell command -v task 2>/dev/null || echo 'go run github.com/go-task/task/v3/cmd/task@v3.52.0')
STAGE ?= alpha

.DEFAULT_GOAL := help
.PHONY: help build build-host test test-integration fmt lint tidy wire mocks docs docs-check package deploy deploy-fn remove db db-stop run-api run-public e2e clean

help: ## Show the available targets
	@grep -hE '^[a-z0-9-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

build: ## Build and package every function for Lambda
	@$(TASK) build:dist

build-host: ## Build every function for the host platform
	@$(TASK) build:host

test: ## Run the unit tests with coverage
	@$(TASK) test

test-integration: ## Run the integration tests against DynamoDB Local
	@$(TASK) test:integration

fmt: ## Format the Go sources and the OpenAPI annotations
	@$(TASK) fmt

lint: ## Run golangci-lint
	@$(TASK) lint

tidy: ## Sync go.mod and go.sum
	@$(TASK) tidy

wire: ## Regenerate the Wire injectors
	@$(TASK) wire:injectors

mocks: ## Regenerate the mocks
	@$(TASK) wire:mocks

docs: ## Regenerate the OpenAPI specification and the rendered reference
	@$(TASK) docs:html

docs-check: ## Fail if the committed OpenAPI specification is out of date
	@$(TASK) docs:check

package: ## Build the CloudFormation template without deploying
	@$(TASK) deploy:package STAGE=$(STAGE)

deploy: ## Deploy the stack (make deploy STAGE=preprod)
	@$(TASK) deploy:stack STAGE=$(STAGE)

deploy-fn: ## Deploy a single function (make deploy-fn NAME=api)
	@$(TASK) deploy:function NAME=$(NAME) STAGE=$(STAGE)

remove: ## Delete the deployed stack
	@$(TASK) deploy:remove STAGE=$(STAGE)

db: ## Start DynamoDB Local and create the items table
	@$(TASK) local:db

db-stop: ## Stop DynamoDB Local
	@$(TASK) local:db:stop

run-api: ## Run the api function on port 8080
	@$(TASK) local:api

run-public: ## Run the public function on port 8081
	@$(TASK) local:public

e2e: ## Run the end-to-end tests
	@cd e2e && npm test

clean: ## Remove every build and test artifact
	@$(TASK) clean:all
