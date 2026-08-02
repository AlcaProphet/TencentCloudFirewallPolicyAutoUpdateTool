package api

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// newRecord 构造测试用日志记录
func newRecord(level slog.Level, msg string) slog.Record {
	r := slog.NewRecord(time.Now(), level, msg, 0)
	return r
}

// TestLogBroadcaster_Replay 写入 3 条 → 订阅回放 3 条且为正序（首条为最早写入）
func TestLogBroadcaster_Replay(t *testing.T) {
	b := NewLogBroadcaster("info")
	for _, msg := range []string{"第一条", "第二条", "第三条"} {
		if err := b.Handle(context.Background(), newRecord(slog.LevelInfo, msg)); err != nil {
			t.Fatalf("Handle 失败: %v", err)
		}
	}

	ch, unsub := b.Subscribe()
	defer unsub()

	for i, want := range []string{"第一条", "第二条", "第三条"} {
		select {
		case line := <-ch:
			// TextHandler 对中文消息会加引号（输出 msg="第一条"），直接检查消息文本即可
			if !strings.Contains(line, want) {
				t.Errorf("回放第 %d 条 = %q, want 包含 %s", i, line, want)
			}
		case <-time.After(time.Second):
			t.Fatalf("回放第 %d 条超时", i)
		}
	}
	// 不应有多余输出
	select {
	case line, ok := <-ch:
		if ok {
			t.Errorf("回放多出内容: %q", line)
		}
	case <-time.After(100 * time.Millisecond):
	}
}

// TestLogBroadcaster_RingOverflow 写入超过容量 → 订阅恰好回放最近 logRingSize 条（最旧的被淘汰）
func TestLogBroadcaster_RingOverflow(t *testing.T) {
	b := NewLogBroadcaster("info")
	for i := 0; i < logRingSize+5; i++ {
		if err := b.Handle(context.Background(), newRecord(slog.LevelInfo, "消息")); err != nil {
			t.Fatalf("Handle 失败: %v", err)
		}
	}

	ch, unsub := b.Subscribe()
	defer unsub()

	for i := 0; i < logRingSize; i++ {
		select {
		case <-ch:
		case <-time.After(3 * time.Second):
			t.Fatalf("回放第 %d 条超时", i)
		}
	}
	select {
	case line, ok := <-ch:
		if ok {
			t.Errorf("回放应恰好 %d 条，多出: %q", logRingSize, line)
		}
	case <-time.After(100 * time.Millisecond):
	}
}

// TestLogBroadcaster_Format 行格式与 slog.TextHandler 一致（含 level= 与 msg=）
func TestLogBroadcaster_Format(t *testing.T) {
	b := NewLogBroadcaster("info")
	if err := b.Handle(context.Background(), newRecord(slog.LevelInfo, "同步完成")); err != nil {
		t.Fatalf("Handle 失败: %v", err)
	}

	ch, unsub := b.Subscribe()
	defer unsub()

	select {
	case line := <-ch:
		// 注意：TextHandler 对中文消息加引号（输出 msg="同步完成"），此处只断言级别与消息文本
		if !strings.Contains(line, "level=INFO") || !strings.Contains(line, "同步完成") {
			t.Errorf("行格式不符合 TextHandler 规范: %q", line)
		}
	case <-time.After(time.Second):
		t.Fatal("回放超时")
	}
}

// TestLogBroadcaster_LevelFilter 级别过滤：info 级别下 debug 日志不进入缓冲
// 注意：必须经 slog.Logger 写入（先检查 Enabled），直接调用 Handle 会绕过过滤导致断言失败
func TestLogBroadcaster_LevelFilter(t *testing.T) {
	b := NewLogBroadcaster("info")
	if b.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("info 级别下 debug 应被过滤")
	}
	logger := slog.New(b)
	logger.Debug("调试") // slog.Logger 先检查 Enabled → false → 不调用 Handle，ring 保持为空

	ch, unsub := b.Subscribe()
	defer unsub()
	select {
	case line, ok := <-ch:
		if ok {
			t.Errorf("debug 日志不应进入缓冲: %q", line)
		}
	case <-time.After(100 * time.Millisecond):
	}
}
