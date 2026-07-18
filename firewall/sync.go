package firewall

import (
	"log/slog"
	"sync"
	"time"

	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"

	"github.com/your-username/fwalizer/config"
	"github.com/your-username/fwalizer/dns"
)

// Syncer 定时同步器
type Syncer struct {
	cfg      *config.Config
	client   *Client
	resolver *dns.Resolver
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewSyncer 创建同步器
func NewSyncer(cfg *config.Config) (*Syncer, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Syncer{
		cfg:      cfg,
		client:   client,
		resolver: dns.New(cfg.DNSServer),
		stopCh:   make(chan struct{}),
	}, nil
}

// Run 启动定时同步循环
func (s *Syncer) Run() {
	s.wg.Add(1)
	defer s.wg.Done()

	// 启动后立即执行一次
	s.syncAll()

	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.syncAll()
		case <-s.stopCh:
			slog.Info("同步循环已停止")
			return
		}
	}
}

// Shutdown 优雅关闭
func (s *Syncer) Shutdown() {
	close(s.stopCh)
	s.wg.Wait()
}

// syncAll 执行一次完整的同步流程
func (s *Syncer) syncAll() {
	slog.Info("开始同步轮次")

	// 1. 查询当前全量规则
	allRules, _, err := s.client.GetRules(s.cfg.InstanceID)
	if err != nil {
		slog.Error("查询防火墙规则失败，跳过本轮", "error", err)
		return
	}

	// 2. 逐个域名处理
	for _, rule := range s.cfg.DomainRules {
		s.syncDomain(rule, allRules)
	}

	slog.Info("同步轮次完成")
}

// syncDomain 同步单个域名的规则
func (s *Syncer) syncDomain(rule config.DomainRule, allRules []*lighthouse.FirewallRuleInfo) {
	desc := s.cfg.RuleDescription(rule.Host)

	// 1. DNS 解析
	resolved, err := s.resolver.Lookup(rule.Host)
	if err != nil {
		slog.Warn("DNS 解析失败，保留现有规则不变", "host", rule.Host, "error", err)
		return
	}

	if len(resolved) == 0 {
		slog.Warn("DNS 解析结果为空，保留现有规则不变", "host", rule.Host)
		return
	}

	// 2. 过滤本工具管理的、属于该域名的规则
	owned := ownedRules(allRules, s.cfg.RuleTag, rule.Host)

	// 3. diff 对比
	toAdd, toDelete := Diff(rule.Host, resolved, rule, desc, owned)

	// 4. 重试写入（最多 3 次）
	if err := s.applyWithRetry(toAdd, toDelete, 3); err != nil {
		slog.Error("规则同步失败", "host", rule.Host, "error", err)
	}
}

// applyWithRetry 带重试的规则写入
func (s *Syncer) applyWithRetry(toAdd, toDelete []*lighthouse.FirewallRule, maxRetries int) error {
	if len(toAdd) == 0 && len(toDelete) == 0 {
		return nil
	}

	var lastErr error
	for i := range maxRetries {
		if i > 0 {
			backoff := time.Duration(1<<uint(i-1)) * time.Second
			slog.Info("重试中...", "attempt", i+1, "backoff", backoff)
			time.Sleep(backoff)

			// 重新拉取规则以获取最新版本号
			_, _, err := s.client.GetRules(s.cfg.InstanceID)
			if err != nil {
				lastErr = err
				continue
			}
		}

		// 先删后加，减少冲突窗口
		if err := s.client.DeleteRules(s.cfg.InstanceID, toDelete); err != nil {
			lastErr = err
			slog.Warn("删除规则失败，将重试", "attempt", i+1, "error", err)
			continue
		}

		if err := s.client.CreateRules(s.cfg.InstanceID, toAdd); err != nil {
			lastErr = err
			slog.Warn("添加规则失败，将重试", "attempt", i+1, "error", err)
			continue
		}

		slog.Info("规则写入成功", "added", len(toAdd), "deleted", len(toDelete))
		return nil
	}

	return lastErr
}
