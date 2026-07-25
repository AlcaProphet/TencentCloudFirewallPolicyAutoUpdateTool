package provider

import (
	"net"
	"testing"

	"github.com/alcaprophet/fwalizer/config"
	"github.com/alcaprophet/fwalizer/dns"
	"github.com/alcaprophet/fwalizer/internal/portconv"
)

// mockProvider 用于测试的模拟 Provider
type mockProvider struct {
	cloudType   config.CloudType
	targetIndex int
}

func (m *mockProvider) Name() string              { return "mock" }
func (m *mockProvider) CloudType() config.CloudType { return m.cloudType }
func (m *mockProvider) TargetIndex() int           { return m.targetIndex }
func (m *mockProvider) GetRules() ([]config.RuleInfo, error) { return nil, nil }
func (m *mockProvider) CreateRules(rules []config.RuleAction) error { return nil }
func (m *mockProvider) DeleteRules(rules []config.RuleInfo) error   { return nil }
func (m *mockProvider) ConvertPorts(port string) []string {
	return portconv.Parse(port)
}

func TestOwnedRules(t *testing.T) {
	allRules := []config.RuleInfo{
		{Protocol: "TCP", Port: "80", CidrBlock: "1.2.3.4/32", Description: "[auto-dns] 生产API"},
		{Protocol: "TCP", Port: "443", CidrBlock: "1.2.3.4/32", Description: "[auto-dns] 生产API"},
		{Protocol: "TCP", Port: "22", CidrBlock: "0.0.0.0/0", Description: "手动规则"},
		{Protocol: "", Port: "", CidrBlock: "", Description: "[auto-dns] 模板规则"},
	}
	owned := OwnedRules(allRules, "auto-dns")
	if len(owned) != 2 {
		t.Errorf("OwnedRules 数量 = %d, want 2", len(owned))
	}
}

func TestDiff_AddNew(t *testing.T) {
	p := &mockProvider{cloudType: config.CloudTCLighthouse}
	resolved := []dns.ResolvedIP{
		{IP: net.ParseIP("1.2.3.4"), IsIPv6: false},
	}
	rule := config.DomainRule{
		Host:     "api.example.com",
		Protocol: "TCP",
		Ports:    "443",
		Action:   "ACCEPT",
	}
	desc := "[auto-dns] 生产API"

	// 无现有规则 → 应新增
	diff := Diff(resolved, rule, desc, nil, p)
	if len(diff.ToAdd) != 1 {
		t.Fatalf("ToAdd 数量 = %d, want 1", len(diff.ToAdd))
	}
	if diff.ToAdd[0].CidrBlock != "1.2.3.4/32" {
		t.Errorf("ToAdd[0].CidrBlock = %s, want 1.2.3.4/32", diff.ToAdd[0].CidrBlock)
	}
	if len(diff.ToDelete) != 0 {
		t.Errorf("ToDelete 数量 = %d, want 0", len(diff.ToDelete))
	}
}

func TestDiff_NoChange(t *testing.T) {
	p := &mockProvider{cloudType: config.CloudTCLighthouse}
	resolved := []dns.ResolvedIP{
		{IP: net.ParseIP("1.2.3.4"), IsIPv6: false},
	}
	rule := config.DomainRule{
		Host:     "api.example.com",
		Protocol: "TCP",
		Ports:    "443",
		Action:   "ACCEPT",
	}
	desc := "[auto-dns] 生产API"
	existing := []config.RuleInfo{
		{Protocol: "TCP", Port: "443", CidrBlock: "1.2.3.4/32", Action: "ACCEPT", Description: desc},
	}

	diff := Diff(resolved, rule, desc, existing, p)
	if len(diff.ToAdd) != 0 {
		t.Errorf("ToAdd 数量 = %d, want 0", len(diff.ToAdd))
	}
	if len(diff.ToDelete) != 0 {
		t.Errorf("ToDelete 数量 = %d, want 0", len(diff.ToDelete))
	}
}

