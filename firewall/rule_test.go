package firewall

import (
	"strings"
	"testing"

	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"

	"github.com/alcaprophet/fwalizer/config"
	"github.com/alcaprophet/fwalizer/dns"
)

func TestOwnedRules(t *testing.T) {
	desc1 := "[auto-dns] TCP 443"
	desc2 := "[auto-dns] TCP 80"
	desc3 := "[manual] some rule"

	allRules := []*lighthouse.FirewallRuleInfo{
		{FirewallRuleDescription: &desc1},
		{FirewallRuleDescription: &desc2},
		{FirewallRuleDescription: &desc3},
		{FirewallRuleDescription: nil},
	}

	owned := ownedRules(allRules, "auto-dns")
	if len(owned) != 2 {
		t.Errorf("期望 2 条规则，实际 %d", len(owned))
	}
}

func TestDiff_AddOnly(t *testing.T) {
	resolved := []dns.ResolvedIP{
		{Address: "1.2.3.4", IsIPv6: false},
	}
	rule := config.DomainRule{
		Host:     "api.example.com",
		Protocol: "TCP",
		Ports:    "443",
		Action:   "ACCEPT",
	}
	desc := "[auto-dns]"

	toAdd, toDelete := Diff(resolved, rule, desc, nil)

	if len(toDelete) != 0 {
		t.Errorf("期望无删除，实际 %d", len(toDelete))
	}
	if len(toAdd) != 1 {
		t.Errorf("期望添加 1 条，实际 %d", len(toAdd))
	}
}

func TestDiff_DeleteOnly(t *testing.T) {
	resolved := []dns.ResolvedIP{} // DNS 解析返回空

	rule := config.DomainRule{
		Host:     "api.example.com",
		Protocol: "TCP",
		Ports:    "443",
		Action:   "ACCEPT",
	}
	desc := "[auto-dns]"

	proto := "TCP"
	port := "443"
	action := "ACCEPT"
	cidr := "1.2.3.4/32"
	descInfo := "[auto-dns] TCP 443"

	existing := []*lighthouse.FirewallRuleInfo{
		{
			Protocol:                &proto,
			Port:                    &port,
			Action:                  &action,
			CidrBlock:               &cidr,
			FirewallRuleDescription: &descInfo,
		},
	}

	toAdd, toDelete := Diff(resolved, rule, desc, existing)

	if len(toAdd) != 0 {
		t.Errorf("期望无添加，实际 %d", len(toAdd))
	}
	if len(toDelete) != 1 {
		t.Errorf("期望删除 1 条，实际 %d", len(toDelete))
	}
}

func TestDiff_NoChange(t *testing.T) {
	resolved := []dns.ResolvedIP{
		{Address: "1.2.3.4", IsIPv6: false},
	}

	rule := config.DomainRule{
		Host:     "api.example.com",
		Protocol: "TCP",
		Ports:    "443",
		Action:   "ACCEPT",
	}
	desc := "[auto-dns]"

	proto := "TCP"
	port := "443"
	action := "ACCEPT"
	cidr := "1.2.3.4/32"
	descInfo := "[auto-dns] TCP 443"

	existing := []*lighthouse.FirewallRuleInfo{
		{
			Protocol:                &proto,
			Port:                    &port,
			Action:                  &action,
			CidrBlock:               &cidr,
			FirewallRuleDescription: &descInfo,
		},
	}

	toAdd, toDelete := Diff(resolved, rule, desc, existing)

	if len(toAdd) != 0 || len(toDelete) != 0 {
		t.Errorf("期望无变更，实际 add=%d delete=%d", len(toAdd), len(toDelete))
	}
}

func TestMakeRule(t *testing.T) {
	ip := dns.ResolvedIP{Address: "1.2.3.4", IsIPv6: false}
	r := makeRule(ip, "TCP", "443", "ACCEPT", "[auto-dns:test] TCP 443")

	if *r.Protocol != "TCP" {
		t.Errorf("期望 TCP，实际 %s", *r.Protocol)
	}
	if *r.CidrBlock != "1.2.3.4/32" {
		t.Errorf("期望 1.2.3.4/32，实际 %s", *r.CidrBlock)
	}
	if r.Ipv6CidrBlock != nil {
		t.Error("IPv4 规则不应有 Ipv6CidrBlock")
	}
}

func TestMakeRule_IPv6(t *testing.T) {
	ip := dns.ResolvedIP{Address: "2001:db8::1", IsIPv6: true}
	r := makeRule(ip, "TCP", "443", "ACCEPT", "[auto-dns:test] TCP 443")

	if *r.Ipv6CidrBlock != "2001:db8::1/128" {
		t.Errorf("期望 2001:db8::1/128，实际 %s", *r.Ipv6CidrBlock)
	}
	if r.CidrBlock != nil {
		t.Error("IPv6 规则不应有 CidrBlock")
	}
}

func TestSafeStr(t *testing.T) {
	if safeStr(nil) != "" {
		t.Error("nil 应返回空字符串")
	}

	val := "test"
	if safeStr(&val) != "test" {
		t.Error("非 nil 应返回值")
	}
}

func TestBuildExpectedRules_TCPPlusUDP(t *testing.T) {
	resolved := []dns.ResolvedIP{
		{Address: "1.2.3.4", IsIPv6: false},
	}
	rule := config.DomainRule{
		Protocol: "TCP+UDP",
		Ports:    "443",
		Action:   "ACCEPT",
	}
	desc := "[auto-dns]"

	rules := buildExpectedRules(resolved, rule, desc)
	if len(rules) != 2 {
		t.Fatalf("TCP+UDP 应生成 2 条规则，实际 %d", len(rules))
	}
	if *rules[0].Protocol != "TCP" {
		t.Errorf("第一条应为 TCP，实际 %s", *rules[0].Protocol)
	}
	if *rules[1].Protocol != "UDP" {
		t.Errorf("第二条应为 UDP，实际 %s", *rules[1].Protocol)
	}
}

func TestTruncateDescription(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		detail string
		max    int
		want   string
	}{
		{"不截断", "[auto-dns]", "生产API", 20, "[auto-dns]生产API"},
		{"刚好等于20", "[auto-dns]", "1234567890", 20, "[auto-dns]1234567890"},
		{"超长截断", "[auto-dns]", "very long comment for testing truncation", 28,
			"[auto-dns]very...(truncated)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateDescription(tt.prefix, tt.detail, tt.max)
			if len(got) > tt.max {
				t.Errorf("截断结果长度 %d 超过限制 %d: %s", len(got), tt.max, got)
			}
			// 验证前缀未被截断
			if !strings.HasPrefix(got, tt.prefix) {
				t.Errorf("前缀被截断: %q", got)
			}
			if got != tt.want {
				t.Errorf("truncateDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildDescription(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		comment string
		want    string
	}{
		{"无备注", "[auto-dns]", "", "[auto-dns]"},
		{"有备注", "[auto-dns]", "生产API", "[auto-dns]生产API"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDescription(tt.prefix, tt.comment)
			if got != tt.want {
				t.Errorf("buildDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}
