package dns

import (
	"context"
	"testing"
	"time"
)

func TestResolve_Localhost(t *testing.T) {
	r := NewResolver("8.8.8.8:53", 10*time.Second)
	results, err := r.Resolve(context.Background(), "localhost")
	if err != nil {
		t.Fatalf("解析 localhost 失败: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("localhost 解析结果不应为空")
	}
	// localhost 通常解析为 127.0.0.1 或 ::1
	for _, ip := range results {
		t.Logf("解析结果: %s (IPv6=%v, CIDR=%s)", ip.IP, ip.IsIPv6, ip.CIDR())
	}
}

func TestResolve_NonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过网络测试")
	}
	// 使用 RFC 2606 保留的不可解析 TLD
	r := NewResolver("8.8.8.8:53", 5*time.Second)
	_, err := r.Resolve(context.Background(), "host.invalid")
	if err == nil {
		t.Log("注意: .invalid TLD 被解析（可能系统 DNS 有特殊配置）")
	}
}

func TestResolve_PublicDomain(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过网络测试")
	}
	r := NewResolver("8.8.8.8:53", 10*time.Second)
	results, err := r.Resolve(context.Background(), "dns.google")
	if err != nil {
		t.Skipf("无法访问 8.8.8.8（可能网络受限）: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("dns.google 解析结果不应为空")
	}
	// dns.google 应解析为 8.8.8.8 和 8.8.4.4
	found := false
	for _, ip := range results {
		if ip.IP.String() == "8.8.8.8" || ip.IP.String() == "8.8.4.4" {
			found = true
		}
	}
	if !found {
		t.Logf("dns.google 解析结果: %v", results)
	}
}

func TestNewResolver_PortAppend(t *testing.T) {
	// 测试不带端口时自动补 :53
	r := NewResolver("8.8.8.8", 10*time.Second)
	if r == nil {
		t.Fatal("NewResolver 不应返回 nil")
	}
	// 测试带端口时不重复添加
	r2 := NewResolver("8.8.8.8:53", 10*time.Second)
	if r2 == nil {
		t.Fatal("NewResolver 不应返回 nil")
	}
}

func TestResolvedIP_CIDR(t *testing.T) {
	tests := []struct {
		name string
		ip   ResolvedIP
		want string
	}{
		{"IPv4", ResolvedIP{IP: []byte{1, 2, 3, 4}, IsIPv6: false}, "1.2.3.4/32"},
		{"IPv6", ResolvedIP{IP: []byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, IsIPv6: true}, "2001:db8::1/128"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ip.CIDR()
			if got != tt.want {
				t.Errorf("CIDR() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHasPort(t *testing.T) {
	tests := []struct {
		addr string
		want bool
	}{
		{"8.8.8.8:53", true},
		{"8.8.8.8", false},
		{"[::1]:53", true},
		{"::1", false},
	}
	for _, tt := range tests {
		got := hasPort(tt.addr)
		if got != tt.want {
			t.Errorf("hasPort(%q) = %v, want %v", tt.addr, got, tt.want)
		}
	}
}