func TestDiff_DeleteOld(t *testing.T) {
	p := &mockProvider{cloudType: config.CloudTCLighthouse}
	// DNS 解析到新 IP
	resolved := []dns.ResolvedIP{
		{IP: net.ParseIP("5.6.7.8"), IsIPv6: false},
	}
	rule := config.DomainRule{
		Host:     "api.example.com",
		Protocol: "TCP",
		Ports:    "443",
		Action:   "ACCEPT",
	}
	desc := "[auto-dns] 生产API"
	// 现有规则是旧 IP
	existing := []config.RuleInfo{
		{Protocol: "TCP", Port: "443", CidrBlock: "1.2.3.4/32", Action: "ACCEPT", Description: desc, RuleID: "r-123"},
	}

	diff := Diff(resolved, rule, desc, existing, p)
	if len(diff.ToAdd) != 1 {
		t.Fatalf("ToAdd 数量 = %d, want 1", len(diff.ToAdd))
	}
	if diff.ToAdd[0].CidrBlock != "5.6.7.8/32" {
		t.Errorf("ToAdd[0].CidrBlock = %s, want 5.6.7.8/32", diff.ToAdd[0].CidrBlock)
	}
	if len(diff.ToDelete) != 1 {
		t.Fatalf("ToDelete 数量 = %d, want 1", len(diff.ToDelete))
	}
	if diff.ToDelete[0].RuleID != "r-123" {
		t.Errorf("ToDelete[0].RuleID = %s, want r-123", diff.ToDelete[0].RuleID)
	}
}

func TestDiff_TCPUDPSplit(t *testing.T) {
	// Lighthouse 不支持 TCP+UDP，应拆分为两条
	p := &mockProvider{cloudType: config.CloudTCLighthouse}
	resolved := []dns.ResolvedIP{
		{IP: net.ParseIP("1.2.3.4"), IsIPv6: false},
	}
	rule := config.DomainRule{
		Host:     "game.example.com",
		Protocol: "TCP+UDP",
		Ports:    "8000",
		Action:   "ACCEPT",
	}
	desc := "[auto-dns] 游戏"

	diff := Diff(resolved, rule, desc, nil, p)
	if len(diff.ToAdd) != 2 {
		t.Fatalf("TCP+UDP 拆分后 ToAdd 数量 = %d, want 2", len(diff.ToAdd))
	}
	protocols := map[string]bool{}
	for _, a := range diff.ToAdd {
		protocols[a.Protocol] = true
	}
	if !protocols["TCP"] || !protocols["UDP"] {
		t.Errorf("应包含 TCP 和 UDP, got %v", protocols)
	}
}

func TestDiff_SWASNoSplit(t *testing.T) {
	// SWAS 原生支持 TCP+UDP，不拆分
	p := &mockProvider{cloudType: config.CloudAliSWAS}
	resolved := []dns.ResolvedIP{
		{IP: net.ParseIP("1.2.3.4"), IsIPv6: false},
	}
	rule := config.DomainRule{
		Host:     "game.example.com",
		Protocol: "TCP+UDP",
		Ports:    "8000",
		Action:   "ACCEPT",
	}
	desc := "[auto-dns] 游戏"

	diff := Diff(resolved, rule, desc, nil, p)
	if len(diff.ToAdd) != 1 {
		t.Fatalf("SWAS TCP+UDP 不拆分, ToAdd 数量 = %d, want 1", len(diff.ToAdd))
	}
	if diff.ToAdd[0].Protocol != "TCP+UDP" {
		t.Errorf("Protocol = %s, want TCP+UDP", diff.ToAdd[0].Protocol)
	}
}

