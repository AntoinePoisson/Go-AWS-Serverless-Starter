package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
)

const maxBodySize = 1 << 20

// WriteJSON writes v as a JSON body with the given status code. A nil value
// writes the status code only.
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

// DecodeJSON reads the request body into v. Unknown fields are rejected, and a
// body over 1 MiB is a 413 rather than a parse error.
func DecodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()

	// Not io.LimitReader: truncating reaches the decoder as "unexpected EOF"
	// and reports a body the client sent correctly as malformed. The nil writer
	// only costs the early connection close.
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
