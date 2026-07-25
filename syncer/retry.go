package syncer

import (
	"log/slog"
	"strings"
	"time"

	"github.com/alcaprophet/fwalizer/config"
	"github.com/alcaprophet/fwalizer/dns"
	"github.com/alcaprophet/fwalizer/internal/tag"
	"github.com/alcaprophet/fwalizer/provider"
)

const maxRetries = 3

// retrySync 带重试的完整同步流程（Describe → Diff → Create/Delete）
// 每次重试都重新获取最新规则状态（乐观锁）
func (s *Syncer) retrySync(p provider.Provider, rule config.DomainRule, resolved []dns.ResolvedIP) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			backoff := time.Duration(1<<uint(i-1)) * time.Second
			slog.Warn("重试同步", "attempt", i+1, "backoff", backoff, "provider", p.Name())
			time.Sleep(backoff)
		}

		// 1. 重新获取当前规则（乐观锁核心）
		allRules, err := p.GetRules()
		if err != nil {
			lastErr = err
			if !isRetryable(err) {
				return err
			}
			continue
		}

		// 2. 筛选本工具规则 + Diff
		owned := provider.OwnedRules(allRules, s.cfg.Tag)
		desc := truncateDesc(tag.Format(s.cfg.Tag, rule.Comment), p.CloudType())
		// ECS 不支持 ICMPv6 入站规则创建，跳过 IPv6 部分
		if rule.Protocol == "ICMP" && p.CloudType() == config.CloudAliECS {
			slog.Warn("ECS 不支持 ICMPv6 入站规则，IPv6 地址将被跳过", "domain", rule.Host)
		}
		diff := provider.Diff(resolved, rule, desc, owned, p)

		// 3. 执行删除
		if len(diff.ToDelete) > 0 {
			if err := p.DeleteRules(diff.ToDelete); err != nil {
				if isIdempotentDelete(err) {
					slog.Warn("规则已不存在，跳过", "provider", p.Name())
				} else {
					lastErr = err
					if !isRetryable(err) {
						return err
					}
					continue
				}
			}
		}

		// 4. 执行添加
		if len(diff.ToAdd) > 0 {
			if err := p.CreateRules(diff.ToAdd); err != nil {
				if isIdempotentCreate(err) {
					slog.Warn("规则已存在，跳过", "provider", p.Name())
				} else {
					lastErr = err
					if !isRetryable(err) {
						return err
					}
					continue
				}
			}
		}

		return nil // 成功
	}
	return lastErr
}

// isRetryable 判断是否可重试
func isRetryable(err error) bool {
	msg := err.Error()
	retryable := []string{
		"RequestLimitExceeded",
		"InternalError",
		"FirewallBusy",
		"timeout",
		"connection refused",
	}
	for _, r := range retryable {
		if strings.Contains(msg, r) {
			return true
		}
	}
	return false
}

// isIdempotentCreate 判断"规则已存在"
func isIdempotentCreate(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "FirewallRulesExist") ||
		strings.Contains(msg, "FirewallRuleAlreadyExist") ||
		strings.Contains(msg, "DuplicatePolicy") ||
		strings.Contains(msg, "FirewallRulesDuplicated")
}

// isIdempotentDelete 判断"规则已不存在"
func isIdempotentDelete(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "FirewallRulesNotFound") ||
		strings.Contains(msg, "InvalidParam.SecurityGroupRuleId") ||
		strings.Contains(msg, "InvalidSecurityGroupRuleId.NotFound") ||
		strings.Contains(msg, "InvalidSecurityGroupRule.RuleNotExist") ||
		strings.Contains(msg, "InvalidInstanceId.NotFound")
}

// truncateDesc 按云厂商描述字段长度限制截断（保证 [TAG] 前缀完整）
// 使用 rune 切片避免截断 UTF-8 多字节字符（如中文）
func truncateDesc(desc string, ct config.CloudType) string {
	maxLen := 0
	switch ct {
	case config.CloudTCLighthouse:
		maxLen = 64 // FirewallRuleDescription ≤ 64 字符
	default:
		return desc // 其他云厂商限制宽松，无需截断
	}
	runes := []rune(desc)
	if len(runes) <= maxLen {
		return desc
	}
	return string(runes[:maxLen])
}