func TestDiff_SWASSkipsIPv6(t *testing.T) {
	// SWAS 不支持 IPv6
	p := &mockProvider{cloudType: config.CloudAliSWAS}
	resolved := []dns.ResolvedIP{
		{IP: net.ParseIP("1.2.3.4"), IsIPv6: false},
		{IP: net.ParseIP("2001:db8::1"), IsIPv6: true},
	}
	rule := config.DomainRule{
		Host:     "api.example.com",
		Protocol: "TCP",
		Ports:    "443",
		Action:   "ACCEPT",
	}
	desc := "[auto-dns]"

	diff := Diff(resolved, rule, desc, nil, p)
	// 仅 IPv4 生成规则，IPv6 跳过
	if len(diff.ToAdd) != 1 {
		t.Fatalf("SWAS 跳过 IPv6, ToAdd 数量 = %d, want 1", len(diff.ToAdd))
	}
	if diff.ToAdd[0].CidrBlock != "1.2.3.4/32" {
		t.Errorf("CidrBlock = %s, want 1.2.3.4/32", diff.ToAdd[0].CidrBlock)
	}
}

func TestDiff_ECSSkipsICMPv6(t *testing.T) {
	// ECS 不支持 ICMPv6
	p := &mockProvider{cloudType: config.CloudAliECS}
	resolved := []dns.ResolvedIP{
		{IP: net.ParseIP("1.2.3.4"), IsIPv6: false},
		{IP: net.ParseIP("2001:db8::1"), IsIPv6: true},
	}
	rule := config.DomainRule{
		Host:     "ping.example.com",
		Protocol: "ICMP",
		Ports:    "ALL",
		Action:   "ACCEPT",
	}
	desc := "[auto-dns] ping"

	diff := Diff(resolved, rule, desc, nil, p)
	// IPv4 ICMP 生成规则，IPv6 ICMP 跳过
	if len(diff.ToAdd) != 1 {
		t.Fatalf("ECS 跳过 ICMPv6, ToAdd 数量 = %d, want 1", len(diff.ToAdd))
	}
	if diff.ToAdd[0].CidrBlock != "1.2.3.4/32" {
		t.Errorf("CidrBlock = %s, want 1.2.3.4/32", diff.ToAdd[0].CidrBlock)
	}
}

func TestDiff_DomainIsolation(t *testing.T) {
	// 不同域名的规则互不影响
	p := &mockProvider{cloudType: config.CloudTCLighthouse}
	resolved := []dns.ResolvedIP{
		{IP: net.ParseIP("1.2.3.4"), IsIPv6: false},
	}
	rule := config.DomainRule{
		Host:     "api.example.com",
		Protocol: "TCP",
		Ports:    "443",
		Action:   "ACCEPT",
	}
	desc := "[auto-dns] 生产API"
	// 现有规则属于另一个域名
	existing := []config.RuleInfo{
		{Protocol: "TCP", Port: "80", CidrBlock: "5.6.7.8/32", Action: "ACCEPT", Description: "[auto-dns] 其他域名"},
	}

	diff := Diff(resolved, rule, desc, existing, p)
	// 应新增（当前域名无规则），不删除（其他域名的规则不动）
	if len(diff.ToAdd) != 1 {
		t.Errorf("ToAdd 数量 = %d, want 1", len(diff.ToAdd))
	}
	if len(diff.ToDelete) != 0 {
		t.Errorf("ToDelete 数量 = %d, want 0（不应删除其他域名的规则）", len(diff.ToDelete))
	}
}

func TestClientPool(t *testing.T) {
	pool := NewClientPool()
	callCount := 0

	create := func() (any, error) {
		callCount++
		return "client-1", nil
	}

	// 第一次创建
	c1, err := pool.GetOrCreate("tc_lighthouse|ap-guangzhou|AKID1", create)
	if err != nil {
		t.Fatalf("GetOrCreate 失败: %v", err)
	}
	if c1 != "client-1" {
		t.Errorf("client = %v, want client-1", c1)
	}

	// 第二次复用
	c2, err := pool.GetOrCreate("tc_lighthouse|ap-guangzhou|AKID1", create)
	if err != nil {
		t.Fatalf("GetOrCreate 失败: %v", err)
	}
	if c2 != "client-1" {
		t.Errorf("client = %v, want client-1", c2)
	}

	// create 只调用一次
	if callCount != 1 {
		t.Errorf("create 调用次数 = %d, want 1", callCount)
	}
}
