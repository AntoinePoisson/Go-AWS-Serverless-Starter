// Package awsx builds the AWS clients we use.
package awsx

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/wire"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/config"
)

// WireSet is the AWS clients.
var WireSet = wire.NewSet(NewDynamoDBClient)

// NewDynamoDBClient returns a DynamoDB client. Hits DYNAMODB_ENDPOINT when
// set (DynamoDB Local).
func NewDynamoDBClient(cfg *config.Config) (*dynamodb.Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
		if cfg.DynamoDBEndpoint != "" {
			o.BaseEndpoint = &cfg.DynamoDBEndpoint
		}
	}), nil
}
