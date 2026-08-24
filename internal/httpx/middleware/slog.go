package middleware

import (
	"context"
	"log/slog"
)

// LogRequestID puts the request id on every record from this context,
// not just the line Logger writes.
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

// Rewrap both. The embedded handler hands back a bare one and we'd lose the id.
func (h *requestIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &requestIDHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *requestIDHandler) WithGroup(name string) slog.Handler {
	return &requestIDHandler{Handler: h.Handler.WithGroup(name)}
}
