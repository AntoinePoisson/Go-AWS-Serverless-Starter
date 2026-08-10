// Package awsx builds the AWS SDK clients used by the service.
package awsx

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/wire"

	"github.com/antoinepoisson/bootstrap-go-aws/internal/config"
)

// WireSet provides the AWS clients to the dependency graph.
var WireSet = wire.NewSet(NewDynamoDBClient)

// NewDynamoDBClient returns a DynamoDB client. When DYNAMODB_ENDPOINT is set it
// targets that endpoint, which is how the service talks to DynamoDB Local.
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
