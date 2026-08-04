package item

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxNameLength = 200
	maxTags       = 20
	defaultLimit  = 25
	maxLimit      = 100
)

var (
	// ErrNotFound is returned when an item does not exist.
	ErrNotFound = errors.New("item not found")
	// ErrInvalidInput is returned when a payload fails validation.
	ErrInvalidInput = errors.New("invalid input")
)

var _ ServiceAPI = (*Service)(nil)

// CreateInput is the payload accepted by Service.Create.
type CreateInput struct {
	Name     string            `json:"name"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
}

// Service implements the item use cases on top of a Repository.
type Service struct {
	repo  Repository
	now   func() time.Time
	newID func() string
}

// NewService returns a Service backed by repo.
func NewService(repo Repository) *Service {
	return &Service{
		repo:  repo,
		now:   time.Now,
		newID: uuid.NewString,
	}
}

// Create validates the input and stores a new item.
func (s *Service) Create(ctx context.Context, input CreateInput) (*Item, error) {
	name := strings.TrimSpace(input.Name)
	switch {
	case name == "":
		return nil, fmt.Errorf("%w: name is required", ErrInvalidInput)
	case len(name) > maxNameLength:
		return nil, fmt.Errorf("%w: name must be at most %d characters", ErrInvalidInput, maxNameLength)
	case len(input.Tags) > maxTags:
		return nil, fmt.Errorf("%w: at most %d tags are allowed", ErrInvalidInput, maxTags)
	}

	item := &Item{
		ID:        s.newID(),
		Name:      name,
		Tags:      input.Tags,
		Metadata:  input.Metadata,
		CreatedAt: s.now().UTC().Truncate(time.Second),
	}
	if err := s.repo.Put(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// Get returns the item with the given id.
func (s *Service) Get(ctx context.Context, id string) (*Item, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	return s.repo.Get(ctx, id)
}

// Delete removes the item with the given id.
func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	return s.repo.Delete(ctx, id)
}

// List returns up to limit items. A limit outside [1, 100] falls back to 25.
func (s *Service) List(ctx context.Context, limit int32) ([]Item, error) {
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}
	return s.repo.List(ctx, limit)
}
