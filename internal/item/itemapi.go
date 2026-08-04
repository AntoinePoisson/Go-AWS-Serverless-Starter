package item

import "context"

// Repository persists items.
type Repository interface {
	Put(ctx context.Context, in *Item) error
	Get(ctx context.Context, id string) (*Item, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit int32) ([]Item, error)
}

// ServiceAPI exposes the item use cases to the handlers.
type ServiceAPI interface {
	Create(ctx context.Context, input CreateInput) (*Item, error)
	Get(ctx context.Context, id string) (*Item, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit int32) ([]Item, error)
}
