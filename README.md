# bootstrap-go-aws

A serverless API in Go you can deploy on the first day and still trust on the
hundredth: API Gateway HTTP API, two Lambda functions, one DynamoDB table.

- **Runs anywhere without an emulator.** Both functions are plain `net/http`
  servers; `internal/httpx` adapts the same binary to the Lambda runtime.
- **Wired, not sketched.** Routing, middlewares (request id, structured logs,
  recovery, API key), Wire dependency injection, DynamoDB repository, unit,
  integration and end-to-end tests.
- **A pipeline that says no.** Lint, tests, e2e, rendered infrastructure and a
  drift check on every generated file — a red check stops the deploy on every
  stage.
- **Three stages, three branches, one command.** Conventional commits, Release
  Please, deploy first and release after.
- **Documentation that cannot drift.** `docs/openapi.yaml` is generated from the
  annotations and CI fails when the committed file is stale.

```
                     API Gateway (HTTP API, payload 2.0)
                                    |
                +-------------------+-------------------+
          api (Go, arm64)                        public (Go, arm64)
          X-Api-Key required                     no authentication
                +-------------------+-------------------+
                                    |
                           DynamoDB table (items)
```

## Getting started

Needs Go 1.25, Node.js 22 and Docker.

```sh
make init          # make the template yours, then install and start everything
make run-api       # in one shell: api on :8080
make run-public    # in another:   public on :8081
```

```sh
curl -s -X POST localhost:8080/items \
  -H 'X-Api-Key: local-dev-key' \
  -d '{"name":"first item","tags":["demo"]}'
```

`make init` asks for a Go module path and a service name, rewrites both across
the repository, then resolves the dependencies, installs the git hooks and
creates the local table. It is the only step that knows about the template —
everything after it is your project. Run it directly to skip the prompts:

```sh
./initialize.sh --module github.com/acme/orders-api --service orders-api
```

The `items` resource is an example: a small CRUD that exercises every layer end
to end. Replace it with your own resource and the plumbing stays.

## Commands

```sh
make help              # list every target
make test              # unit tests, race detector and coverage
make test-integration  # repository against DynamoDB Local
make e2e               # Playwright, against the running functions
make fmt lint          # one config for both, .golangci.yml
make docs              # regenerate docs/openapi.yaml and docs/api.html
make wire mocks        # regenerate the injectors and the mocks
make deploy STAGE=alpha
```

Every route carries swag annotations; run `make docs` after touching one and
commit `docs/openapi.yaml` with the change. A route also lives in
`serverless/function-*.js` — the `infra` CI job catches what the two disagree
on.

## Layout

```
lambda/api        authenticated CRUD, Wire injector, middleware chain
lambda/public     health and read-only endpoints
internal/config   environment-backed settings
internal/httpx    router, JSON responses, errors, middlewares, Lambda adapter
internal/item     model, DynamoDB repository, use cases
serverless/       function definitions, DynamoDB table, stage files
tasks/            build, deploy, local, codegen and documentation tasks
e2e/              Playwright tests
```

## Configuration

Every setting comes from the environment. Local runs and integration tests read
the committed `.env.example`, then `.env` if it exists; a real environment
variable wins over both. Deployment tasks read neither, so nothing local can
reach a deployed stage.

| Variable            | Default  | Description |
| ------------------- | -------- | ----------- |
| `ITEMS_TABLE`       | required | DynamoDB table name |
| `API_KEY`           | —        | Secret expected by `api` in `X-Api-Key`; deployed stages read it from SSM |
| `STAGE`             | `local`  | Reported by `/health` |
| `LOG_LEVEL`         | `info`   | `debug`, `info`, `warn` or `error` |
| `LISTEN_ADDR`       | `:8080`  | Local listen address, ignored on Lambda |
| `DYNAMODB_ENDPOINT` | —        | Set to target DynamoDB Local |
| `VERSION`           | `dev`    | Reported by `/health` |

## Stages and releases

One branch per stage. A push deploys once every check passes, and the release
job only runs after a successful deploy, so a tag can never point at code that
was never shipped.

| Branch | Deploys to | Releases |
| ------ | ---------- | -------- |
| `dev`  | `alpha`    | — |
| `main` | `preprod`  | prereleases `vX.Y.Z-pre.N` |
| `prod` | `prod`     | stable `vX.Y.Z` |

Stages are declared in `serverless/stage/`; add a file there to add one.
`alpha` is throwaway — verbose logs, CORS open, table dropped with the stack.
`preprod` mirrors `prod`, so a release is rehearsed under the same constraints.

Before the first deploy, create the API key each stage reads from SSM:

```sh
aws ssm put-parameter --name /bootstrap-go-aws/alpha/api-key \
  --type SecureString --value "$(openssl rand -hex 32)"
```

CI authenticates through OIDC: set the repository variable `AWS_REGION` and the
secret `AWS_DEPLOY_ROLE_ARN`. Production approval belongs in the `prod` GitHub
Environment as a required reviewer.

Two things to know before your first release: commits must follow
[Conventional Commits](https://www.conventionalcommits.org), which `commitlint`
enforces in a hook; and auto-merge must stay off on the release pull requests,
because a push authenticated with `GITHUB_TOKEN` triggers no workflow — the
tagged version would never deploy.

Hooks are installed by `npm install`: formatting and lint on commit,
regeneration, build and tests on push.
