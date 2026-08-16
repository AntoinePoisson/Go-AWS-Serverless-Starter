package middleware

import "net/http"

// responseWriter records the status a handler sent, and whether it started
// writing at all.
type responseWriter struct {
	http.ResponseWriter
	status  int
	written bool
}

// wrapResponseWriter reuses the wrapper an outer middleware already installed.
// Logger and Recover both need one and are always chained together, so without
// this every request carries two of them, the inner one shadowing the outer.
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

// Unwrap is what http.ResponseController follows. Without it, flushing and
// hijacking report "feature not supported" behind any middleware.
func (w *responseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
