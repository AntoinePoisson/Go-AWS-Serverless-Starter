package middleware

import (
	"context"
	"log/slog"
)

// LogRequestID wraps h so every record carries the request id from its
// context, not just the lines Logger writes.
func LogRequestID(h slog.Handler) slog.Handler {
	return &requestIDHandler{Handler: h}
}

type requestIDHandler struct{ slog.Handler }

func (h *requestIDHandler) Handle(ctx context.Context, record slog.Record) error {
	if id := RequestIDFrom(ctx); id != "" {
		record.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, record)
}

// Both have to rewrap. The embedded handler returns a bare one and would drop
// the id from there on.
func (h *requestIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &requestIDHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *requestIDHandler) WithGroup(name string) slog.Handler {
	return &requestIDHandler{Handler: h.Handler.WithGroup(name)}
}
