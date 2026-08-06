// Package httpx contains the HTTP plumbing shared by the functions: routing,
// JSON responses, error rendering and the Lambda entry point.
package httpx

import "net/http"

// Handler registers its own routes on a mux.
type Handler interface {
	AddRoutes(mux *http.ServeMux)
}

// Middleware wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// NewRouter builds a mux from the given handlers.
func NewRouter(handlers ...Handler) *http.ServeMux {
	mux := http.NewServeMux()
	for _, h := range handlers {
		h.AddRoutes(mux)
	}
	return mux
}

// Chain wraps h with the given middlewares. The first one is the outermost.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
