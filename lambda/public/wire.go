//go:build wireinject

package main

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/wire"

	"github.com/antoinepoisson/bootstrap-go-aws/internal/awsx"
	"github.com/antoinepoisson/bootstrap-go-aws/internal/config"
	"github.com/antoinepoisson/bootstrap-go-aws/internal/item"
	"github.com/antoinepoisson/bootstrap-go-aws/lambda/public/internal/handler/health"
	"github.com/antoinepoisson/bootstrap-go-aws/lambda/public/internal/handler/items"
)

func inject(cfg *config.Config) (http.Handler, error) {
	wire.Build(
		awsx.WireSet,
		wire.Bind(new(item.DynamoDBAPI), new(*dynamodb.Client)),
		item.WireSet,
		items.WireSet,
		health.WireSet,
		newHandler,
	)
	return nil, nil
}
