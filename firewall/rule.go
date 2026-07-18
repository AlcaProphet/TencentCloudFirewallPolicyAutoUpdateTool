package firewall

import (
	"log/slog"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"

	"github.com/alcaprophet/fwalizer/config"
	"github.com/alcaprophet/fwalizer/dns"
)

// ruleKey 规则唯一标识 = protocol + port + cidrBlock + ipv6CidrBlock + action
type ruleKey struct {
	protocol      string
	port          string
	cidrBlock     string
	ipv6CidrBlock string
	action        string
}

// ownedRules 从全量规则中提取本工具管理的、属于指定域名的规则
// 匹配条件: FirewallRuleDescription 以 [RULE_TAG:hostname] 开头
func ownedRules(allRules []*lighthouse.FirewallRuleInfo, tagPrefix, hostname string) []*lighthouse.FirewallRuleInfo {
	prefix := "[" + tagPrefix + ":" + hostname + "]"
	var owned []*lighthouse.FirewallRuleInfo
	for _, r := range allRules {
		if r.FirewallRuleDescription == nil {
			continue
		}
		if strings.HasPrefix(*r.FirewallRuleDescription, prefix) {
			owned = append(owned, r)
		}
	}
	return owned
}

// toKey 将规则信息转为比较用的 key
func toKey(r *lighthouse.FirewallRuleInfo) ruleKey {
	return ruleKey{
		protocol:      safeStr(r.Protocol),
		port:          safeStr(r.Port),
		cidrBlock:     safeStr(r.CidrBlock),
		ipv6CidrBlock: safeStr(r.Ipv6CidrBlock),
		action:        safeStr(r.Action),
	}
}

// buildExpectedRules 根据 DNS 解析结果构建期望的规则列表
// 同一域名的所有规则共享同一条描述（仅用于标识归属，不影响规则匹配）
func buildExpectedRules(resolved []dns.ResolvedIP, rule config.DomainRule, desc string) []*lighthouse.FirewallRule {
	var rules []*lighthouse.FirewallRule
	ruleDesc := buildDescription(desc, rule.Comment)

	for _, ip := range resolved {
		switch rule.Protocol {
		case "TCP":
			rules = append(rules, makeRule(ip, "TCP", rule.Ports, rule.Action, ruleDesc))
		case "UDP":
			rules = append(rules, makeRule(ip, "UDP", rule.Ports, rule.Action, ruleDesc))
		case "TCP+UDP":
			rules = append(rules, makeRule(ip, "TCP", rule.Ports, rule.Action, ruleDesc))
			rules = append(rules, makeRule(ip, "UDP", rule.Ports, rule.Action, ruleDesc))
		}
	}

	return rules
}

// buildDescription 构建规则描述: [prefix] [comment]
// 若无备注则仅保留 prefix；prefix 不会被截断
func buildDescription(prefix, comment string) string {
	if comment == "" {
		return prefix
	}
	return truncateDescription(prefix, comment, config.MaxFirewallRuleDescriptionBytes)
}

// truncateDescription 截断描述文本，prefix 不会被截断（ownedRules 依赖它做匹配）
// 仅截断 detail 部分，超出 maxBytes 时追加 "...(truncated)"
//
// 注意：若 prefix 本身已接近或超过 maxBytes，截断后的总长度可能超出 maxBytes，
// 但仍在腾讯云 API 硬限制（64 字节）之内。用户应避免使用过长的 RULE_TAG 或域名。
func truncateDescription(prefix, detail string, maxBytes int) string {
	full := prefix + " " + detail
	if len(full) <= maxBytes {
		return full
	}
	suffix := "...(truncated)"
	prefixLen := len(prefix) + 1 // +1 为 prefix 和 detail 之间的空格
	available := maxBytes - prefixLen - len(suffix)
	if available < 1 {
		available = 1
	}
	if available > len(detail) {
		available = len(detail)
	}
	return prefix + " " + detail[:available] + suffix
}

// makeRule 创建单条防火墙规则
func makeRule(ip dns.ResolvedIP, protocol, port, action, description string) *lighthouse.FirewallRule {
	r := &lighthouse.FirewallRule{
		Protocol:                common.StringPtr(protocol),
		Port:                    common.StringPtr(port),
		Action:                  common.StringPtr(action),
		FirewallRuleDescription: common.StringPtr(description),
	}
	if ip.IsIPv6 {
		r.Ipv6CidrBlock = common.StringPtr(ip.Address + "/128")
	} else {
		r.CidrBlock = common.StringPtr(ip.Address + "/32")
	}
	return r
}

// toFirewallRule 将 FirewallRuleInfo 转为可用于 Delete 的 FirewallRule
func toFirewallRule(info *lighthouse.FirewallRuleInfo) *lighthouse.FirewallRule {
	return &lighthouse.FirewallRule{
		Protocol:      info.Protocol,
		Port:          info.Port,
		CidrBlock:     info.CidrBlock,
		Ipv6CidrBlock: info.Ipv6CidrBlock,
		Action:        info.Action,
	}
}

// Diff 计算需要添加和删除的规则
func Diff(
	resolved []dns.ResolvedIP,
	rule config.DomainRule,
	desc string,
	existing []*lighthouse.FirewallRuleInfo,
) (toAdd []*lighthouse.FirewallRule, toDelete []*lighthouse.FirewallRule) {

	// 期望集合
	expected := buildExpectedRules(resolved, rule, desc)
	expectedKeys := make(map[ruleKey]bool)
	for _, r := range expected {
		expectedKeys[ruleKey{
			protocol:      safeStr(r.Protocol),
			port:          safeStr(r.Port),
			cidrBlock:     safeStr(r.CidrBlock),
			ipv6CidrBlock: safeStr(r.Ipv6CidrBlock),
			action:        safeStr(r.Action),
		}] = true
	}

	// 实际集合（只考虑本工具管理的规则）
	actualKeys := make(map[ruleKey]*lighthouse.FirewallRuleInfo)
	for _, r := range existing {
		k := toKey(r)
		actualKeys[k] = r
	}

	// 待删除: 在实际中但不在期望中
	for k, r := range actualKeys {
		if !expectedKeys[k] {
			toDelete = append(toDelete, toFirewallRule(r))
		}
	}

	// 待添加: 在期望中但不在实际中
	for _, r := range expected {
		k := ruleKey{
			protocol:      safeStr(r.Protocol),
			port:          safeStr(r.Port),
			cidrBlock:     safeStr(r.CidrBlock),
			ipv6CidrBlock: safeStr(r.Ipv6CidrBlock),
			action:        safeStr(r.Action),
		}
		if _, exists := actualKeys[k]; !exists {
			toAdd = append(toAdd, r)
		}
	}

	// Diff 是每次同步的核心比对环节，在此处记录差异摘要便于 docker logs 快速定位问题。
	// 日志位于函数内而非调用方，是为了确保无论从哪个路径（首次同步 / 重试）调用都能输出。
	slog.Info("规则 diff 完成",
		"hostname", rule.Host,
		"expected", len(expected),
		"existing", len(existing),
		"toAdd", len(toAdd),
		"toDelete", len(toDelete),
	)

	return
}

// safeStr 安全获取字符串指针的值
func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
