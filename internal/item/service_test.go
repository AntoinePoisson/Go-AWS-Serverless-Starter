package item_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/item"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/item/mock_item"
)

func TestServiceCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock_item.NewMockRepository(ctrl)

	var stored *item.Item
	repo.EXPECT().Put(gomock.Any(), gomock.Any()).DoAndReturn(func(_ any, i *item.Item) error {
		stored = i
		return nil
	})

	created, err := item.NewService(repo).Create(t.Context(), item.CreateInput{
		Name:     "  demo  ",
		Tags:     []string{"starter"},
		Metadata: map[string]string{"owner": "team"},
	})
	require.NoError(t, err)

	assert.Equal(t, "demo", created.Name)
	assert.NotEmpty(t, created.ID)
	assert.WithinDuration(t, time.Now(), created.CreatedAt, time.Minute)
	assert.Equal(t, []string{"starter"}, created.Tags)
	assert.Equal(t, created, stored)
}

func TestServiceCreateRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name  string
		input item.CreateInput
	}{
		{name: "empty name", input: item.CreateInput{Name: ""}},
		{name: "blank name", input: item.CreateInput{Name: "   "}},
		{name: "name too long", input: item.CreateInput{Name: strings.Repeat("a", 201)}},
		{name: "name too long in runes", input: item.CreateInput{Name: strings.Repeat("é", 201)}},
		{name: "too many tags", input: item.CreateInput{Name: "demo", Tags: make([]string, 21)}},
		// these would also get refused by DynamoDB. better we say 400 here
		// than let it come back as a 500.
		{
			name:  "tag too long",
			input: item.CreateInput{Name: "demo", Tags: []string{strings.Repeat("a", 51)}},
		},
		{
			name:  "too many metadata entries",
			input: item.CreateInput{Name: "demo", Metadata: manyEntries(21)},
		},
		{
			name: "metadata key too long",
			input: item.CreateInput{
				Name:     "demo",
				Metadata: map[string]string{strings.Repeat("k", 101): "value"},
			},
		},
		{
			name: "metadata value too long",
			input: item.CreateInput{
				Name:     "demo",
				Metadata: map[string]string{"owner": strings.Repeat("v", 1001)},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mock_item.NewMockRepository(ctrl)

			_, err := item.NewService(repo).Create(t.Context(), c.input)

			assert.ErrorIs(t, err, item.ErrInvalidInput)
		})
	}
}

func manyEntries(n int) map[string]string {
	entries := make(map[string]string, n)
	for i := range n {
		entries[strconv.Itoa(i)] = "value"
	}
	return entries
}

// If we say it fits, it has to go through.
func TestServiceCreateAcceptsThePayloadAtEveryBound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock_item.NewMockRepository(ctrl)

	repo.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil)

	tags := make([]string, 20)
	for i := range tags {
		tags[i] = strings.Repeat("a", 50)
	}

	metadata := manyEntries(20)
	metadata["0"] = strings.Repeat("v", 1000)

	_, err := item.NewService(repo).Create(t.Context(), item.CreateInput{
		Name:     strings.Repeat("a", 200),
		Tags:     tags,
		Metadata: metadata,
	})

	require.NoError(t, err)
}

func TestServiceCreateMeasuresTheNameInRunes(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock_item.NewMockRepository(ctrl)

	repo.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil)

	name := strings.Repeat("é", 200) // 200 characters, 400 bytes

	created, err := item.NewService(repo).Create(t.Context(), item.CreateInput{Name: name})
	require.NoError(t, err)

	assert.Equal(t, name, created.Name)
}

func TestServiceCreatePropagatesRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock_item.NewMockRepository(ctrl)
	failure := errors.New("throughput exceeded")

	repo.EXPECT().Put(gomock.Any(), gomock.Any()).Return(failure)

	_, err := item.NewService(repo).Create(t.Context(), item.CreateInput{Name: "demo"})

	assert.ErrorIs(t, err, failure)
}

func TestServiceGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock_item.NewMockRepository(ctrl)
	stored := &item.Item{ID: "42", Name: "demo"}

	repo.EXPECT().Get(gomock.Any(), "42").Return(stored, nil)

	got, err := item.NewService(repo).Get(t.Context(), "42")
	require.NoError(t, err)

	assert.Equal(t, stored, got)
}

func TestServiceGetRejectsEmptyID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock_item.NewMockRepository(ctrl)

	_, err := item.NewService(repo).Get(t.Context(), " ")

	assert.ErrorIs(t, err, item.ErrInvalidInput)
}

func TestServiceDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock_item.NewMockRepository(ctrl)

	repo.EXPECT().Delete(gomock.Any(), "42").Return(item.ErrNotFound)

	err := item.NewService(repo).Delete(t.Context(), "42")

	assert.ErrorIs(t, err, item.ErrNotFound)
}

func TestServiceListClampsLimit(t *testing.T) {
	cases := []struct {
		name      string
		limit     int32
		wantLimit int32
	}{
		{name: "zero falls back to default", limit: 0, wantLimit: 25},
		{name: "negative falls back to default", limit: -5, wantLimit: 25},
		{name: "above maximum falls back to default", limit: 500, wantLimit: 25},
		{name: "within range is kept", limit: 10, wantLimit: 10},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mock_item.NewMockRepository(ctrl)

			repo.EXPECT().List(gomock.Any(), c.wantLimit).Return([]item.Item{}, nil)

			_, err := item.NewService(repo).List(t.Context(), c.limit)
			require.NoError(t, err)
		})
	}
}
