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

// Every bound is in runes and not in bytes, accents shouldnt halve the limit.
// Together they keep the worst case item under 100 KiB, well inside the 400 KiB
// a DynamoDB item is allowed: without them a payload the body limit accepts
// reaches the table, gets rejected there and reads as a server error when it is
// the caller who sent too much.
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

// Service implements the item use cases over a Repository.
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

// validate reports the first thing wrong with a creation payload, if anything.
// name comes trimmed, the caller is the one that stores it.
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

// List returns up to limit items. Anything outside [1, 100] falls back to 25.
func (s *Service) List(ctx context.Context, limit int32) ([]Item, error) {
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}
	return s.repo.List(ctx, limit)
}
