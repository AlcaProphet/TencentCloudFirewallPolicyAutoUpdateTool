package firewall

import (
	"log/slog"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"

	"github.com/your-username/fwalizer/config"
	"github.com/your-username/fwalizer/dns"
)

// ruleKey 规则唯一标识 = protocol + port + cidrBlock + ipv6CidrBlock + action
type ruleKey struct {
	protocol      string
	port          string
	cidrBlock     string
	ipv6CidrBlock string
	action        string
}

// ownedRules 从全量规则中提取本工具管理的规则
// 匹配条件: FirewallRuleDescription 以 [RULE_TAG: 开头
func ownedRules(allRules []*lighthouse.FirewallRuleInfo, tagPrefix string) []*lighthouse.FirewallRuleInfo {
	var owned []*lighthouse.FirewallRuleInfo
	for _, r := range allRules {
		if r.FirewallRuleDescription == nil {
			continue
		}
		if strings.HasPrefix(*r.FirewallRuleDescription, "["+tagPrefix+":") {
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
func buildExpectedRules(hostname string, resolved []dns.ResolvedIP, rule config.DomainRule, desc string) []*lighthouse.FirewallRule {
	var rules []*lighthouse.FirewallRule

	for _, ip := range resolved {
		r := &lighthouse.FirewallRule{
			Action:                  common.StringPtr(rule.Action),
			FirewallRuleDescription: common.StringPtr(desc),
		}

		if ip.IsIPv6 {
			r.Ipv6CidrBlock = common.StringPtr(ip.Address + "/128")
		} else {
			r.CidrBlock = common.StringPtr(ip.Address + "/32")
		}

		// 根据协议类型生成规则
		switch rule.Protocol {
		case "TCP":
			r.Protocol = common.StringPtr("TCP")
			r.Port = common.StringPtr(rule.Ports)
		case "UDP":
			r.Protocol = common.StringPtr("UDP")
			r.Port = common.StringPtr(rule.Ports)
		case "TCP+UDP":
			// TCP+UDP 需要生成两条独立规则
			tcpRule := &lighthouse.FirewallRule{
				Protocol:                common.StringPtr("TCP"),
				Port:                    common.StringPtr(rule.Ports),
				Action:                  common.StringPtr(rule.Action),
				FirewallRuleDescription: common.StringPtr(desc + " TCP"),
			}
			if ip.IsIPv6 {
				tcpRule.Ipv6CidrBlock = common.StringPtr(ip.Address + "/128")
			} else {
				tcpRule.CidrBlock = common.StringPtr(ip.Address + "/32")
			}
			rules = append(rules, tcpRule)

			r.Protocol = common.StringPtr("UDP")
			r.Port = common.StringPtr(rule.Ports)
			r.FirewallRuleDescription = common.StringPtr(desc + " UDP")
		}

		rules = append(rules, r)
	}

	return rules
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
	hostname string,
	resolved []dns.ResolvedIP,
	rule config.DomainRule,
	desc string,
	existing []*lighthouse.FirewallRuleInfo,
) (toAdd []*lighthouse.FirewallRule, toDelete []*lighthouse.FirewallRule) {

	// 期望集合
	expected := buildExpectedRules(hostname, resolved, rule, desc)
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

	slog.Info("规则 diff 完成",
		"hostname", hostname,
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
