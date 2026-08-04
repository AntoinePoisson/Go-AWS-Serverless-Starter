package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/antoinepoisson/bootstrap-go-aws/internal/httpx"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()

	httpx.WriteJSON(rec, http.StatusCreated, map[string]string{"id": "42"})

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"id":"42"}`, rec.Body.String())
}

func TestWriteJSONWithoutBody(t *testing.T) {
	rec := httptest.NewRecorder()

	httpx.WriteJSON(rec, http.StatusNoContent, nil)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
}

func TestDecodeJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	cases := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "valid", body: `{"name":"demo"}`},
		{name: "empty body", body: ``, wantErr: true},
		{name: "malformed", body: `{"name":`, wantErr: true},
		{name: "unknown field", body: `{"name":"demo","extra":1}`, wantErr: true},
		{name: "trailing value", body: `{"name":"demo"}{"name":"other"}`, wantErr: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(c.body))

			var got payload
			err := httpx.DecodeJSON(req, &got)

			if c.wantErr {
				var httpErr *httpx.Error
				require.ErrorAs(t, err, &httpErr)
				assert.Equal(t, http.StatusBadRequest, httpErr.Status)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "demo", got.Name)
		})
	}
}
