//go:build wireinject

package main

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/wire"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/awsx"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/config"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/item"
	"github.com/AntoinePoisson/go-aws-serverless-starter/lambda/public/internal/handler/health"
	"github.com/AntoinePoisson/go-aws-serverless-starter/lambda/public/internal/handler/items"
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
