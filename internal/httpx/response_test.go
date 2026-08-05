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
		name       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{name: "valid", body: `{"name":"demo"}`},
		{name: "empty body", body: ``, wantStatus: http.StatusBadRequest, wantCode: "invalid_body"},
		{name: "malformed", body: `{"name":`, wantStatus: http.StatusBadRequest, wantCode: "invalid_body"},
		{name: "unknown field", body: `{"name":"demo","extra":1}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_body"},
		{name: "trailing value", body: `{"name":"demo"}{"name":"other"}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_body"},
		{
			name:       "body over the limit",
			body:       `{"name":"` + strings.Repeat("a", 1<<20) + `"}`,
			wantStatus: http.StatusRequestEntityTooLarge,
			wantCode:   "body_too_large",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(c.body))

			var got payload
			err := httpx.DecodeJSON(req, &got)

			if c.wantStatus != 0 {
				var httpErr *httpx.Error
				require.ErrorAs(t, err, &httpErr)
				assert.Equal(t, c.wantStatus, httpErr.Status)
				assert.Equal(t, c.wantCode, httpErr.Code)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "demo", got.Name)
		})
	}
}
