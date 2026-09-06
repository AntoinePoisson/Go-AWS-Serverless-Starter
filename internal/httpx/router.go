// Package httpx is the shared HTTP bits: router, JSON, errors, Lambda entry.
package httpx

import "net/http"

// Handler registers its own routes.
type Handler interface {
	AddRoutes(mux *http.ServeMux)
}

// Middleware wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// NewRouter wires handlers onto a mux and turns 404/405 into our JSON errors.
func NewRouter(handlers ...Handler) http.Handler {
	mux := http.NewServeMux()
	for _, h := range handlers {
		h.AddRoutes(mux)
	}
	return &router{mux: mux}
}

type router struct{ mux *http.ServeMux }

func (rt *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, pattern := rt.mux.Handler(r)
	if pattern != "" {
		// Through the mux so {id} and friends get set.
		rt.mux.ServeHTTP(w, r)
		return
	}

	// Empty pattern = mux 404/405. Keep status + Allow, drop the text body.
	probe := &routingErrorWriter{header: w.Header()}
	h.ServeHTTP(probe, r)

	switch probe.status {
	case http.StatusMethodNotAllowed:
		WriteError(w, r, Errorf(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed"))
	default:
		WriteError(w, r, Errorf(http.StatusNotFound, "not_found", "route not found"))
	}
}

type routingErrorWriter struct {
	header http.Header
	status int
}

func (w *routingErrorWriter) Header() http.Header { return w.header }

func (w *routingErrorWriter) WriteHeader(status int) { w.status = status }

func (w *routingErrorWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return len(body), nil
}

// Chain wraps h, first middleware outermost.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
