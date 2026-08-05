package item_test

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/config"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/item"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/item/mock_item"
)

func newRepository(t *testing.T) (*item.DynamoDBRepository, *mock_item.MockDynamoDBAPI) {
	t.Helper()

	client := mock_item.NewMockDynamoDBAPI(gomock.NewController(t))
	return item.NewDynamoDBRepository(client, &config.Config{ItemsTable: "items"}), client
}

func TestRepositoryPut(t *testing.T) {
	repo, client := newRepository(t)

	var input *dynamodb.PutItemInput
	client.EXPECT().PutItem(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ any, in *dynamodb.PutItemInput, _ ...any) (*dynamodb.PutItemOutput, error) {
			input = in
			return &dynamodb.PutItemOutput{}, nil
		})

	err := repo.Put(t.Context(), &item.Item{ID: "42", Name: "demo", CreatedAt: time.Unix(0, 0).UTC()})
	require.NoError(t, err)

	assert.Equal(t, "items", *input.TableName)
	assert.Equal(t, &types.AttributeValueMemberS{Value: "42"}, input.Item["id"])
	assert.Equal(t, &types.AttributeValueMemberS{Value: "demo"}, input.Item["name"])
}

func TestRepositoryGet(t *testing.T) {
	repo, client := newRepository(t)

	client.EXPECT().GetItem(gomock.Any(), gomock.Any()).Return(&dynamodb.GetItemOutput{
		Item: map[string]types.AttributeValue{
			"id":   &types.AttributeValueMemberS{Value: "42"},
			"name": &types.AttributeValueMemberS{Value: "demo"},
		},
	}, nil)

	got, err := repo.Get(t.Context(), "42")
	require.NoError(t, err)

	assert.Equal(t, "42", got.ID)
	assert.Equal(t, "demo", got.Name)
}

func TestRepositoryGetMissingItem(t *testing.T) {
	repo, client := newRepository(t)

	client.EXPECT().GetItem(gomock.Any(), gomock.Any()).Return(&dynamodb.GetItemOutput{}, nil)

	_, err := repo.Get(t.Context(), "42")

	assert.ErrorIs(t, err, item.ErrNotFound)
}

func TestRepositoryGetFailure(t *testing.T) {
	repo, client := newRepository(t)

	client.EXPECT().GetItem(gomock.Any(), gomock.Any()).Return(nil, errors.New("network down"))

	_, err := repo.Get(t.Context(), "42")

	require.Error(t, err)
	assert.NotErrorIs(t, err, item.ErrNotFound)
}

func TestRepositoryDelete(t *testing.T) {
	repo, client := newRepository(t)

	var input *dynamodb.DeleteItemInput
	client.EXPECT().DeleteItem(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ any, in *dynamodb.DeleteItemInput, _ ...any) (*dynamodb.DeleteItemOutput, error) {
			input = in
			return &dynamodb.DeleteItemOutput{}, nil
		})

	require.NoError(t, repo.Delete(t.Context(), "42"))

	assert.Equal(t, "attribute_exists(id)", *input.ConditionExpression)
}

func TestRepositoryDeleteMissingItem(t *testing.T) {
	repo, client := newRepository(t)

	client.EXPECT().DeleteItem(gomock.Any(), gomock.Any()).
		Return(nil, &types.ConditionalCheckFailedException{})

	err := repo.Delete(t.Context(), "42")

	assert.ErrorIs(t, err, item.ErrNotFound)
}

func TestRepositoryList(t *testing.T) {
	repo, client := newRepository(t)

	var input *dynamodb.ScanInput
	client.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ any, in *dynamodb.ScanInput, _ ...any) (*dynamodb.ScanOutput, error) {
			input = in
			return &dynamodb.ScanOutput{Items: []map[string]types.AttributeValue{
				{"id": &types.AttributeValueMemberS{Value: "1"}},
				{"id": &types.AttributeValueMemberS{Value: "2"}},
			}}, nil
		})

	got, err := repo.List(t.Context(), 10)
	require.NoError(t, err)

	assert.Equal(t, int32(10), *input.Limit)
	assert.Len(t, got, 2)
	assert.Equal(t, "1", got[0].ID)
}

func TestRepositoryListEmptyTable(t *testing.T) {
	repo, client := newRepository(t)

	client.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(&dynamodb.ScanOutput{}, nil)

	got, err := repo.List(t.Context(), 10)
	require.NoError(t, err)

	assert.NotNil(t, got)
	assert.Empty(t, got)
}
