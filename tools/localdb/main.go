// Command localdb creates the items table in a local DynamoDB instance.
package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/awsx"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.DynamoDBEndpoint == "" {
		log.Fatal("DYNAMODB_ENDPOINT must point to a local DynamoDB instance")
	}

	client, err := awsx.NewDynamoDBClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.CreateTable(context.Background(), &dynamodb.CreateTableInput{
		TableName:   aws.String(cfg.ItemsTable),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
	})

	var inUse *types.ResourceInUseException
	switch {
	case errors.As(err, &inUse):
		log.Printf("table %s already exists", cfg.ItemsTable)
	case err != nil:
		log.Printf("create table %s: %v", cfg.ItemsTable, err)
		os.Exit(1)
	default:
		log.Printf("table %s created", cfg.ItemsTable)
	}
}
