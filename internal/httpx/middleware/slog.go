package middleware

import (
	"context"
	"log/slog"
)

// LogRequestID wraps h so every record carries the request identifier held by
// its context, not only the lines the Logger middleware writes.
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

// Both have to rewrap: the embedded handler returns a bare one, which would
// silently drop the identifier from that point on.
func (h *requestIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &requestIDHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *requestIDHandler) WithGroup(name string) slog.Handler {
	return &requestIDHandler{Handler: h.Handler.WithGroup(name)}
}
