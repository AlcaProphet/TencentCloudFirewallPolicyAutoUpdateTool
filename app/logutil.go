package app

import (
	"context"
	"log/slog"
)

// MultiHandler 将日志同时写入多个 Handler
type MultiHandler struct {
	Handlers []slog.Handler
}

// NewMultiHandler 创建多路 Handler
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{Handlers: handlers}
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.Handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.Handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.Handlers))
	for i, h := range m.Handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{Handlers: handlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.Handlers))
	for i, h := range m.Handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{Handlers: handlers}
}
