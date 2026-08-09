package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx/middleware"
)

// A middleware must not cut a handler off from the writer underneath.
func TestResponseControllerReachesTheWrappedWriter(t *testing.T) {
	wrappers := map[string]func(http.Handler) http.Handler{
		"Logger":  middleware.Logger,
		"Recover": middleware.Recover,
	}

	for name, wrap := range wrappers {
		t.Run(name, func(t *testing.T) {
			var flushErr error
			handler := wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				flushErr = http.NewResponseController(w).Flush()
			}))

			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/items", nil))

			assert.NoError(t, flushErr)
		})
	}
}
