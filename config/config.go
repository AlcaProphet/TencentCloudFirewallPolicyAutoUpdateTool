package config

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MaxFirewallRuleDescriptionBytes 腾讯云 API 对 FirewallRuleDescription 的字节数上限
const MaxFirewallRuleDescriptionBytes = 64

// DomainRule 单条域名规则
type DomainRule struct {
	Host     string // 域名，如 api.example.com
	Protocol string // TCP / UDP / TCP+UDP
	Ports    string // 逗号分隔端口号 或 ALL
	Action   string // ACCEPT / DROP
	Comment  string // 可选备注，用于 FirewallRuleDescription
}

// Config 应用配置
type Config struct {
	SecretID    string
	SecretKey   string
	InstanceID  string
	Region      string
	DomainRules []DomainRule
	RuleTag     string
	Interval    time.Duration
	DNSServer   string
}

// Load 从环境变量加载并校验配置
func Load() (*Config, error) {
	cfg := &Config{
		SecretID:   strings.TrimSpace(os.Getenv("TENCENTCLOUD_SECRET_ID")),
		SecretKey:  strings.TrimSpace(os.Getenv("TENCENTCLOUD_SECRET_KEY")),
		InstanceID: strings.TrimSpace(os.Getenv("LIGHTHOUSE_INSTANCE_ID")),
		Region:     strings.TrimSpace(os.Getenv("LIGHTHOUSE_REGION")),
		RuleTag:    getEnv("RULE_TAG", "auto-dns"),
		DNSServer:  getEnv("DNS_SERVER", "8.8.8.8:53"),
	}

	// 校验必填项
	if cfg.SecretID == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("TENCENTCLOUD_SECRET_ID 和 TENCENTCLOUD_SECRET_KEY 为必填项")
	}
	if cfg.InstanceID == "" {
		return nil, fmt.Errorf("LIGHTHOUSE_INSTANCE_ID 为必填项")
	}
	if !strings.HasPrefix(cfg.InstanceID, "lhins-") {
		return nil, fmt.Errorf("LIGHTHOUSE_INSTANCE_ID 格式错误，应以 lhins- 开头，实际: %s", cfg.InstanceID)
	}
	if cfg.Region == "" {
		return nil, fmt.Errorf("LIGHTHOUSE_REGION 为必填项")
	}

	// 校验 RULE_TAG：仅允许字母、数字、连字符、下划线
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, cfg.RuleTag); !matched {
		return nil, fmt.Errorf("RULE_TAG 包含非法字符（仅允许字母、数字、-、_），实际: %s", cfg.RuleTag)
	}

	// 校验 DNS_SERVER 格式
	if _, _, err := net.SplitHostPort(cfg.DNSServer); err != nil {
		return nil, fmt.Errorf("DNS_SERVER 格式错误（应为 host:port），实际: %s: %w", cfg.DNSServer, err)
	}

	// 解析 DOMAIN_RULES
	rulesRaw := strings.TrimSpace(os.Getenv("DOMAIN_RULES"))
	if rulesRaw == "" {
		return nil, fmt.Errorf("DOMAIN_RULES 为必填项")
	}
	rules, err := parseDomainRules(rulesRaw)
	if err != nil {
		return nil, fmt.Errorf("DOMAIN_RULES 解析失败: %w", err)
	}
	cfg.DomainRules = rules

	// 解析检查间隔
	intervalRaw := getEnv("CHECK_INTERVAL", "5m")
	interval, err := time.ParseDuration(intervalRaw)
	if err != nil {
		return nil, fmt.Errorf("CHECK_INTERVAL 解析失败: %w", err)
	}
	if interval < 10*time.Second {
		return nil, fmt.Errorf("CHECK_INTERVAL 不能小于 10s（API 频率保护）")
	}
	cfg.Interval = interval

	return cfg, nil
}

// RuleDescription 生成规则描述标识
// 格式: [RULE_TAG]
func (c *Config) RuleDescription() string {
	return fmt.Sprintf("[%s]", c.RuleTag)
}

// getEnv 读取环境变量，带默认值（自动去除首尾空白）
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}

// parseDomainRules 解析 "host|proto|ports|action[|comment];..." 格式
func parseDomainRules(raw string) ([]DomainRule, error) {
	segments := strings.Split(raw, ";")
	var rules []DomainRule

	for i, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}

		parts := strings.SplitN(seg, "|", 5)
		if len(parts) < 4 {
			return nil, fmt.Errorf("第 %d 条规则格式错误，期望 host|protocol|ports|action[|comment]，实际: %s", i+1, seg)
		}

		comment := ""
		if len(parts) >= 5 {
			comment = strings.TrimSpace(parts[4])
		}

		rule := DomainRule{
			Host:     strings.TrimSpace(parts[0]),
			Protocol: strings.TrimSpace(parts[1]),
			Ports:    strings.TrimSpace(parts[2]),
			Action:   strings.TrimSpace(parts[3]),
			Comment:  comment,
		}

		// 校验 hostname
		if rule.Host == "" {
			return nil, fmt.Errorf("第 %d 条规则 hostname 不能为空", i+1)
		}

		// 校验 protocol
		switch rule.Protocol {
		case "TCP", "UDP", "TCP+UDP":
		default:
			return nil, fmt.Errorf("第 %d 条规则协议不合法: %s（仅支持 TCP/UDP/TCP+UDP）", i+1, rule.Protocol)
		}

		// 校验 action
		switch rule.Action {
		case "ACCEPT", "DROP":
		default:
			return nil, fmt.Errorf("第 %d 条规则动作不合法: %s（仅支持 ACCEPT/DROP）", i+1, rule.Action)
		}

		// 校验端口号（ALL 表示全部端口，遵循腾讯云 API 规范）
		if rule.Ports != "ALL" {
			for _, portStr := range strings.Split(rule.Ports, ",") {
				portStr = strings.TrimSpace(portStr)
				port, err := strconv.Atoi(portStr)
				if err != nil || port < 1 || port > 65535 {
					return nil, fmt.Errorf("第 %d 条规则端口不合法: %s（应为 1-65535 或 ALL）", i+1, portStr)
				}
			}
		}

		rules = append(rules, rule)
	}

	if len(rules) == 0 {
		return nil, fmt.Errorf("DOMAIN_RULES 至少需要一条规则")
	}

	return rules, nil
}
