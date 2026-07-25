package syncer

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alcaprophet/fwalizer/config"
	"github.com/alcaprophet/fwalizer/dns"
	"github.com/alcaprophet/fwalizer/internal/tag"
	"github.com/alcaprophet/fwalizer/notifier"
	"github.com/alcaprophet/fwalizer/provider"
)

// Syncer 同步引擎
type Syncer struct {
	cfg       *config.Config
	providers []provider.Provider
	resolver  *dns.Resolver
	cb        *dns.CircuitBreaker
	bus       *notifier.EventBus
	configCh  chan *config.Config
	triggerCh chan struct{} // 手动触发同步
	stopCh    chan struct{}
	doneCh    chan struct{} // Run 退出时关闭，用于等待当前轮次完成

	// 状态追踪（保护以下字段）
	mu       sync.RWMutex
	running  bool
	lastSync time.Time
}

// New 创建同步引擎
func New(cfg *config.Config, providers []provider.Provider, resolver *dns.Resolver) *Syncer {
	return &Syncer{
		cfg:       cfg,
		providers: providers,
		resolver:  resolver,
		cb:        dns.NewCircuitBreaker(cfg.DNSFailThreshold),
		bus:       notifier.NewEventBus(),
		configCh:  make(chan *config.Config, 1),
		triggerCh: make(chan struct{}, 1),
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
}

// EventBus 返回事件总线（供外部订阅）
func (s *Syncer) EventBus() *notifier.EventBus {
	return s.bus
}

// Run 启动同步主循环（阻塞，直到收到停止信号）
func (s *Syncer) Run() {
	defer close(s.doneCh)
	s.setRunning(true)
	defer s.setRunning(false)

	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	// 启动时立即执行一次
	s.syncAll()

	for {
		select {
		case <-ticker.C:
			s.syncAll()
		case <-s.triggerCh:
			slog.Info("手动触发同步")
			s.syncAll()
		case newCfg := <-s.configCh:
			slog.Info("配置热重载")
			s.cfg = newCfg
			ticker.Reset(newCfg.Interval)
		case <-s.stopCh:
			slog.Info("同步引擎停止")
			return
		}
	}
}

// Stop 优雅停止
func (s *Syncer) Stop() {
	close(s.stopCh)
}

// Reload 热重载配置
func (s *Syncer) Reload(cfg *config.Config) {
	s.configCh <- cfg
}

// TriggerSync 手动触发一次同步（非阻塞）
func (s *Syncer) TriggerSync() {
	select {
	case s.triggerCh <- struct{}{}:
	default: // 已有待处理的触发，跳过
	}
}

// SyncStatus 同步状态
type SyncStatus struct {
	Running  bool       `json:"running"`
	LastSync *time.Time `json:"last_sync"`
}

// Status 返回当前同步状态
func (s *Syncer) Status() SyncStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	status := SyncStatus{Running: s.running}
	if !s.lastSync.IsZero() {
		t := s.lastSync
		status.LastSync = &t
	}
	return status
}

func (s *Syncer) setRunning(v bool) {
	s.mu.Lock()
	s.running = v
	s.mu.Unlock()
}

// DryRunResult 试运行结果
type DryRunResult struct {
	Provider string `json:"provider"`
	Domain   string `json:"domain"`
	ToAdd    int    `json:"to_add"`
	ToDelete int    `json:"to_delete"`
	Error    string `json:"error,omitempty"`
}

// DryRun 试运行：DNS 解析 + Diff，不写入不触发事件
func (s *Syncer) DryRun() ([]DryRunResult, error) {
	var results []DryRunResult
	for _, p := range s.providers {
		rules := filterRulesForTarget(s.cfg.DomainRules, p.TargetIndex())
		for _, rule := range rules {
			result := DryRunResult{Provider: p.Name(), Domain: rule.Host}

			resolved, err := s.resolver.Resolve(context.Background(), rule.Host)
			if err != nil {
				result.Error = err.Error()
				results = append(results, result)
				continue
			}

			allRules, err := p.GetRules()
			if err != nil {
				result.Error = err.Error()
				results = append(results, result)
				continue
			}

			owned := provider.OwnedRules(allRules, s.cfg.Tag)
			desc := truncateDesc(tag.Format(s.cfg.Tag, rule.Comment), p.CloudType())
			diff := provider.Diff(resolved, rule, desc, owned, p)
			result.ToAdd = len(diff.ToAdd)
			result.ToDelete = len(diff.ToDelete)
			results = append(results, result)
		}
	}
	return results, nil
}

