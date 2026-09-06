package middleware

import "net/http"

// responseWriter remembers the status and whether anything was written.
type responseWriter struct {
	http.ResponseWriter
	status  int
	written bool
}

// wrapResponseWriter reuses the wrapper an outer middleware already put on.
// Logger and Recover both need one and always sit together, so without this
// we'd stack two and the inner one would hide the outer.
func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	if wrapped, ok := w.(*responseWriter); ok {
		return wrapped
	}
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (w *responseWriter) WriteHeader(status int) {
	if !w.written {
		w.status = status
		w.written = true
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.written = true
	return w.ResponseWriter.Write(b)
}

// Unwrap is what ResponseController walks. Without it, flush/hijack say
// "feature not supported" behind any middleware.
func (w *responseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
