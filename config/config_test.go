package config

import (
	"os"
	"testing"
	"time"
)

func TestParseDomainRules(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    int
		wantErr bool
	}{
		{"单条TCP规则", "api.example.com|TCP|443,80|ACCEPT", 1, false},
		{"单条UDP规则", "dns.example.com|UDP|53|ACCEPT", 1, false},
		{"TCP+UDP规则", "cdn.example.com|TCP+UDP|443|ACCEPT", 1, false},
		{"多条规则", "api.example.com|TCP|443|ACCEPT;cdn.example.com|UDP|53|ACCEPT", 2, false},
		{"全端口", "api.example.com|TCP|*|ACCEPT", 1, false},
		{"DROP动作", "bad.example.com|TCP|80|DROP", 1, false},
		{"空字符串", "", 0, true},
		{"缺少字段", "api.example.com|TCP|443", 0, true},
		{"无效协议", "api.example.com|ICMP|443|ACCEPT", 0, true},
		{"无效动作", "api.example.com|TCP|443|ALLOW", 0, true},
		{"无效端口", "api.example.com|TCP|99999|ACCEPT", 0, true},
		{"负数端口", "api.example.com|TCP|-1|ACCEPT", 0, true},
		{"端口为0", "api.example.com|TCP|0|ACCEPT", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := parseDomainRules(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDomainRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(rules) != tt.want {
				t.Errorf("parseDomainRules() got %d rules, want %d", len(rules), tt.want)
			}
		})
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	// 确保环境变量为空
	os.Unsetenv("TENCENTCLOUD_SECRET_ID")
	os.Unsetenv("TENCENTCLOUD_SECRET_KEY")
	os.Unsetenv("LIGHTHOUSE_INSTANCE_ID")
	os.Unsetenv("LIGHTHOUSE_REGION")
	os.Unsetenv("DOMAIN_RULES")

	_, err := Load()
	if err == nil {
		t.Error("期望缺少必填项时报错")
	}
}

func TestLoad_InvalidInstanceID(t *testing.T) {
	os.Setenv("TENCENTCLOUD_SECRET_ID", "test-id")
	os.Setenv("TENCENTCLOUD_SECRET_KEY", "test-key")
	os.Setenv("LIGHTHOUSE_INSTANCE_ID", "invalid-instance")
	os.Setenv("LIGHTHOUSE_REGION", "ap-guangzhou")
	os.Setenv("DOMAIN_RULES", "api.example.com|TCP|443|ACCEPT")

	_, err := Load()
	if err == nil {
		t.Error("期望无效 InstanceID 格式时报错")
	}
}

func TestLoad_InvalidDNSServer(t *testing.T) {
	os.Setenv("TENCENTCLOUD_SECRET_ID", "test-id")
	os.Setenv("TENCENTCLOUD_SECRET_KEY", "test-key")
	os.Setenv("LIGHTHOUSE_INSTANCE_ID", "lhins-test1234")
	os.Setenv("LIGHTHOUSE_REGION", "ap-guangzhou")
	os.Setenv("DOMAIN_RULES", "api.example.com|TCP|443|ACCEPT")
	os.Setenv("DNS_SERVER", "invalid-format")

	_, err := Load()
	if err == nil {
		t.Error("期望无效 DNS_SERVER 格式时报错")
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	os.Setenv("TENCENTCLOUD_SECRET_ID", "test-id")
	os.Setenv("TENCENTCLOUD_SECRET_KEY", "test-key")
	os.Setenv("LIGHTHOUSE_INSTANCE_ID", "lhins-test1234")
	os.Setenv("LIGHTHOUSE_REGION", "ap-guangzhou")
	os.Setenv("DOMAIN_RULES", "api.example.com|TCP|443,80|ACCEPT;cdn.example.com|TCP+UDP|443|ACCEPT")
	os.Setenv("CHECK_INTERVAL", "30s")
	os.Setenv("DNS_SERVER", "1.1.1.1:53")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() 不应该报错: %v", err)
	}

	if len(cfg.DomainRules) != 2 {
		t.Errorf("期望 2 条规则，实际 %d", len(cfg.DomainRules))
	}
	if cfg.Interval != 30*time.Second {
		t.Errorf("期望间隔 30s，实际 %v", cfg.Interval)
	}
	if cfg.DNSServer != "1.1.1.1:53" {
		t.Errorf("期望 DNS 1.1.1.1:53，实际 %s", cfg.DNSServer)
	}
}
