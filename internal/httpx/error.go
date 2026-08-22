package httpx

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// Error is an error carrying the status code to return to the client.
type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code" validate:"required" example:"not_found"`
	Message string `json:"message" validate:"required" example:"item not found"`
}

// Errorf builds an Error with a formatted message.
func Errorf(status int, code, format string, args ...any) *Error {
	return &Error{
		Status:  status,
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// HandlerFunc is an http.HandlerFunc allowed to return an error.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// Handle adapts a HandlerFunc to http.HandlerFunc and renders its errors.
func Handle(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			WriteError(w, r, err)
		}
	}
}

// WriteError renders err as JSON. Anything that is not an *Error gets logged
// and reported as an internal error.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var httpErr *Error
	if !errors.As(err, &httpErr) {
		slog.ErrorContext(r.Context(), "unhandled error",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path,
		)
		httpErr = &Error{
			Status:  http.StatusInternalServerError,
			Code:    "internal_error",
			Message: "internal server error",
		}
	}
	WriteJSON(w, httpErr.Status, httpErr)
}
