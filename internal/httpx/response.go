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

// DecodeJSON reads the request body into v. Unknown fields and bodies larger
// than 1 MiB are rejected.
func DecodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()

	dec := json.NewDecoder(io.LimitReader(r.Body, maxBodySize))
	dec.DisallowUnknownFields()

	switch err := dec.Decode(v); {
	case errors.Is(err, io.EOF):
		return Errorf(http.StatusBadRequest, "invalid_body", "request body is empty")
	case err != nil:
		return Errorf(http.StatusBadRequest, "invalid_body", "%s", err)
	}

	if dec.More() {
		return Errorf(http.StatusBadRequest, "invalid_body", "request body must hold a single JSON value")
	}
	return nil
}
