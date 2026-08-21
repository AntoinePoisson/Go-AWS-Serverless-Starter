// Command localdb creates the items table in a local DynamoDB.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/awsx"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/config"
)

// `docker compose up -d` returns as soon as the container is started, a second
// or two before DynamoDB Local accepts a connection. The SDK retries a network
// error for about a second, which is not enough, so the first call waits here.
const (
	startupTimeout = 30 * time.Second
	retryInterval  = 500 * time.Millisecond
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

	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	switch err := createTable(ctx, client, cfg.ItemsTable); {
	case errors.Is(err, errTableExists):
		log.Printf("table %s already exists", cfg.ItemsTable)
	case err != nil:
		log.Printf("create table %s: %v", cfg.ItemsTable, err)
		os.Exit(1)
	default:
		log.Printf("table %s created", cfg.ItemsTable)
	}
}

var errTableExists = errors.New("table already exists")

// createTable keeps trying until DynamoDB Local answers or ctx runs out, and
// reports the last failure rather than the deadline.
func createTable(ctx context.Context, client *dynamodb.Client, table string) error {
	input := &dynamodb.CreateTableInput{
		TableName:   aws.String(table),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
	}

	var waited bool

	for {
		_, err := client.CreateTable(ctx, input)

		var inUse *types.ResourceInUseException
		switch {
		case errors.As(err, &inUse):
			return errTableExists
		case err == nil:
			return nil
		case ctx.Err() != nil:
			return err
		}

		if !waited {
			waited = true
			log.Print("waiting for DynamoDB Local to accept connections")
		}

		select {
		case <-ctx.Done():
			return err
		case <-time.After(retryInterval):
		}
	}
}
