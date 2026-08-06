# bootstrap-go-aws

A starting point for a serverless API in Go: API Gateway HTTP API, two Lambda
functions, one DynamoDB table.

The goal is a repository that already answers the boring questions — how the
handlers are wired, where the configuration comes from, what runs before a
push — so that the first day of a project is spent on the resource, not on the
plumbing.

## Getting started

Needs Go 1.25.

```sh
make test
make lint
```

## Layout

```
internal/  the reusable packages
lambda/    one directory per function
tasks/     the go-task files behind the Makefile
```
