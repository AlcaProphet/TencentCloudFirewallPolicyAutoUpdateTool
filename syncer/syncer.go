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
	"github.com/alcaprophet/fwalizer/provider"
)

// Syncer 同步引擎
type Syncer struct {
	cfg       *config.Config
	providers []provider.Provider
	resolver  *dns.Resolver
	configCh  chan *config.Config
	stopCh    chan struct{}
	doneCh    chan struct{} // Run 退出时关闭，用于等待当前轮次完成
}

// New 创建同步引擎
func New(cfg *config.Config, providers []provider.Provider, resolver *dns.Resolver) *Syncer {
	return &Syncer{
		cfg:       cfg,
		providers: providers,
		resolver:  resolver,
		configCh:  make(chan *config.Config, 1),
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
}

// Run 启动同步主循环（阻塞，直到收到停止信号）
func (s *Syncer) Run() {
	defer close(s.doneCh)
	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	// 启动时立即执行一次
	s.syncAll()

	for {
		select {
		case <-ticker.C:
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

	slog.Info("同步完成", "耗时", time.Since(start).Round(time.Millisecond))
}

// syncDomain 同步单个域名到单个 Provider
func (s *Syncer) syncDomain(p provider.Provider, rule config.DomainRule) {
	// 1. DNS 解析（失败则保留现有规则，不删除）
	resolved, err := s.resolver.Resolve(context.Background(), rule.Host)
	if err != nil {
		slog.Warn("DNS 解析失败，保留现有规则", "domain", rule.Host, "error", err)
		return
	}

	// 2. 带重试的完整同步流程（Describe → Diff → Create/Delete）
	if err := s.retrySync(p, rule, resolved); err != nil {
		slog.Error("同步失败", "provider", p.Name(), "domain", rule.Host, "error", err)
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
