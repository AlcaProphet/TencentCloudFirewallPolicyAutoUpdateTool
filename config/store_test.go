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

// TestScannedResources 覆盖式替换 + 跨地域汇总 + 清理
func TestScannedResources(t *testing.T) {
	store, err := OpenStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer store.Close()

	// 首次扫描：tc_cvm ap-guangzhou 两条
	if err := store.ReplaceScannedResources("tc_cvm", "ap-guangzhou", []ScannedResource{
		{ResourceID: "sg-aaa", ResourceName: "安全组A"},
		{ResourceID: "sg-bbb", ResourceName: "安全组B"},
	}); err != nil {
		t.Fatalf("ReplaceScannedResources 失败: %v", err)
	}
	// 再次扫描同一地域：覆盖（仅保留新的一条）
	if err := store.ReplaceScannedResources("tc_cvm", "ap-guangzhou", []ScannedResource{
		{ResourceID: "sg-ccc", ResourceName: "安全组C"},
	}); err != nil {
		t.Fatalf("二次 ReplaceScannedResources 失败: %v", err)
	}
	// 另一地域扫描：跨地域汇总
	if err := store.ReplaceScannedResources("tc_cvm", "ap-beijing", []ScannedResource{
		{ResourceID: "sg-ddd", ResourceName: "安全组D"},
	}); err != nil {
		t.Fatalf("跨地域 ReplaceScannedResources 失败: %v", err)
	}

	resources, err := store.GetScannedResources("tc_cvm")
	if err != nil {
		t.Fatalf("GetScannedResources 失败: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("GetScannedResources = %d 条, want 2（覆盖后跨地域汇总）", len(resources))
	}
	seen := map[string]bool{}
	for _, r := range resources {
		seen[r.ResourceID] = true
		if r.CloudType != "tc_cvm" {
			t.Errorf("资源 CloudType 错误: %s", r.CloudType)
		}
	}
	if !seen["sg-ccc"] || !seen["sg-ddd"] {
		t.Errorf("期望资源 sg-ccc/sg-ddd，实际 %v", seen)
	}

	// 清理：仅清理 tc_cvm
	if err := store.DeleteScannedResources("tc_cvm"); err != nil {
		t.Fatalf("DeleteScannedResources 失败: %v", err)
	}
	if resources, err := store.GetScannedResources("tc_cvm"); err != nil || len(resources) != 0 {
		t.Errorf("清理后 GetScannedResources = %d 条, err = %v, want 0", len(resources), err)
	}
}

// TestResetAll 写入多类数据后 ResetAll 清空全部表
func TestResetAll(t *testing.T) {
	store, err := OpenStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer store.Close()

	// 写入各表数据
	if err := store.AddTarget(TargetConfig{CloudType: CloudTCLighthouse, Region: "ap-guangzhou", ResourceID: "lhins-abc"}); err != nil {
		t.Fatalf("AddTarget 失败: %v", err)
	}
	if err := store.AddRule(DomainRule{Host: "example.com", Protocol: "TCP", Ports: "80"}); err != nil {
		t.Fatalf("AddRule 失败: %v", err)
	}
	if err := store.SetSetting("tc_access_id", "AKIDxxx"); err != nil {
		t.Fatalf("SetSetting 失败: %v", err)
	}
	if err := store.AddSyncLog(SyncLog{Timestamp: time.Now(), Target: "lhins-abc", Result: "success"}); err != nil {
		t.Fatalf("AddSyncLog 失败: %v", err)
	}
	if err := store.SaveAlertEmail(&AlertEmailConfig{Enabled: true, Host: "smtp.example.com"}); err != nil {
		t.Fatalf("SaveAlertEmail 失败: %v", err)
	}
	if err := store.SaveAlertWebhook(&AlertWebhookConfig{Enabled: true, URL: "https://example.com/hook"}); err != nil {
		t.Fatalf("SaveAlertWebhook 失败: %v", err)
	}
	if err := store.ReplaceScannedResources("tc_lighthouse", "ap-guangzhou", []ScannedResource{{ResourceID: "lhins-xyz"}}); err != nil {
		t.Fatalf("ReplaceScannedResources 失败: %v", err)
	}

	// ResetAll 清空全部
	if err := store.ResetAll(); err != nil {
		t.Fatalf("ResetAll 失败: %v", err)
	}

	if targets, _ := store.GetTargets(); len(targets) != 0 {
		t.Errorf("ResetAll 后 targets 仍有 %d 条", len(targets))
	}
	if rules, _ := store.GetRules(); len(rules) != 0 {
		t.Errorf("ResetAll 后 rules 仍有 %d 条", len(rules))
	}
	if settings, _ := store.GetSettings(); len(settings) != 0 {
		t.Errorf("ResetAll 后 settings 仍有 %d 条", len(settings))
	}
	if logs, _ := store.GetSyncLogs(10); len(logs) != 0 {
		t.Errorf("ResetAll 后 sync_logs 仍有 %d 条", len(logs))
	}
	if email, _ := store.GetAlertEmail(); email.Enabled || email.Host != "" {
		t.Errorf("ResetAll 后 alert_email 仍残留: %+v", email)
	}
	if webhook, _ := store.GetAlertWebhook(); webhook.Enabled || webhook.URL != "" {
		t.Errorf("ResetAll 后 alert_webhook 仍残留: %+v", webhook)
	}
	if scanned, _ := store.GetScannedResources("tc_lighthouse"); len(scanned) != 0 {
		t.Errorf("ResetAll 后 scanned_resources 仍有 %d 条", len(scanned))
	}
}
