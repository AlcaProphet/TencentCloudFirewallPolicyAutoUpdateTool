package notifier

import (
	"log/slog"
	"sync"
	"time"
)

// EventType 事件类型
type EventType string

const (
	EventSyncStart    EventType = "sync:start"
	EventSyncComplete EventType = "sync:complete"
	EventSyncError    EventType = "sync:error"
	EventRuleChanged  EventType = "rule:changed"
	EventDNSFailed    EventType = "dns:failed"
)

// Event 事件
type Event struct {
	Type      EventType
	Timestamp time.Time
	Data      map[string]any
}

// Subscriber 事件订阅者接口
type Subscriber interface {
	OnEvent(event Event) error
}

// EventBus 事件总线
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[EventType][]Subscriber
}

// NewEventBus 创建事件总线
func NewEventBus() *EventBus {
	return &EventBus{subscribers: make(map[EventType][]Subscriber)}
}

// Subscribe 订阅事件
func (b *EventBus) Subscribe(eventType EventType, sub Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], sub)
}

// Publish 异步投递，不阻塞调用方
func (b *EventBus) Publish(event Event) {
	b.mu.RLock()
	subs := b.subscribers[event.Type]
	b.mu.RUnlock()

	for _, sub := range subs {
		go func(s Subscriber) {
			if err := s.OnEvent(event); err != nil {
				slog.Warn("事件处理失败", "type", event.Type, "error", err)
			}
		}(sub)
	}
}
