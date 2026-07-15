package item

import (
	"errors"
	"net/http"

	"github.com/antoinepoisson/bootstrap-go-aws/internal/httpx"
)

// HTTPError maps a service error to its HTTP representation. Errors that are
// not part of the domain are returned untouched and reported as a 500.
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
