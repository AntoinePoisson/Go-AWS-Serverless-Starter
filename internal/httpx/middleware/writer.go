package middleware

import "net/http"

// responseWriter records the status a handler sent, and whether it started
// writing at all.
type responseWriter struct {
	http.ResponseWriter
	status  int
	written bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
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
