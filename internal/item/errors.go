package item

import (
	"errors"
	"net/http"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx"
)

// HTTPError maps a domain error to HTTP. Anything else is left alone and
// becomes a 500 further up.
func HTTPError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrNotFound):
		return httpx.Errorf(http.StatusNotFound, "not_found", "%s", err)
	case errors.Is(err, ErrInvalidInput):
		return httpx.Errorf(http.StatusBadRequest, "invalid_input", "%s", err)
	default:
		return err
	}
}
