package httpx_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx"
)

func TestWriteErrorRendersHTTPError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/items/1", nil)

	httpx.WriteError(rec, req, httpx.Errorf(http.StatusNotFound, "not_found", "item %s not found", "1"))

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.JSONEq(t, `{"code":"not_found","message":"item 1 not found"}`, rec.Body.String())
}

func TestWriteErrorHidesUnknownError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/items/1", nil)

	httpx.WriteError(rec, req, errors.New("connection refused to internal host"))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "connection refused")
	assert.JSONEq(t, `{"code":"internal_error","message":"internal server error"}`, rec.Body.String())
}

func TestWriteErrorUnwrapsWrappedError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	wrapped := fmt.Errorf("list items: %w", httpx.Errorf(http.StatusBadRequest, "invalid_limit", "limit must be an integer"))

	httpx.WriteError(rec, req, wrapped)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleRendersReturnedError(t *testing.T) {
	handler := httpx.Handle(func(http.ResponseWriter, *http.Request) error {
		return httpx.Errorf(http.StatusTeapot, "teapot", "short and stout")
	})

	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusTeapot, rec.Code)
}
