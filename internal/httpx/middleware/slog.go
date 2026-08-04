package middleware

import (
	"context"
	"log/slog"
)

// LogRequestID wraps h so every record carries the request identifier held by
// its context. Without it only the Logger middleware knows the identifier, and
// an error logged inside a handler cannot be traced back to the request that
// caused it.
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

// WithAttrs and WithGroup have to rewrap: the embedded handler returns a bare
// handler, which would silently drop the identifier from that point on.
func (h *requestIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &requestIDHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *requestIDHandler) WithGroup(name string) slog.Handler {
	return &requestIDHandler{Handler: h.Handler.WithGroup(name)}
}
