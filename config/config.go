package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// DomainRule 单条域名规则
type DomainRule struct {
	Host     string // 域名，如 api.example.com
	Protocol string // TCP / UDP / TCP+UDP
	Ports    string // 逗号分隔端口号 或 *
	Action   string // ACCEPT / DROP
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
		SecretID:   os.Getenv("TENCENTCLOUD_SECRET_ID"),
		SecretKey:  os.Getenv("TENCENTCLOUD_SECRET_KEY"),
		InstanceID: os.Getenv("LIGHTHOUSE_INSTANCE_ID"),
		Region:     os.Getenv("LIGHTHOUSE_REGION"),
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
	if cfg.Region == "" {
		return nil, fmt.Errorf("LIGHTHOUSE_REGION 为必填项")
	}

	// 解析 DOMAIN_RULES
	rulesRaw := os.Getenv("DOMAIN_RULES")
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
// 格式: [RULE_TAG:hostname]
func (c *Config) RuleDescription(hostname string) string {
	return fmt.Sprintf("[%s:%s]", c.RuleTag, hostname)
}

// getEnv 读取环境变量，带默认值
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// parseDomainRules 解析 "host|proto|ports|action;..." 格式
func parseDomainRules(raw string) ([]DomainRule, error) {
	segments := strings.Split(raw, ";")
	var rules []DomainRule

	for i, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}

		parts := strings.SplitN(seg, "|", 4)
		if len(parts) != 4 {
			return nil, fmt.Errorf("第 %d 条规则格式错误，期望 host|protocol|ports|action，实际: %s", i+1, seg)
		}

		rule := DomainRule{
			Host:     strings.TrimSpace(parts[0]),
			Protocol: strings.TrimSpace(parts[1]),
			Ports:    strings.TrimSpace(parts[2]),
			Action:   strings.TrimSpace(parts[3]),
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

		rules = append(rules, rule)
	}

	if len(rules) == 0 {
		return nil, fmt.Errorf("DOMAIN_RULES 至少需要一条规则")
	}

	return rules, nil
}
