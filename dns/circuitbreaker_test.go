package dns

import "testing"

func TestCircuitBreaker_Trigger(t *testing.T) {
	cb := NewCircuitBreaker(3)

	// 未达阈值，不熔断
	cb.RecordFailure("example.com")
	cb.RecordFailure("example.com")
	if cb.IsOpen("example.com") {
		t.Error("未达阈值不应熔断")
	}

	// 达到阈值，触发熔断
	cb.RecordFailure("example.com")
	if !cb.IsOpen("example.com") {
		t.Error("达到阈值应熔断")
	}

	// 其他域名不受影响
	if cb.IsOpen("other.com") {
		t.Error("其他域名不应熔断")
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(2)

	cb.RecordFailure("example.com")
	cb.RecordFailure("example.com")
	if !cb.IsOpen("example.com") {
		t.Error("应已熔断")
	}

	// 成功后解除
	cb.RecordSuccess("example.com")
	if cb.IsOpen("example.com") {
		t.Error("成功后应解除熔断")
	}
}
