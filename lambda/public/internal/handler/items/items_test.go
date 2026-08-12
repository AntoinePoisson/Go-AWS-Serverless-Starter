package items_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/antoinepoisson/bootstrap-go-aws/internal/handlertest"
	"github.com/antoinepoisson/bootstrap-go-aws/internal/httpx"
	"github.com/antoinepoisson/bootstrap-go-aws/internal/item"
	"github.com/antoinepoisson/bootstrap-go-aws/internal/item/mock_item"
	"github.com/antoinepoisson/bootstrap-go-aws/lambda/public/internal/handler/items"
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
	handlertest.Run(t, []handlertest.Case{
		{
			Name: "get returns the item",
			Path: "/public/items/42",
			NewHandler: newHandler(func(s *mock_item.MockServiceAPI) {
				s.EXPECT().Get(gomock.Any(), "42").Return(&item.Item{ID: "42", Name: "demo"}, nil)
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
			Path: "/public/items/unknown",
			NewHandler: newHandler(func(s *mock_item.MockServiceAPI) {
				s.EXPECT().Get(gomock.Any(), "unknown").Return(nil, item.ErrNotFound)
			}),
			WantStatus: http.StatusNotFound,
		},
		{
			Name:       "writes are not exposed",
			Method:     http.MethodDelete,
			Path:       "/public/items/42",
			NewHandler: newHandler(nil),
			WantStatus: http.StatusMethodNotAllowed,
		},
	})
}