// syncAll 执行一轮完整同步
func (s *Syncer) syncAll() {
	slog.Info("开始同步", "targets", len(s.providers), "rules", len(s.cfg.DomainRules))
	start := time.Now()

	// 按云厂商分组，跨云并行
	groups := s.groupByCloud()
	var wg sync.WaitGroup
	for ct, providers := range groups {
		wg.Add(1)
		go func(ct config.CloudType, ps []provider.Provider) {
			defer wg.Done()
			for _, p := range ps {
				rules := filterRulesForTarget(s.cfg.DomainRules, p.TargetIndex())
				for _, rule := range rules {
					s.syncDomain(p, rule)
					time.Sleep(rateLimitInterval(ct))
				}
			}
		}(ct, providers)
	}
	wg.Wait()

	s.mu.Lock()
	s.lastSync = time.Now()
	s.mu.Unlock()
	slog.Info("同步完成", "耗时", time.Since(start).Round(time.Millisecond))
}

// syncDomain 同步单个域名到单个 Provider
func (s *Syncer) syncDomain(p provider.Provider, rule config.DomainRule) {
	// 0. 熔断检查：已熔断的域名跳过（半开状态仍尝试一次探测）
	if s.cb.IsOpen(rule.Host) {
		slog.Debug("域名已熔断，半开探测", "domain", rule.Host)
	}

	// 1. DNS 解析（失败则保留现有规则，不删除）
	resolved, err := s.resolver.Resolve(context.Background(), rule.Host)
	if err != nil {
		s.cb.RecordFailure(rule.Host)
		slog.Warn("DNS 解析失败，保留现有规则", "domain", rule.Host, "error", err)
		s.bus.Publish(notifier.Event{
			Type:      notifier.EventDNSFailed,
			Timestamp: time.Now(),
			Data:      map[string]any{"domain": rule.Host, "error": err.Error()},
		})
		return
	}
	s.cb.RecordSuccess(rule.Host)

	// 2. 带重试的完整同步流程（Describe → Diff → Create/Delete）
	if err := s.retrySync(p, rule, resolved); err != nil {
		slog.Error("同步失败", "provider", p.Name(), "domain", rule.Host, "error", err)
		s.bus.Publish(notifier.Event{
			Type:      notifier.EventSyncError,
			Timestamp: time.Now(),
			Data:      map[string]any{"provider": p.Name(), "domain": rule.Host, "error": err.Error()},
		})
		return
	}

	slog.Info("同步完成", "provider", p.Name(), "domain", rule.Host)
}

func (s *Syncer) groupByCloud() map[config.CloudType][]provider.Provider {
	groups := make(map[config.CloudType][]provider.Provider)
	for _, p := range s.providers {
		ct := p.CloudType()
		groups[ct] = append(groups[ct], p)
	}
	return groups
}

// filterRulesForTarget 筛选适用于指定目标的规则
func filterRulesForTarget(rules []config.DomainRule, targetIndex int) []config.DomainRule {
	var filtered []config.DomainRule
	for _, r := range rules {
		if len(r.Targets) == 0 {
			filtered = append(filtered, r) // 空 = 所有目标
			continue
		}
		for _, t := range r.Targets {
			if t == targetIndex+1 { // Targets 是 1-based
				filtered = append(filtered, r)
				break
			}
		}
	}
	return filtered
}

// WaitForSignal 等待停止信号，并等待 Run 完全退出（确保当前轮次完成）
func WaitForSignal(s *Syncer) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh
	slog.Info("收到停止信号，等待当前轮次完成...")
	s.Stop()
	<-s.doneCh // 等待 Run 完全退出
}
