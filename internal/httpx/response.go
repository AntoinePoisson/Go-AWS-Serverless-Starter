package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
)

const maxBodySize = 1 << 20

// WriteJSON writes v as JSON. nil writes the status and nothing else.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "error", err)
	}
}

// DecodeJSON reads the body into v. Unknown fields are rejected. Over 1 MiB
// is a 413, not a parse error.
func DecodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()

	// LimitReader would truncate and the decoder would say "unexpected EOF"
	// on a perfectly fine body. The nil writer just skips the early close.
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, maxBodySize))
	dec.DisallowUnknownFields()

	var tooLarge *http.MaxBytesError

	switch err := dec.Decode(v); {
	case errors.Is(err, io.EOF):
		return Errorf(http.StatusBadRequest, "invalid_body", "request body is empty")
	case errors.As(err, &tooLarge):
		return Errorf(http.StatusRequestEntityTooLarge, "body_too_large",
			"request body must be at most %d bytes", maxBodySize)
	case err != nil:
		return Errorf(http.StatusBadRequest, "invalid_body", "%s", err)
	}

	if dec.More() {
		return Errorf(http.StatusBadRequest, "invalid_body", "request body must hold a single JSON value")
	}
	return nil
}
