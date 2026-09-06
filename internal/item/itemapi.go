package item

import "context"

// Repository is the persistence side.
type Repository interface {
	Put(ctx context.Context, in *Item) error
	Get(ctx context.Context, id string) (*Item, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit int32) ([]Item, error)
}

// ServiceAPI is what the handlers call.
type ServiceAPI interface {
	Create(ctx context.Context, input CreateInput) (*Item, error)
	Get(ctx context.Context, id string) (*Item, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit int32) ([]Item, error)
}
