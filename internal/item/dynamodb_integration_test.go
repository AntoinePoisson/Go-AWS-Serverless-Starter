//go:build integration

package item_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/awsx"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/config"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/item"
)

const integrationTable = "items-integration"

func newIntegrationRepository(t *testing.T) *item.DynamoDBRepository {
	t.Helper()

	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8000"
	}

	t.Setenv("AWS_REGION", "eu-west-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "local")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "local")

	cfg := &config.Config{ItemsTable: integrationTable, DynamoDBEndpoint: endpoint}
	client, err := awsx.NewDynamoDBClient(cfg)
	require.NoError(t, err)

	createTable(t, client)

	return item.NewDynamoDBRepository(client, cfg)
}

func createTable(t *testing.T, client *dynamodb.Client) {
	t.Helper()

	_, err := client.CreateTable(t.Context(), &dynamodb.CreateTableInput{
		TableName:   aws.String(integrationTable),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
	})

	var inUse *types.ResourceInUseException
	if err != nil && !errors.As(err, &inUse) {
		require.NoError(t, err, "DynamoDB Local must be reachable, run: make local-db")
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTable(context.WithoutCancel(t.Context()), &dynamodb.DeleteTableInput{
			TableName: aws.String(integrationTable),
		})
	})
}

func TestRepositoryLifecycle(t *testing.T) {
	repo := newIntegrationRepository(t)
	ctx := t.Context()

	stored := &item.Item{
		ID:        "integration-1",
		Name:      "demo",
		Tags:      []string{"starter"},
		Metadata:  map[string]string{"owner": "team"},
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	require.NoError(t, repo.Put(ctx, stored))

	got, err := repo.Get(ctx, stored.ID)
	require.NoError(t, err)
	assert.Equal(t, stored, got)

	listed, err := repo.List(ctx, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, listed)

	require.NoError(t, repo.Delete(ctx, stored.ID))

	_, err = repo.Get(ctx, stored.ID)
	assert.ErrorIs(t, err, item.ErrNotFound)

	assert.ErrorIs(t, repo.Delete(ctx, stored.ID), item.ErrNotFound)
}
