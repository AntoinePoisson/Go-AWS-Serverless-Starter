package item

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Limits are in runes, not bytes, so accents shouldnt eat half the budget.
// Worst case stays under 100 KiB, well inside DynamoDB's 400 KiB. Without this
// a body the HTTP limit accepts gets rejected by the table and looks like a 500.
const (
	maxNameLength          = 200
	maxTags                = 20
	maxTagLength           = 50
	maxMetadataEntries     = 20
	maxMetadataKeyLength   = 100
	maxMetadataValueLength = 1000

	defaultLimit = 25
	maxLimit     = 100
)

var (
	// ErrNotFound means the item is gone, or never existed.
	ErrNotFound = errors.New("item not found")
	// ErrInvalidInput is a bad payload, not a server problem.
	ErrInvalidInput = errors.New("invalid input")
)

var _ ServiceAPI = (*Service)(nil)

// CreateInput is what Create accepts.
type CreateInput struct {
	// Name is required, trimmed, 200 chars max.
	Name string `json:"name" validate:"required,max=200" example:"first item"`
	// Tags, 20 max, 50 chars each.
	Tags []string `json:"tags" validate:"max=20" maxLength:"50" example:"demo,starter"`
	// Metadata, 20 entries max. Keys 100 chars, values 1000.
	Metadata map[string]string `json:"metadata" example:"owner:platform"`
}

// Service is the item use cases on top of a Repository.
type Service struct {
	repo  Repository
	now   func() time.Time
	newID func() string
}

// NewService wraps repo.
func NewService(repo Repository) *Service {
	return &Service{
		repo:  repo,
		now:   time.Now,
		newID: uuid.NewString,
	}
}

// Create checks the payload then stores it.
func (s *Service) Create(ctx context.Context, input CreateInput) (*Item, error) {
	name := strings.TrimSpace(input.Name)
	if err := validate(name, input); err != nil {
		return nil, err
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

// validate stops at the first problem. name is already trimmed.
func validate(name string, input CreateInput) error {
	switch {
	case name == "":
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	case utf8.RuneCountInString(name) > maxNameLength:
		return fmt.Errorf("%w: name must be at most %d characters", ErrInvalidInput, maxNameLength)
	case len(input.Tags) > maxTags:
		return fmt.Errorf("%w: at most %d tags are allowed", ErrInvalidInput, maxTags)
	case len(input.Metadata) > maxMetadataEntries:
		return fmt.Errorf("%w: at most %d metadata entries are allowed", ErrInvalidInput, maxMetadataEntries)
	}

	for _, tag := range input.Tags {
		if utf8.RuneCountInString(tag) > maxTagLength {
			return fmt.Errorf("%w: a tag must be at most %d characters", ErrInvalidInput, maxTagLength)
		}
	}

	for key, value := range input.Metadata {
		switch {
		case utf8.RuneCountInString(key) > maxMetadataKeyLength:
			return fmt.Errorf("%w: metadata key %q must be at most %d characters",
				ErrInvalidInput, key, maxMetadataKeyLength)
		case utf8.RuneCountInString(value) > maxMetadataValueLength:
			return fmt.Errorf("%w: the metadata value of %q must be at most %d characters",
				ErrInvalidInput, key, maxMetadataValueLength)
		}
	}

	return nil
}

// Get looks up one item.
func (s *Service) Get(ctx context.Context, id string) (*Item, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	return s.repo.Get(ctx, id)
}

// Delete removes one item.
func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	return s.repo.Delete(ctx, id)
}

// List returns up to limit items. Outside [1, 100] we fall back to 25.
func (s *Service) List(ctx context.Context, limit int32) ([]Item, error) {
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}
	return s.repo.List(ctx, limit)
}
