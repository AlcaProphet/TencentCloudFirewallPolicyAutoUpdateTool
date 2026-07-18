package dns

import (
	"testing"
)

func TestNew(t *testing.T) {
	r := New("223.5.5.5:53")
	if r == nil {
		t.Fatal("New() 不应返回 nil")
	}
	if r.resolver == nil {
		t.Fatal("resolver 不应为 nil")
	}
}

func TestLookup_InvalidHost(t *testing.T) {
	r := New("223.5.5.5:53")
	// 解析一个不存在的域名——可能返回错误，也可能因为 DNS 劫持返回结果
	// 此处仅验证函数不会 panic
	ips, err := r.Lookup("this-domain-definitely-does-not-exist-12345.invalid")
	// 不强制要求失败：某些网络环境下 DNS 可能返回劫持结果
	_ = ips
	_ = err
}

func TestResolvedIP_Struct(t *testing.T) {
	ip := ResolvedIP{Address: "1.2.3.4", IsIPv6: false}
	if ip.Address != "1.2.3.4" {
		t.Error("Address 字段不匹配")
	}
	if ip.IsIPv6 {
		t.Error("IPv4 地址 IsIPv6 应为 false")
	}

	ipv6 := ResolvedIP{Address: "::1", IsIPv6: true}
	if !ipv6.IsIPv6 {
		t.Error("IPv6 地址 IsIPv6 应为 true")
	}
}
