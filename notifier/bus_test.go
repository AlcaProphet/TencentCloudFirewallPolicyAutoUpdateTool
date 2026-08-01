package notifier

import (
	"sync"
	"testing"
	"time"
)

type mockSubscriber struct {
	mu     sync.Mutex
	events []Event
}

func (m *mockSubscriber) OnEvent(event Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func TestEventBus_Publish(t *testing.T) {
	bus := NewEventBus()
	sub := &mockSubscriber{}

	bus.Subscribe(EventSyncComplete, sub)

	bus.Publish(Event{
		Type:      EventSyncComplete,
		Timestamp: time.Now(),
		Data:      map[string]any{"domain": "example.com"},
	})

	// 等待异步投递完成
	time.Sleep(100 * time.Millisecond)

	sub.mu.Lock()
	defer sub.mu.Unlock()
	if len(sub.events) != 1 {
		t.Fatalf("收到事件数 = %d, want 1", len(sub.events))
	}
	if sub.events[0].Type != EventSyncComplete {
		t.Errorf("事件类型 = %s, want sync:complete", sub.events[0].Type)
	}
}

func TestEventBus_NoCrossTalk(t *testing.T) {
	bus := NewEventBus()
	sub := &mockSubscriber{}

	// 只订阅 sync:error
	bus.Subscribe(EventSyncError, sub)

	// 发布 sync:complete（不应收到）
	bus.Publish(Event{Type: EventSyncComplete, Timestamp: time.Now()})

	time.Sleep(100 * time.Millisecond)

	sub.mu.Lock()
	defer sub.mu.Unlock()
	if len(sub.events) != 0 {
		t.Errorf("不应收到非订阅类型的事件, got %d", len(sub.events))
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus()
	sub := &mockSubscriber{}

	bus.Subscribe(EventSyncError, sub)
	bus.Unsubscribe(EventSyncError, sub)

	// 发布不应收到的事件
	bus.Publish(Event{Type: EventSyncError, Timestamp: time.Now()})
	time.Sleep(100 * time.Millisecond)

	sub.mu.Lock()
	defer sub.mu.Unlock()
	if len(sub.events) != 0 {
		t.Errorf("取消订阅后不应收到事件, got %d", len(sub.events))
	}
}

func TestEventBus_UnsubscribeIdempotent(t *testing.T) {
	bus := NewEventBus()
	sub := &mockSubscriber{}

	// 对未订阅的类型取消订阅 — 不应 panic
	bus.Unsubscribe(EventSyncComplete, sub)

	// 重复取消 — 不应 panic
	bus.Subscribe(EventSyncError, sub)
	bus.Unsubscribe(EventSyncError, sub)
	bus.Unsubscribe(EventSyncError, sub) // 幂等
}

func TestEventBus_UnsubscribeOnlyTargetType(t *testing.T) {
	bus := NewEventBus()
	sub := &mockSubscriber{}

	bus.Subscribe(EventSyncError, sub)
	bus.Subscribe(EventDNSFailed, sub)

	// 仅取消 sync:error 的订阅
	bus.Unsubscribe(EventSyncError, sub)

	// 发布 dns:failed — 应收到
	bus.Publish(Event{Type: EventDNSFailed, Timestamp: time.Now()})
	time.Sleep(100 * time.Millisecond)

	sub.mu.Lock()
	defer sub.mu.Unlock()
	if len(sub.events) != 1 {
		t.Errorf("dns:failed 应收到事件, got %d", len(sub.events))
	}
}
