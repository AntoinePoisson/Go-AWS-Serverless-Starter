package item

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/config"
)

var _ Repository = (*DynamoDBRepository)(nil)

// DynamoDBRepository stores items in one table, keyed by id.
type DynamoDBRepository struct {
	client DynamoDBAPI
	table  string
}

// NewDynamoDBRepository points at the configured table.
func NewDynamoDBRepository(client DynamoDBAPI, cfg *config.Config) *DynamoDBRepository {
	return &DynamoDBRepository{client: client, table: cfg.ItemsTable}
}

// Put writes the item. Same id overwrites.
func (r *DynamoDBRepository) Put(ctx context.Context, item *Item) error {
	attributes, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("marshal item: %w", err)
	}

	if _, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      attributes,
	}); err != nil {
		return fmt.Errorf("put item %s: %w", item.ID, err)
	}
	return nil
}

// Get returns the item, or ErrNotFound.
func (r *DynamoDBRepository) Get(ctx context.Context, id string) (*Item, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key:       primaryKey(id),
	})
	if err != nil {
		return nil, fmt.Errorf("get item %s: %w", id, err)
	}
	if len(out.Item) == 0 {
		return nil, ErrNotFound
	}

	var item Item
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, fmt.Errorf("unmarshal item %s: %w", id, err)
	}
	return &item, nil
}

// Delete removes the item, or returns ErrNotFound.
func (r *DynamoDBRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName:           aws.String(r.table),
		Key:                 primaryKey(id),
		ConditionExpression: aws.String("attribute_exists(id)"),
	})

	var conditionFailed *types.ConditionalCheckFailedException
	switch {
	case errors.As(err, &conditionFailed):
		return ErrNotFound
	case err != nil:
		return fmt.Errorf("delete item %s: %w", id, err)
	}
	return nil
}

// List returns up to limit items, no particular order.
func (r *DynamoDBRepository) List(ctx context.Context, limit int32) ([]Item, error) {
	out, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.table),
		Limit:     aws.Int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("scan items: %w", err)
	}

	items := make([]Item, 0, len(out.Items))
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &items); err != nil {
		return nil, fmt.Errorf("unmarshal items: %w", err)
	}
	return items, nil
}

func primaryKey(id string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: id}}
}
