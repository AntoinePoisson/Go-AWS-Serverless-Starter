package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/AntoinePoisson/go-aws-serverless-starter/internal/httpx"
)

type stubHandler struct{}

func (stubHandler) AddRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"pong": "true"})
	})
}

func TestNewRouterRegistersHandlerRoutes(t *testing.T) {
	router := httpx.NewRouter(stubHandler{})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestNewRouterRejectsUnknownMethod(t *testing.T) {
	router := httpx.NewRouter(stubHandler{})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/ping", nil))

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestChainAppliesMiddlewaresOutsideIn(t *testing.T) {
	var order []string

	tag := func(name string) httpx.Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	handler := httpx.Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) { order = append(order, "handler") }),
		tag("first"),
		tag("second"),
	)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, []string{"first", "second", "handler"}, order)
}
