// Package handlertest runs table-driven tests against HTTP handlers.
package handlertest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx"
)

// Case describes a single handler test case.
type Case struct {
	Name       string
	Method     string
	Path       string
	Body       string
	Headers    map[string]string
	NewHandler func(*gomock.Controller) httpx.Handler
	WantStatus int
	Check      func(t *testing.T, res *http.Response)
}

// Run executes every case against a router holding only the handler under test.
func Run(t *testing.T, cases []Case) {
	t.Helper()

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			router := httpx.NewRouter(c.NewHandler(ctrl))

			method := c.Method
			if method == "" {
				method = http.MethodGet
			}

			var body io.Reader
			if c.Body != "" {
				body = strings.NewReader(c.Body)
			}

			req := httptest.NewRequestWithContext(t.Context(), method, c.Path, body)
			for name, value := range c.Headers {
				req.Header.Set(name, value)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, c.WantStatus, res.StatusCode)
			if c.Check != nil {
				c.Check(t, res)
			}
		})
	}
}

// DecodeBody reads a JSON response body into v.
func DecodeBody(t *testing.T, res *http.Response, v any) {
	t.Helper()
	require.NoError(t, json.NewDecoder(res.Body).Decode(v))
}
