package firewall

import (
	"log/slog"
	"sync"
	"time"

	"fwalizer/config"
	"fwalizer/dns"
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

	// 逐个域名处理（每个域名内部独立 Describe + Diff + 重试）
	for _, rule := range s.cfg.DomainRules {
		s.syncDomain(rule)
	}

	slog.Info("同步轮次完成")
}

// syncDomain 同步单个域名的规则
func (s *Syncer) syncDomain(rule config.DomainRule) {
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

	// 2. diff 对比 + 重试写入（乐观锁，最多 3 次，每次重试前重新 Describe + Diff）
	if err := s.applyWithRetry(rule.Host, rule, desc, resolved, 3); err != nil {
		slog.Error("规则同步失败", "host", rule.Host, "error", err)
	}
}

// applyWithRetry 带重试的规则写入（乐观锁）
// 重试流程：重新 Describe → 重新 diff → 重新 Create/Delete
func (s *Syncer) applyWithRetry(hostname string, rule config.DomainRule, desc string, resolved []dns.ResolvedIP, maxRetries int) error {
	var lastErr error
	for i := range maxRetries {
		if i > 0 {
			backoff := time.Duration(1<<uint(i-1)) * time.Second
			slog.Info("重试中...", "attempt", i+1, "backoff", backoff)
			time.Sleep(backoff)
		}

		// 重新拉取最新规则
		allRules, _, err := s.client.GetRules(s.cfg.InstanceID)
		if err != nil {
			lastErr = err
			slog.Warn("重新查询规则失败，将重试", "attempt", i+1, "error", err)
			continue
		}

		// 重新过滤本工具管理的规则
		owned := ownedRules(allRules, s.cfg.RuleTag, hostname)

		// 重新 diff
		toAdd, toDelete := Diff(resolved, rule, desc, owned)

		if len(toAdd) == 0 && len(toDelete) == 0 {
			slog.Info("规则已是最新，无需变更", "hostname", hostname)
			return nil
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
