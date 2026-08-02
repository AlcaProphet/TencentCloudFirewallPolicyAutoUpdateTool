package config

import (
	"path/filepath"
	"testing"
	"time"
)

// TestClearSyncLogs 写入 3 条 → 清空 → GetSyncLogs 为空
func TestClearSyncLogs(t *testing.T) {
	store, err := OpenStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer store.Close()

	for i := 0; i < 3; i++ {
		if err := store.AddSyncLog(SyncLog{Timestamp: time.Now(), Target: "lhins-abc", Result: "success"}); err != nil {
			t.Fatalf("AddSyncLog 失败: %v", err)
		}
	}
	if logs, err := store.GetSyncLogs(10); err != nil || len(logs) != 3 {
		t.Fatalf("写入后 GetSyncLogs = %d 条, err = %v, want 3", len(logs), err)
	}

	if err := store.ClearSyncLogs(); err != nil {
		t.Fatalf("ClearSyncLogs 失败: %v", err)
	}
	if logs, err := store.GetSyncLogs(10); err != nil || len(logs) != 0 {
		t.Errorf("清空后 GetSyncLogs = %d 条, err = %v, want 0", len(logs), err)
	}
}
