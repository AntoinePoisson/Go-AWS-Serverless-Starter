package items_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/handlertest"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/item"
	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/item/mock_item"
	"github.com/AntoinePoisson/go-aws-serverless-starter/lambda/api/internal/handler/items"
)

func newHandler(expect func(*mock_item.MockServiceAPI)) func(*gomock.Controller) httpx.Handler {
	return func(ctrl *gomock.Controller) httpx.Handler {
		service := mock_item.NewMockServiceAPI(ctrl)
		if expect != nil {
			expect(service)
		}
		return items.New(service)
	}
}

func TestHandler(t *testing.T) {
	stored := &item.Item{ID: "42", Name: "demo"}

	handlertest.Run(t, []handlertest.Case{
		{
			Name:   "create returns the stored item",
			Method: http.MethodPost,
			Path:   "/items",
			Body:   `{"name":"demo"}`,
			NewHandler: newHandler(func(s *mock_item.MockServiceAPI) {
				s.EXPECT().Create(gomock.Any(), item.CreateInput{Name: "demo"}).Return(stored, nil)
			}),
			WantStatus: http.StatusCreated,
			Check: func(t *testing.T, res *http.Response) {
				var got item.Item
				handlertest.DecodeBody(t, res, &got)

				assert.Equal(t, "42", got.ID)
				assert.Equal(t, "/items/42", res.Header.Get("Location"))
			},
		},
		{
			Name:       "create rejects a malformed body",
			Method:     http.MethodPost,
			Path:       "/items",
			Body:       `{"name":`,
			NewHandler: newHandler(nil),
			WantStatus: http.StatusBadRequest,
		},
		{
			Name:   "create reports a validation failure",
			Method: http.MethodPost,
			Path:   "/items",
			Body:   `{"name":""}`,
			NewHandler: newHandler(func(s *mock_item.MockServiceAPI) {
				s.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, item.ErrInvalidInput)
			}),
			WantStatus: http.StatusBadRequest,
		},
		{
			Name:   "create hides an unexpected failure",
			Method: http.MethodPost,
			Path:   "/items",
			Body:   `{"name":"demo"}`,
			NewHandler: newHandler(func(s *mock_item.MockServiceAPI) {
				s.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("table missing"))
			}),
			WantStatus: http.StatusInternalServerError,
			Check: func(t *testing.T, res *http.Response) {
				var got httpx.Error
				handlertest.DecodeBody(t, res, &got)

				assert.Equal(t, "internal_error", got.Code)
			},
		},
		{
			Name: "get returns the item",
			Path: "/items/42",
			NewHandler: newHandler(func(s *mock_item.MockServiceAPI) {
				s.EXPECT().Get(gomock.Any(), "42").Return(stored, nil)
			}),
			WantStatus: http.StatusOK,
			Check: func(t *testing.T, res *http.Response) {
				var got item.Item
				handlertest.DecodeBody(t, res, &got)

				assert.Equal(t, "demo", got.Name)
			},
		},
		{
			Name: "get reports a missing item",
			Path: "/items/unknown",
			NewHandler: newHandler(func(s *mock_item.MockServiceAPI) {
				s.EXPECT().Get(gomock.Any(), "unknown").Return(nil, item.ErrNotFound)
			}),
			WantStatus: http.StatusNotFound,
		},
		{
			Name:   "delete returns no content",
			Method: http.MethodDelete,
			Path:   "/items/42",
			NewHandler: newHandler(func(s *mock_item.MockServiceAPI) {
				s.EXPECT().Delete(gomock.Any(), "42").Return(nil)
			}),
			WantStatus: http.StatusNoContent,
		},
		{
			Name:   "delete reports a missing item",
			Method: http.MethodDelete,
			Path:   "/items/unknown",
			NewHandler: newHandler(func(s *mock_item.MockServiceAPI) {
				s.EXPECT().Delete(gomock.Any(), "unknown").Return(item.ErrNotFound)
			}),
			WantStatus: http.StatusNotFound,
		},
		{
			Name: "list returns the items and their count",
			Path: "/items",
			NewHandler: newHandler(func(s *mock_item.MockServiceAPI) {
				s.EXPECT().List(gomock.Any(), int32(0)).Return([]item.Item{*stored}, nil)
			}),
			WantStatus: http.StatusOK,
			Check: func(t *testing.T, res *http.Response) {
				var got struct {
					Items []item.Item `json:"items"`
					Count int         `json:"count"`
				}
				handlertest.DecodeBody(t, res, &got)

				assert.Equal(t, 1, got.Count)
				assert.Len(t, got.Items, 1)
			},
		},
		{
			Name: "list forwards the limit",
			Path: "/items?limit=5",
			NewHandler: newHandler(func(s *mock_item.MockServiceAPI) {
				s.EXPECT().List(gomock.Any(), int32(5)).Return(nil, nil)
			}),
			WantStatus: http.StatusOK,
			Check: func(t *testing.T, res *http.Response) {
				var got map[string]any
				handlertest.DecodeBody(t, res, &got)

				assert.Equal(t, []any{}, got["items"], "an empty list must serialise as [], not null")
			},
		},
		{
			Name:       "list rejects a non numeric limit",
			Path:       "/items?limit=many",
			NewHandler: newHandler(nil),
			WantStatus: http.StatusBadRequest,
		},
	})
}
