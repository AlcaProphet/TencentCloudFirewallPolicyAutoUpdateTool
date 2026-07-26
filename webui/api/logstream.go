package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
)

// LogBroadcaster 将 slog 日志广播到 SSE 订阅者
type LogBroadcaster struct {
	mu   sync.RWMutex
	subs map[int]chan string
	next int
}

// NewLogBroadcaster 创建日志广播器
func NewLogBroadcaster() *LogBroadcaster {
	return &LogBroadcaster{subs: make(map[int]chan string)}
}

// Subscribe 订阅日志流，返回 channel 和取消函数
func (b *LogBroadcaster) Subscribe() (<-chan string, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	id := b.nextID()
	ch := make(chan string, 64)
	b.subs[id] = ch
	return ch, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if c, ok := b.subs[id]; ok {
			close(c)
			delete(b.subs, id)
		}
	}
}

func (b *LogBroadcaster) nextID() int {
	id := b.next
	b.next++
	return id
}

// ─── slog.Handler 实现 ───

func (b *LogBroadcaster) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelDebug
}

func (b *LogBroadcaster) Handle(_ context.Context, r slog.Record) error {
	line := fmt.Sprintf("%s [%s] %s",
		r.Time.Format("15:04:05"), r.Level.String(), r.Message)
	// 附加属性
	r.Attrs(func(a slog.Attr) bool {
		line += fmt.Sprintf(" %s=%v", a.Key, a.Value.Any())
		return true
	})

	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs {
		select {
		case ch <- line:
		default: // 满则跳过
		}
	}
	return nil
}

func (b *LogBroadcaster) WithAttrs(attrs []slog.Attr) slog.Handler { return b }
func (b *LogBroadcaster) WithGroup(name string) slog.Handler     { return b }

// ─── multiHandler 多路输出 ───

// MultiHandler 将日志同时写入多个 Handler
type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: handlers}
}

// ─── SSE 端点 ───

func (d *Deps) handleLogStream(w http.ResponseWriter, r *http.Request) {
	if d.LogBroadcaster == nil {
		writeError(w, http.StatusBadRequest, "日志流不可用")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "SSE 不可用")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, unsubscribe := d.LogBroadcaster.Subscribe()
	defer unsubscribe()

	for {
		select {
		case line, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", line)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
