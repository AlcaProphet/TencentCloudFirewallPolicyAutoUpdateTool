# FWAlizer 功能构建计划（Build4：WebUI 体验优化 + 同步日志修复）

> **文档定位：** 本文档是「WebUI 体验优化 + 同步日志修复」（[Design3.md](./Design3.md) 十二项改进）的**当前构建方案**（依据 AGENTS.md §12.1：Build 文档为详细构建方案，非强规则）。
> - 设计构想：[Design3.md](./Design3.md)（十二项改进的根因分析、方案与验收标准；与 AGENTS.md 或用户决策冲突时以用户确认为准）
> - 架构参考：[Design1.md](./Design1.md) / [Design2.md](./Design2.md)（历史设计构想，均已归档）
> - 编码指令：[AGENTS.md](./AGENTS.md)（**唯一强要求**：简单轻量化、不过度防御、内部使用导向、中文注释、log/slog、增量添加+精确删除）
> - 历史构建与问题记录：见 [Build1.md](./Build1.md)、[Build2.md](./Build2.md)、[Build3.md](./Build3.md)、[Issue1.md](./Issue1.md)、[Issue2.md](./Issue2.md)、[Issue3.md](./Issue3.md)（均已归档）
>
> **执行原则（与 Build1-3 一致）：**
> - 每一步完成后均可编译、可测试。不跳步、不并行多步。
> - AI 执行指令：每次仅执行一个 Step，完成后运行验收命令，确认通过后再进入下一步。
> - **排序原则：先修复后构建、先安全后优化、先依赖后独立**——Step 1-3 为后端数据/日志修复阶段（计数链路 → 日志流 → 清空端点），Step 4-7 为前端构建阶段（日志页 → 主题 → 仪表盘 → 细节），Step 8 为文档同步。
> - 每步的新增逻辑必须配套单元测试（用户决策）。

---

## 一、构建进度追踪

| Step | 内容 | 设计依据 | 状态 |
|------|------|---------|------|
| 1 | 同步日志计数链路修复（`retrySync` 返回计数 → 事件 Data 携带 → `StoreLogWriter` 落库） | Design3 §四 | ✅ 验收通过（2026-08-02：build/vet/test -race 全绿；新增 TestRetrySync_Counts、TestStoreLogWriter_Counts、TestStoreLogWriter_ErrorDetail） |
| 2 | 实时日志流增强（环形缓冲历史回放 + TextHandler 格式统一 + 通道扩容） | Design3 §五 | ✅ 验收通过（2026-08-02：build/vet/test -race 全绿；新增 TestLogBroadcaster_Replay/RingOverflow/Format/LevelFilter 均 PASS） |
| 3 | 历史记录清空端点（`ClearSyncLogs` + handler + 路由注册） | Design3 §8.2 | ✅ 验收通过（2026-08-02：build/vet/test -race 全绿；新增 TestClearSyncLogs PASS） |
| 4 | 同步日志页重排（删实时事件 / 历史置顶 / 默认展开 / failed 弹窗 / 清空按钮 / 行数上限） | Design3 §七/§八 | ✅ 验收通过（2026-08-02：npm run build 零错误；go build 不受影响） |
| 5 | `useSettings` composable + 全局明暗主题（DB 持久化） | Design3 §三 | ✅ 验收通过（2026-08-02：npm run build 零错误；useSettings 含 load/refresh 双拉取） |
| 6 | 仪表盘 2×2 大卡片重设计 + 移除运行测试链接 + 首次使用引导 | Design3 §六/§9.1 | ✅ 验收通过（2026-08-02：npm run build 零错误；四卡等高/大字体大按钮/引导条） |
| 7 | 规则页「适用目标」中文云产品名 + 目标页 Keys 缺失提示 | Design3 §9.2 | ✅ 验收通过（2026-08-02：npm run build 零错误；cloudLabelMap 中文名 + NAlert 提示条） |
| 8 | 文档同步（AGENTS.md 文档体系 / Design3.md 引用） | Design3 §13.1 | ✅ 验收通过（2026-08-02：全量回归 go build/vet/test -race + npm run build 全绿；文档链接核验无残留） |

> 状态标记：☐ 未开始 / ◧ 进行中 / ✅ 验收通过

---

## 二、构建概要（文件清单总览）

| Step | 涉及文件 | 要点 |
|------|---------|------|
| 1 | `syncer/retry.go`、`syncer/syncer.go`、`webui/api/logwriter.go`、`syncer/syncer_test.go`（测试）、`webui/api/logwriter_test.go`（新增测试） | `retrySync` 返回 `(added, deleted, err)`，幂等跳过不计；`EventDomainSyncComplete` Data 增加计数；`StoreLogWriter` 读取写入 DB |
| 2 | `webui/api/logstream.go`、`webui/api/logstream_test.go`（新增测试） | 环形缓冲 1000 条 + 订阅回放；行渲染改 `slog.TextHandler`（与 stdout 逐字符一致）；订阅通道 64→1256（`logRingSize+256`） |
| 3 | `config/store.go`、`webui/api/sync.go`、`webui/api/deps.go`、`config/store_test.go`（新增测试） | `ClearSyncLogs()`；`DELETE /api/sync/logs` handler 与路由 |
| 4 | `webui/frontend/src/views/Logs.vue`、`webui/frontend/src/types.ts` | 删除实时事件版块与事件 SSE；历史记录置顶；运行日志默认展开；failed 点击弹窗；清空按钮（NPopconfirm）；logLines 上限 200→1000；`SyncLogEntry.error?` |
| 5 | `webui/frontend/src/composables/useSettings.ts`（新增）、`webui/frontend/src/App.vue` | 模块级单例设置状态（主题/凭据）；`darkTheme` 动态切换 + 侧边栏开关 + DB 持久化 |
| 6 | `webui/frontend/src/views/Dashboard.vue` | 2×2 大卡片（引擎状态/上次同步/统计概览/操作中心）；移除运行测试链接；首次引导 NAlert |
| 7 | `webui/frontend/src/views/Rules.vue`、`webui/frontend/src/views/Targets.vue` | `cloudLabelMap` 中文名；凭据状态缓存 + `watch(cloud_type)` 提示条（不阻止保存） |
| 8 | `AGENTS.md`、`Design3.md` | 文档体系表与引用更新（Build4 为当前，Build3 转历史归档） |

---

## 三、构建顺序依赖图

```
Step 1 (计数链路) ────────┐
Step 2 (日志流增强) ───────┼─→ Step 4 (日志页重排: 验收计数/回放/清空/弹窗)
Step 3 (清空端点) ────────┘        （前端验收依赖 Step 1-3 后端就绪）
Step 5 (useSettings + 主题) ──→ Step 6 (仪表盘: 复用 useSettings 引导)
                          └──→ Step 7 (Rules/Targets: 复用 useSettings 凭据状态)
Step 8 (文档同步) ── 依赖 Step 1-7 全部验收通过后
```

**关键约束：**
- **Step 1-3 必须先于 Step 4**（日志页前端要验收计数、回放、清空三项后端能力），每步完成后 `go build ./... && go vet ./... && go test ./... -race` 全绿才可继续；
- Step 4 与 Step 5 相互独立，按编号线性执行；Step 6、7 依赖 Step 5 的 `useSettings`；
- Step 8 最后执行（文档反映构建完成后的真实状态）；
- 前端每步以 `npm run build`（vue-tsc 类型检查 + vite 构建）为硬性验收。

---

## 四、分步构建计划

---

### Step 1：同步日志计数链路修复（后端）

**目标：** 打通「执行计数 → 事件 → 落库」链路，修复历史记录「新增/删除」恒为 0（Design3 §四 根因）。

**前置条件：** 无

**产出文件与操作：**

#### 1.1 `syncer/retry.go` —— `retrySync` 返回实际写入计数

将 `retrySync` 签名由 `error` 改为 `(added, deleted int, err error)`。计数规则（Design3 §4.2）：

- 累计各轮次中**云 API 调用成功**的写入量（`DeleteRules`/`CreateRules` 返回 nil）；
- 重试轮会重新 Describe → Diff（乐观锁），已生效规则不会重复出现在 diff 中，**天然避免重复计数**；
- 幂等跳过（规则已存在/已不存在）视为成功但**不计数**，与 Dry Run 的 `to_add`/`to_delete` 口径一致。

```go
// retrySync 带重试的完整同步流程（Describe → Diff → Create/Delete）
// 返回实际写入计数 (added, deleted)：累计各轮次中云 API 调用成功的写入量；
// 重试轮重新 Diff（云端状态已更新），已生效规则不重复出现，天然避免重复计数；
// 幂等跳过（规则已存在/已不存在）不计入，与 Dry Run 的 to_add/to_delete 口径一致
func (s *Syncer) retrySync(p provider.Provider, rule config.DomainRule, resolved []dns.ResolvedIP) (added, deleted int, err error) {
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
				return added, deleted, err
			}
			continue
		}

		// 2. 筛选本工具规则 + Diff
		owned := provider.OwnedRules(allRules, s.cfg.Tag)
		desc := truncateDesc(tag.Format(s.cfg.Tag, rule.Comment), p.CloudType())
		diff := provider.Diff(resolved, rule, desc, owned, p)

		// 3. 执行删除（成功才计数；幂等"已不存在"视为成功但不计数）
		if len(diff.ToDelete) > 0 {
			if err := p.DeleteRules(diff.ToDelete); err != nil {
				if isIdempotentDelete(err) {
					slog.Warn("规则已不存在，跳过", "provider", p.Name())
				} else {
					lastErr = err
					if !isRetryable(err) {
						return added, deleted, err
					}
					continue
				}
			} else {
				deleted += len(diff.ToDelete)
			}
		}

		// 4. 执行添加（成功才计数；幂等"已存在"视为成功但不计数）
		if len(diff.ToAdd) > 0 {
			if err := p.CreateRules(diff.ToAdd); err != nil {
				if isIdempotentCreate(err) {
					slog.Warn("规则已存在，跳过", "provider", p.Name())
				} else {
					lastErr = err
					if !isRetryable(err) {
						return added, deleted, err
					}
					continue
				}
			} else {
				added += len(diff.ToAdd)
			}
		}

		return added, deleted, nil // 成功
	}
	return added, deleted, lastErr
}
```

要点：失败路径携带已累计计数返回（调用方在 `err != nil` 时走 `EventSyncError`，不写计数，计数无实际用途但保持语义一致）。

#### 1.2 `syncer/syncer.go` —— `syncDomainInternal` 将计数写入成功事件

`syncDomainInternal` 内 `retrySync` 调用点（唯一调用方，已确认）改造：

```go
	added, deleted, err := s.retrySync(p, rule, resolved)
	if err != nil {
		slog.Error("同步失败", "provider", p.Name(), "domain", rule.Host, "error", err)
		s.bus.Publish(notifier.Event{
			Type:      notifier.EventSyncError,
			Timestamp: time.Now(),
			Data:      map[string]any{"provider": p.Name(), "domain": rule.Host, "error": err.Error()},
		})
		return
	}

	slog.Info("同步完成", "provider", p.Name(), "domain", rule.Host)
	s.bus.Publish(notifier.Event{
		Type:      notifier.EventDomainSyncComplete,
		Timestamp: time.Now(),
		Data:      map[string]any{"provider": p.Name(), "domain": rule.Host, "added": added, "deleted": deleted},
	})
```

#### 1.3 `webui/api/logwriter.go` —— `StoreLogWriter` 读取计数并落库

```go
// toInt 兼容事件 Data 中的数字类型（进程内为 int；事件数据若经 JSON 往返则为 float64）
func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	default:
		return 0
	}
}

// OnEvent 实现 notifier.Subscriber 接口
func (w *StoreLogWriter) OnEvent(event notifier.Event) error {
	// ... Target/Domain 提取不变 ...
	switch event.Type {
	case notifier.EventSyncError:
		log.Result = "failed"
		if v, ok := event.Data["error"].(string); ok {
			log.Error = v
		}
	case notifier.EventDomainSyncComplete:
		log.Result = "success"
		// 读取实际写入计数（Build4 Step 1：计数链路打通）
		if v, ok := event.Data["added"]; ok {
			log.Added = toInt(v)
		}
		if v, ok := event.Data["deleted"]; ok {
			log.Deleted = toInt(v)
		}
	default:
		return nil
	}
	if err := w.Store.AddSyncLog(log); err != nil {
		slog.Warn("写入同步日志失败", "error", err)
	}
	return nil
}
```

#### 1.4 单元测试

**`syncer/syncer_test.go` 追加**（需新增 `context` import；沿用现有 `stubProvider` 模式）：

```go
// ─── Build4 Step 1：计数链路测试 ───

// countingProvider 测试用 Provider：记录 CreateRules/DeleteRules 成功收到的规则数量
type countingProvider struct {
	*stubProvider
	created atomic.Int32
	deleted atomic.Int32
}

func (m *countingProvider) CreateRules(rules []config.RuleAction) error {
	m.created.Add(int32(len(rules)))
	return nil
}

func (m *countingProvider) DeleteRules(rules []config.RuleInfo) error {
	m.deleted.Add(int32(len(rules)))
	return nil
}

// TestRetrySync_Counts 空云端规则 + 单域名规则 → 全部新增，added=1 deleted=0
func TestRetrySync_Counts(t *testing.T) {
	p := &countingProvider{stubProvider: &stubProvider{cloudType: config.CloudTCCVM, targetIndex: 0}}
	cfg := &config.Config{Tag: "auto-dns"}
	s := New(cfg, []provider.Provider{p}, localResolver(t))

	resolved, err := s.resolver.Resolve(context.Background(), "localhost")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	resolved = filterIPv4(resolved) // 与 syncDomain 实际执行路径一致（LookupIPAddr 对 localhost 同时返回 127.0.0.1 与 ::1，过滤后恒为 1 条 IPv4，保证 added=1 断言稳定）
	added, deleted, err := s.retrySync(p, config.DomainRule{
		Host: "localhost", Protocol: "TCP", Ports: "443", Action: "ACCEPT", Targets: []int{0},
	}, resolved)
	if err != nil {
		t.Fatalf("retrySync 失败: %v", err)
	}
	if added != 1 || deleted != 0 {
		t.Errorf("计数 = added:%d deleted:%d, want 1/0", added, deleted)
	}
	if p.created.Load() != 1 || p.deleted.Load() != 0 {
		t.Errorf("Provider 调用 = created:%d deleted:%d, want 1/0", p.created.Load(), p.deleted.Load())
	}
}
```

**`webui/api/logwriter_test.go`（新建）**：

```go
package api

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/alcaprophet/fwalizer/config"
	"github.com/alcaprophet/fwalizer/notifier"
)

// TestStoreLogWriter_Counts 成功事件携带计数 → 落库 added/deleted 正确
func TestStoreLogWriter_Counts(t *testing.T) {
	store, err := config.OpenStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer store.Close()

	w := &StoreLogWriter{Store: store}
	if err := w.OnEvent(notifier.Event{
		Type:      notifier.EventDomainSyncComplete,
		Timestamp: time.Now(),
		Data:      map[string]any{"provider": "tc_lighthouse(lhins-abc)", "domain": "api.example.com", "added": 2, "deleted": 1},
	}); err != nil {
		t.Fatalf("OnEvent 失败: %v", err)
	}

	logs, err := store.GetSyncLogs(10)
	if err != nil || len(logs) != 1 {
		t.Fatalf("GetSyncLogs = %v, err = %v, want 1 条", logs, err)
	}
	l := logs[0]
	if l.Result != "success" || l.Added != 2 || l.Deleted != 1 {
		t.Errorf("日志 = result:%s added:%d deleted:%d, want success/2/1", l.Result, l.Added, l.Deleted)
	}
}

// TestStoreLogWriter_ErrorDetail 失败事件 → result=failed + error 落库，计数为 0
func TestStoreLogWriter_ErrorDetail(t *testing.T) {
	store, err := config.OpenStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer store.Close()

	w := &StoreLogWriter{Store: store}
	if err := w.OnEvent(notifier.Event{
		Type:      notifier.EventSyncError,
		Timestamp: time.Now(),
		Data:      map[string]any{"provider": "tc_lighthouse(lhins-abc)", "domain": "api.example.com", "error": "请求超时"},
	}); err != nil {
		t.Fatalf("OnEvent 失败: %v", err)
	}

	logs, err := store.GetSyncLogs(10)
	if err != nil || len(logs) != 1 {
		t.Fatalf("GetSyncLogs = %v, err = %v, want 1 条", logs, err)
	}
	l := logs[0]
	if l.Result != "failed" || l.Error != "请求超时" {
		t.Errorf("日志 = result:%s error:%s, want failed/请求超时", l.Result, l.Error)
	}
	if l.Added != 0 || l.Deleted != 0 {
		t.Errorf("失败记录计数应为 0, got added:%d deleted:%d", l.Added, l.Deleted)
	}
}
```

**测试与验收：**
```bash
go build ./... && go vet ./...
go test ./syncer/... ./webui/api/... -race -v
```

---

### Step 2：实时日志流增强（环形缓冲回放 + 格式统一）（后端）

**目标：** 修复实时日志与 `docker compose logs` 不一致且显示不全（Design3 §五 根因）：① 格式统一为 `slog.TextHandler` 渲染；② 订阅时回放最近 1000 条；③ 订阅通道容量扩容。

**前置条件：** 无（独立于 Step 1）

**产出文件与操作：**

#### 2.1 `webui/api/logstream.go` —— 全文件重写

```go
package api

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
)

// 环形缓冲容量：回放最近 N 条日志（与前端显示上限 1000 一致）
const logRingSize = 1000

// LogBroadcaster 将 slog 日志广播到 SSE 订阅者
// level 与 stdout 日志级别一致（debug/info/warn/error，默认 info），保证 WebUI 日志流与终端输出级别一致
// 行格式与 stdout（slog.TextHandler）逐字符一致，保证 WebUI 与 docker compose logs 输出对齐（Build4 Step 2）
// 支持历史回放：订阅时先回放环形缓冲中的最近 logRingSize 条，再进入增量推送（弥补"页面打开前的日志不显示"）
type LogBroadcaster struct {
	mu    sync.Mutex
	subs  map[int]chan string
	next  int
	level slog.Level // 日志流级别（与 cfg.LogLevel 一致）

	// 环形缓冲（最近 logRingSize 条）
	ring    [logRingSize]string
	ringPos int // 写指针（下一个写入位置）
	ringCnt int // 已写入条数（≤ logRingSize）
}

// NewLogBroadcaster 创建日志广播器（level: debug/info/warn/error 字符串）
func NewLogBroadcaster(level string) *LogBroadcaster {
	return &LogBroadcaster{subs: make(map[int]chan string), level: parseLevel(level)}
}

// parseLevel 解析日志级别字符串（与 app.InitLogger 语义一致，默认 info）
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Subscribe 订阅日志流：先回放最近 logRingSize 条历史，再返回 channel 和取消函数
func (b *LogBroadcaster) Subscribe() (<-chan string, func()) {
	b.mu.Lock()
	id := b.nextID()
	// 通道容量 ≥ 回放条数 + 增量余量：锁外回放写入不阻塞
	ch := make(chan string, logRingSize+256)
	b.subs[id] = ch

	// 拷贝环形缓冲快照（时间正序：最旧 → 最新）
	history := make([]string, 0, b.ringCnt)
	if b.ringCnt < logRingSize {
		history = append(history, b.ring[:b.ringCnt]...)
	} else {
		for i := 0; i < logRingSize; i++ {
			history = append(history, b.ring[(b.ringPos+i)%logRingSize])
		}
	}
	b.mu.Unlock()

	// 锁外回放（通道容量充足，阻塞写入不会死锁；SSE handler 在 Subscribe 返回后立即消费）
	for _, line := range history {
		ch <- line
	}

	return ch, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if c, ok := b.subs[id]; ok {
			close(c)
			delete(b.subs, id)
		}
	}
}

func (b *LogBroadcaster) nextID() int {
	id := b.next
	b.next++
	return id
}

// ─── slog.Handler 实现 ───

func (b *LogBroadcaster) Enabled(_ context.Context, level slog.Level) bool {
	return level >= b.level // 按日志流级别过滤，避免 debug 噪音
}

// renderLine 用 slog.TextHandler 渲染单行（与 stdout 格式完全一致）
// TextHandler 输出形如：time=2026-08-02T10:00:00.000+08:00 level=INFO msg=同步完成 provider=...
func renderLine(level slog.Level, r slog.Record) string {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: level})
	_ = h.Handle(context.Background(), r)
	return buf.String()
}

func (b *LogBroadcaster) Handle(_ context.Context, r slog.Record) error {
	line := renderLine(b.level, r)

	b.mu.Lock()
	defer b.mu.Unlock()

	// 1. 写入环形缓冲
	b.ring[b.ringPos] = line
	b.ringPos = (b.ringPos + 1) % logRingSize
	if b.ringCnt < logRingSize {
		b.ringCnt++
	}

	// 2. 推送订阅者（满则跳过；通道容量充足 + 回放兜底，丢失概率极低）
	for _, ch := range b.subs {
		select {
		case ch <- line:
		default:
		}
	}
	return nil
}

func (b *LogBroadcaster) WithAttrs(attrs []slog.Attr) slog.Handler { return b }
func (b *LogBroadcaster) WithGroup(name string) slog.Handler     { return b }

// ─── LogBroadcaster 已使用 app.MultiHandler（定义于 app/logutil.go） ───

// ─── SSE 端点 ───

func (d *Deps) handleLogStream(w http.ResponseWriter, r *http.Request) {
	if d.LogBroadcaster == nil {
		writeError(w, http.StatusBadRequest, "日志流不可用")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "SSE 不可用")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, unsubscribe := d.LogBroadcaster.Subscribe()
	defer unsubscribe()

	for {
		select {
		case line, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", line)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
```

改动要点对照：

| 项 | 原实现 | 新实现 |
|----|--------|--------|
| 行格式 | 自定义 `15:04:05 [INFO] msg attrs` | `slog.TextHandler` 渲染（含日期/时区/引号转义，与 stdout 一致） |
| 历史回放 | 无（订阅即增量） | 环形缓冲 1000 条，订阅时正序回放 |
| 通道容量 | 64（满则跳过） | `logRingSize+256`（回放不阻塞） |
| 锁 | `RLock`（只读 subs） | `Lock`（ring 写入 + subs 遍历） |

注意：`renderLine` 中 `_ = h.Handle(...)` 忽略渲染错误（bytes.Buffer 写入不会失败，符合 AGENTS.md「不过度防御」，且此处忽略是无害的返回码）。

#### 2.2 `webui/api/logstream_test.go`（新建）

```go
package api

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// newRecord 构造测试用日志记录
func newRecord(level slog.Level, msg string) slog.Record {
	r := slog.NewRecord(time.Now(), level, msg, 0)
	return r
}

// TestLogBroadcaster_Replay 写入 3 条 → 订阅回放 3 条且为正序（首条为最早写入）
func TestLogBroadcaster_Replay(t *testing.T) {
	b := NewLogBroadcaster("info")
	for _, msg := range []string{"第一条", "第二条", "第三条"} {
		if err := b.Handle(context.Background(), newRecord(slog.LevelInfo, msg)); err != nil {
			t.Fatalf("Handle 失败: %v", err)
		}
	}

	ch, unsub := b.Subscribe()
	defer unsub()

	for i, want := range []string{"第一条", "第二条", "第三条"} {
		select {
		case line := <-ch:
			// TextHandler 对中文消息会加引号（输出 msg="第一条"），直接检查消息文本即可
			if !strings.Contains(line, want) {
				t.Errorf("回放第 %d 条 = %q, want 包含 %s", i, line, want)
			}
		case <-time.After(time.Second):
			t.Fatalf("回放第 %d 条超时", i)
		}
	}
	// 不应有多余输出
	select {
	case line, ok := <-ch:
		if ok {
			t.Errorf("回放多出内容: %q", line)
		}
	case <-time.After(100 * time.Millisecond):
	}
}

// TestLogBroadcaster_RingOverflow 写入超过容量 → 订阅恰好回放最近 logRingSize 条（最旧的被淘汰）
func TestLogBroadcaster_RingOverflow(t *testing.T) {
	b := NewLogBroadcaster("info")
	for i := 0; i < logRingSize+5; i++ {
		if err := b.Handle(context.Background(), newRecord(slog.LevelInfo, "消息")); err != nil {
			t.Fatalf("Handle 失败: %v", err)
		}
	}

	ch, unsub := b.Subscribe()
	defer unsub()

	for i := 0; i < logRingSize; i++ {
		select {
		case <-ch:
		case <-time.After(3 * time.Second):
			t.Fatalf("回放第 %d 条超时", i)
		}
	}
	select {
	case line, ok := <-ch:
		if ok {
			t.Errorf("回放应恰好 %d 条，多出: %q", logRingSize, line)
		}
	case <-time.After(100 * time.Millisecond):
	}
}

// TestLogBroadcaster_Format 行格式与 slog.TextHandler 一致（含 level= 与 msg=）
func TestLogBroadcaster_Format(t *testing.T) {
	b := NewLogBroadcaster("info")
	if err := b.Handle(context.Background(), newRecord(slog.LevelInfo, "同步完成")); err != nil {
		t.Fatalf("Handle 失败: %v", err)
	}

	ch, unsub := b.Subscribe()
	defer unsub()

	select {
	case line := <-ch:
		// 注意：TextHandler 对中文消息加引号（输出 msg="同步完成"），此处只断言级别与消息文本
		if !strings.Contains(line, "level=INFO") || !strings.Contains(line, "同步完成") {
			t.Errorf("行格式不符合 TextHandler 规范: %q", line)
		}
	case <-time.After(time.Second):
		t.Fatal("回放超时")
	}
}

// TestLogBroadcaster_LevelFilter 级别过滤：info 级别下 debug 日志不进入缓冲
// 注意：必须经 slog.Logger 写入（先检查 Enabled），直接调用 Handle 会绕过过滤导致断言失败
func TestLogBroadcaster_LevelFilter(t *testing.T) {
	b := NewLogBroadcaster("info")
	if b.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("info 级别下 debug 应被过滤")
	}
	logger := slog.New(b)
	logger.Debug("调试") // slog.Logger 先检查 Enabled → false → 不调用 Handle，ring 保持为空

	ch, unsub := b.Subscribe()
	defer unsub()
	select {
	case line, ok := <-ch:
		if ok {
			t.Errorf("debug 日志不应进入缓冲: %q", line)
		}
	case <-time.After(100 * time.Millisecond):
	}
}
```

**测试与验收：**
```bash
go build ./... && go vet ./...
go test ./webui/api/... -race -v
# 手工验证：启动 WebUI 后打开同步日志页，应先出现最近日志（回放），且格式与 docker compose logs 一致
```

---

### Step 3：历史记录清空端点（后端）

**目标：** 提供清空 `sync_logs` 表的 `DELETE /api/sync/logs` 端点（Design3 §8.2），供 Step 4 前端「清空记录」按钮调用。

**前置条件：** 无（独立，但须在 Step 4 之前）

**产出文件与操作：**

#### 3.1 `config/store.go` —— 新增清空方法（置于 `GetSyncLogs` 之后）

```go
// ClearSyncLogs 清空全部同步历史记录（仅 sync_logs 表，不影响 targets/rules/settings）
func (s *Store) ClearSyncLogs() error {
	_, err := s.db.Exec("DELETE FROM sync_logs")
	return err
}
```

#### 3.2 `webui/api/sync.go` —— 新增 handler（置于 `handleGetSyncLogs` 之后）

```go
// handleClearSyncLogs 清空同步历史记录
func (d *Deps) handleClearSyncLogs(w http.ResponseWriter, r *http.Request) {
	if err := d.Store.ClearSyncLogs(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "历史记录已清空"})
}
```

#### 3.3 `webui/api/deps.go` —— 路由注册（`handleGetSyncLogs` 下方新增一行）

```go
	mux.HandleFunc("GET /api/sync/logs", d.handleGetSyncLogs)
	mux.HandleFunc("DELETE /api/sync/logs", d.handleClearSyncLogs)
```

#### 3.4 `config/store_test.go`（新建）

```go
package config

import (
	"path/filepath"
	"testing"
	"time"
)

// TestClearSyncLogs 写入 3 条 → 清空 → GetSyncLogs 为空
func TestClearSyncLogs(t *testing.T) {
	store, err := OpenStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer store.Close()

	for i := 0; i < 3; i++ {
		if err := store.AddSyncLog(SyncLog{Timestamp: time.Now(), Target: "lhins-abc", Result: "success"}); err != nil {
			t.Fatalf("AddSyncLog 失败: %v", err)
		}
	}
	if logs, err := store.GetSyncLogs(10); err != nil || len(logs) != 3 {
		t.Fatalf("写入后 GetSyncLogs = %d 条, err = %v, want 3", len(logs), err)
	}

	if err := store.ClearSyncLogs(); err != nil {
		t.Fatalf("ClearSyncLogs 失败: %v", err)
	}
	if logs, err := store.GetSyncLogs(10); err != nil || len(logs) != 0 {
		t.Errorf("清空后 GetSyncLogs = %d 条, err = %v, want 0", len(logs), err)
	}
}
```

**测试与验收：**
```bash
go build ./... && go vet ./...
go test ./config/... ./webui/api/... -race -v
# 手工验证：curl -X DELETE http://127.0.0.1:60200/api/sync/logs 返回 {"message":"历史记录已清空"}
```

---

### Step 4：同步日志页重排（前端）

**目标：** 落实 Design3 §七/§八 六项前端改动：删除「实时事件」版块与事件 SSE、历史记录置顶、运行日志默认展开、failed 点击弹出错误报告、清空记录按钮、logLines 上限 200→1000。

**前置条件：** Step 1-3 完成（计数/回放/清空端点已就绪，本步一并验收）

**产出文件与操作：**

#### 4.1 `webui/frontend/src/types.ts` —— `SyncLogEntry` 补充 `error` 字段

```typescript
export interface SyncLogEntry {
  timestamp: string
  target: string
  domain: string
  result: string
  added: number
  deleted: number
  error?: string // 失败详情（后端已返回，Step 4 前端消费）
}
```

#### 4.2 `webui/frontend/src/views/Logs.vue` —— 全文件重写

```vue
<script setup lang="ts">
// 同步日志页：历史记录（最顶部）+ 实时运行日志（默认展开）
// 历史记录：failed 可点击查看错误详情；支持清空（DELETE /api/sync/logs）
import { NDataTable, NTag, NCollapse, NCollapseItem, NModal, NButton, NPopconfirm, NSpace, useMessage } from 'naive-ui'
import { ref, onMounted, onUnmounted, h } from 'vue'
import { request } from '../api'
import type { SyncLogEntry } from '../types'

const logs = ref<SyncLogEntry[]>([])
const logLines = ref<string[]>([])
const message = useMessage()

// failed 错误报告弹窗
const showErrorModal = ref(false)
const errorDetail = ref<SyncLogEntry | null>(null)

let logEs: EventSource | null = null

// ─── 时间格式化（本地时区，自动检测） ───
function formatTime(ts: string): string {
  if (!ts) return '-'
  const d = new Date(ts)
  if (isNaN(d.getTime())) return ts
  const pad = (n: number) => String(n).padStart(2, '0')
  const offset = -d.getTimezoneOffset()
  const sign = offset >= 0 ? '+' : '-'
  const tzStr = `UTC${sign}${pad(Math.floor(Math.abs(offset) / 60))}:${pad(Math.abs(offset) % 60)}`
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())} ${tzStr}`
}

// ─── 生命周期 ───
onMounted(async () => {
  try {
    logs.value = await request<SyncLogEntry[]>('/api/sync/logs')
  } catch (e: any) {
    // 历史日志加载失败不阻塞 SSE 展示
    console.warn('加载同步日志失败:', e.message)
  }

  // SSE 实时日志流（订阅时后端回放最近 1000 条，见 Build4 Step 2）
  logEs = new EventSource('/api/logs/stream')
  logEs.onmessage = (e) => {
    logLines.value.push(e.data)
    if (logLines.value.length > 1000) logLines.value.shift()
  }
})

onUnmounted(() => {
  if (logEs) logEs.close()
})

// ─── 历史记录 ───
function openError(row: SyncLogEntry) {
  errorDetail.value = row
  showErrorModal.value = true
}

async function clearLogs() {
  try {
    await request('/api/sync/logs', { method: 'DELETE' })
    logs.value = []
    message.success('历史记录已清空')
  } catch (e: any) {
    message.error(`清空失败: ${e.message}`)
  }
}

const columns = [
  { title: '时间', key: 'timestamp', render: (row: any) => formatTime(row.timestamp) },
  { title: '目标', key: 'target' },
  { title: '域名', key: 'domain' },
  {
    title: '结果', key: 'result',
    render(row: any) {
      const failed = row.result === 'failed'
      const type = failed ? 'error' : row.result === 'success' ? 'success' : 'warning'
      return h(NTag, {
        type, size: 'small',
        style: failed ? 'cursor: pointer;' : '',
        onClick: failed ? () => openError(row) : undefined,
      }, { default: () => row.result })
    }
  },
  { title: '新增', key: 'added' },
  { title: '删除', key: 'deleted' },
]
</script>

<template>
  <div>
    <h2>同步日志</h2>

    <!-- 历史记录（最顶部，Build4 Step 4：改进 5） -->
    <NSpace justify="space-between" align="center">
      <h3 style="margin: 0">历史记录</h3>
      <NPopconfirm @positive-click="clearLogs">
        <template #trigger>
          <NButton size="small" type="error" tertiary>清空记录</NButton>
        </template>
        将清空全部同步历史记录，此操作不可恢复
      </NPopconfirm>
    </NSpace>
    <NDataTable :columns="columns" :data="logs" :bordered="true" :max-height="400" style="margin-top: 12px" />

    <!-- 实时运行日志（默认展开，Build4 Step 4：改进 6） -->
    <NCollapse style="margin-top: 16px" :default-expanded-names="['logs']">
      <NCollapseItem title="运行日志（实时）" name="logs">
        <pre style="max-height: 300px; overflow-y: auto; background: #1e1e1e; color: #d4d4d4; padding: 12px; border-radius: 6px; font-size: 12px; line-height: 1.6; white-space: pre-wrap; word-break: break-all;">{{ logLines.join('\n') || '等待日志输出...' }}</pre>
      </NCollapseItem>
    </NCollapse>

    <!-- failed 错误报告弹窗（Build4 Step 4：改进 9） -->
    <NModal v-model:show="showErrorModal" preset="card" title="同步失败详情" style="width: 600px">
      <p v-if="errorDetail" style="line-height: 1.9">
        <b>时间：</b>{{ formatTime(errorDetail.timestamp) }}<br />
        <b>目标：</b>{{ errorDetail.target || '-' }}<br />
        <b>域名：</b>{{ errorDetail.domain || '-' }}
      </p>
      <p style="margin-bottom: 8px"><b>错误原因：</b></p>
      <pre v-if="errorDetail?.error" style="background: #1e1e1e; color: #f44336; padding: 12px; border-radius: 6px; font-size: 12px; line-height: 1.6; white-space: pre-wrap; word-break: break-all;">{{ errorDetail.error }}</pre>
      <p v-else style="color: #999">该记录未保存错误详情</p>
    </NModal>
  </div>
</template>
```

删除清单（对比原实现）：

| 删除项 | 原位置 | 说明 |
|--------|--------|------|
| `events` ref、`es` EventSource、`/api/sync/events` 订阅与 `es.close()` | script + onMounted/onUnmounted | 实时事件版块整体移除（后端端点保留） |
| `eventTypeLabels`/`eventTagType`/`formatEventData`/`eventColumns` | script | 事件相关渲染函数与列定义 |
| 「实时事件」h3 与 NDataTable | template | 历史记录移入顶部 |

**测试与验收：**
```bash
cd webui/frontend && npm run build    # vue-tsc 类型检查 + vite 构建，零错误
# 手工验证（与 Step 1-3 联动）：
# 1. 触发一轮同步 → 历史记录「新增/删除」显示实际计数（不再为 0）
# 2. 打开页面即看到最近日志（回放），格式与 docker compose logs 一致
# 3. 点击 failed 标签 → 弹出错误详情弹窗
# 4. 点击「清空记录」→ 二次确认后表格清空
# 5. 页面无「实时事件」版块；运行日志默认展开
```

---

### Step 5：`useSettings` composable + 全局明暗主题（前端）

**目标：** 落实 Design3 §三：全局明暗主题切换（Naive UI `darkTheme` + `settings` 表 `theme` 键持久化）；同时抽出 `useSettings` 共享逻辑（主题/凭据状态），供 Step 6 引导与 Step 7 Keys 提示复用。

**前置条件：** 无（独立；但 Step 6、7 依赖本步产出）

**产出文件与操作：**

#### 5.1 `webui/frontend/src/composables/useSettings.ts`（新增，`composables/` 目录已存在）

```typescript
// useSettings 共享设置逻辑：主题与凭据状态
// 消费方：App.vue（主题切换）、Dashboard.vue（首次引导）、Targets.vue（凭据缺失提示）
// 模块级单例状态：多组件共享一份 /api/settings 数据，避免重复请求
import { ref } from 'vue'
import { request } from '../api'

const settings = ref<Record<string, string>>({})
const loaded = ref(false)
const loading = ref(false)

// 主题状态（light / dark，默认 light）
const theme = ref<'light' | 'dark'>('light')

// 凭据状态（云厂商 Key 是否已配置）
const tcReady = ref(false)
const aliReady = ref(false)

// 加载设置（幂等：已加载或加载中不重复请求；失败静默，下次调用重试）
async function load() {
  if (loaded.value || loading.value) return
  loading.value = true
  try {
    settings.value = await request<Record<string, string>>('/api/settings')
    loaded.value = true
    refreshCredentialState()
  } catch { /* 失败静默 */ } finally {
    loading.value = false
  }
}

// 应用主题：从 DB 读取并同步到 theme ref
async function applyTheme() {
  await load()
  theme.value = (settings.value.theme as 'light' | 'dark') || 'light'
}

// 切换主题：即时生效 + 持久化到 DB（失败不回滚 UI，下次进入页面以 DB 为准）
async function setTheme(v: 'light' | 'dark') {
  theme.value = v
  try {
    await request('/api/settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ theme: v }),
    })
    settings.value.theme = v
  } catch { /* 持久化失败仅跳过 */ }
}

// 刷新凭据状态（基于已加载的 settings）
function refreshCredentialState() {
  tcReady.value = !!(settings.value.tc_access_id && settings.value.tc_access_key)
  aliReady.value = !!(settings.value.ali_access_id && settings.value.ali_access_key)
}

// 强制重新拉取设置（凭据配置变更等需要新鲜数据的场景）
// 与 load 的幂等缓存不同：每次真实请求；失败保留旧数据
async function refresh() {
  loading.value = true
  try {
    settings.value = await request<Record<string, string>>('/api/settings')
    loaded.value = true
    refreshCredentialState()
  } catch { /* 失败静默，保留旧数据 */ } finally {
    loading.value = false
  }
}

export function useSettings() {
  return { settings, theme, tcReady, aliReady, load, refresh, applyTheme, setTheme, refreshCredentialState }
}
```

要点：`load()` 内部调用 `refreshCredentialState()`，消费方 `await load()` 后可直接读取 `tcReady`/`aliReady`；**需要新鲜凭据状态时（如用户在设置页配置凭据后返回本页）须用 `refresh()` 强制拉取**（Step 6 引导、Step 7 提示的挂载时机均使用 refresh）。

#### 5.2 `webui/frontend/src/App.vue` —— 全文件重写

```vue
<script setup lang="ts">
// 应用外壳：侧边栏导航 + 全局主题（light/dark，DB 持久化，Build4 Step 5）
import { NLayout, NLayoutSider, NLayoutContent, NMenu, NConfigProvider, NMessageProvider, NSwitch } from 'naive-ui'
import { darkTheme } from 'naive-ui'
import { useRouter } from 'vue-router'
import { ref, onMounted } from 'vue'
import { useSettings } from './composables/useSettings'

const router = useRouter()
const activeKey = ref('/')
const { theme, applyTheme, setTheme } = useSettings()

const menuOptions = [
  { label: '仪表盘', key: '/' },
  { label: '云资源管理', key: '/targets' },
  { label: '域名规则', key: '/rules' },
  { label: '全局设置', key: '/settings' },
  { label: '同步日志', key: '/logs' },
  { label: '运行测试', key: '/run-test' },
  { label: '告警配置', key: '/alerts' },
]

function handleMenuUpdate(key: string) {
  activeKey.value = key
  router.push(key)
}

onMounted(applyTheme)
</script>

<template>
  <NConfigProvider :theme="theme === 'dark' ? darkTheme : null">
    <NMessageProvider>
      <NLayout has-sider style="height: 100vh">
        <NLayoutSider bordered :width="200">
          <div style="padding: 16px 16px 0; font-weight: bold; font-size: 18px; display: flex; justify-content: space-between; align-items: center">
            <span>FWAlizer</span>
            <!-- 全局主题切换（Build4 Step 5：改进 1） -->
            <NSwitch size="small" :value="theme === 'dark'" @update:value="(v: boolean) => setTheme(v ? 'dark' : 'light')" />
          </div>
          <NMenu
            style="margin-top: 12px"
            :value="activeKey"
            :options="menuOptions"
            @update:value="handleMenuUpdate"
          />
        </NLayoutSider>
        <NLayoutContent content-style="padding: 24px;">
          <router-view />
        </NLayoutContent>
      </NLayout>
    </NMessageProvider>
  </NConfigProvider>
</template>
```

暗色适配说明（Design3 §3.3）：
- 运行日志 `<pre>`（Step 4）保持深色终端风格，明暗主题下均一致，不做跟随；
- `NTag`/`NButton`/`NDataTable` 等组件由 Naive UI 主题自动适配，无需逐页处理；
- 页面内硬编码灰色辅助文字（如 `#999`）在暗色下仍可读，不逐一排查（不过度防御）。

**测试与验收：**
```bash
cd webui/frontend && npm run build
# 手工验证：
# 1. 侧边栏开关切换 → 全站（含表格/弹窗/标签）即时切换
# 2. 刷新页面主题保持；重启应用主题保持
# 3. sqlite3 数据目录/config.db 查 settings 表存在 theme 键
```

---

### Step 6：仪表盘 2×2 大卡片重设计 + 首次引导（前端）

**目标：** 落实 Design3 §六（方案 B：2×2 大卡片）与 §9.1（首次使用引导）：同步引擎/上次同步/统计概览/操作中心四卡等高；大字体大按钮；移除「运行测试」链接（改进 3）；无凭据时顶部展示引导条（改进 11）。

**前置条件：** Step 5（`useSettings` 的 `load`/`tcReady`/`aliReady`）

**产出文件与操作：**

#### 6.1 `webui/frontend/src/views/Dashboard.vue` —— 全文件重写

```vue
<script setup lang="ts">
// 仪表盘：2×2 大卡片（同步引擎 / 上次同步 / 统计概览 / 操作中心）+ 首次使用引导
// 改进 3：不再放置「运行测试」入口（左侧菜单栏为唯一入口）
import { NCard, NGrid, NGi, NButton, NTag, NAlert, NTooltip, NSpace, useMessage } from 'naive-ui'
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { request } from '../api'
import { useSettings } from '../composables/useSettings'
import type { SyncStatus, SyncLogEntry } from '../types'

const status = ref<SyncStatus>({ running: false, last_sync: null, enabled: true })
const switching = ref(false) // 开关请求 loading
const stats = ref({ targets: 0, rules: 0, lastAdded: 0, lastDeleted: 0 })
const showGuide = ref(false) // 首次使用引导条
const router = useRouter()
const message = useMessage()
const { refresh, tcReady, aliReady } = useSettings()
let timer: ReturnType<typeof setInterval> | null = null

// 三态标签：enabled=true → 运行中（绿）；enabled=false → 已暂停（橙，引擎仍存活于暂停子循环）；running=false → 已停止
const engineTag = computed(() => {
  if (!status.value.running) return { type: 'default' as const, text: '已停止' }
  return status.value.enabled
    ? { type: 'success' as const, text: '运行中' }
    : { type: 'warning' as const, text: '已暂停' }
})

async function fetchStatus() {
  try {
    const s = await request<SyncStatus>('/api/sync/status')
    status.value = s
  } catch { /* 轮询失败忽略 */ }
}

// 统计概览：目标数 / 规则数 / 最近一次同步的增删（复用现有端点，零新 API）
async function fetchStats() {
  try {
    const [targets, rules, logs] = await Promise.all([
      request<any[]>('/api/targets'),
      request<any[]>('/api/rules'),
      request<SyncLogEntry[]>('/api/sync/logs'),
    ])
    stats.value = {
      targets: targets.length,
      rules: rules.length,
      lastAdded: logs.length ? (logs[0].added || 0) : 0,
      lastDeleted: logs.length ? (logs[0].deleted || 0) : 0,
    }
  } catch { /* 统计失败忽略，不阻塞主状态 */ }
}

// 首次使用引导：四类凭据均未配置时展示（改进 11）；凭据配置后自动消失
// 用 refresh() 强制拉取：用户在设置页配置凭据后返回本页，引导条应立即消失（避免模块级缓存导致旧状态）
async function checkGuide() {
  await refresh()
  showGuide.value = !(tcReady.value || aliReady.value)
}

function closeGuide() { showGuide.value = false }

async function triggerSync() {
  try {
    await request('/api/sync/trigger', { method: 'POST' })
    message.success('同步已触发')
  } catch (e: any) {
    message.error(`触发失败: ${e.message}`) // 暂停时后端返回 409
  }
  setTimeout(fetchStatus, 1000)
}

async function toggleSync() {
  switching.value = true
  try {
    await request(`/api/sync/${status.value.enabled ? 'pause' : 'resume'}`, { method: 'POST' })
    status.value.enabled = !status.value.enabled
    message.success(status.value.enabled ? '同步已开启' : '同步已暂停')
  } catch (e: any) {
    message.error(`操作失败: ${e.message}`)
  } finally {
    switching.value = false
    fetchStatus()
  }
}

onMounted(() => {
  fetchStatus()
  fetchStats()
  checkGuide()
  timer = setInterval(() => { fetchStatus(); fetchStats() }, 5000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div>
    <h2>仪表盘</h2>

    <!-- 首次使用引导（改进 11） -->
    <NAlert v-if="showGuide" type="warning" closable @close="closeGuide" style="margin-bottom: 16px">
      首次使用：请先在「全局设置」中填写云厂商 API 密钥（SecretId/SecretKey），再添加云资源目标与域名规则。
      <NButton size="small" type="primary" ghost style="margin-left: 12px" @click="router.push('/settings')">去配置</NButton>
    </NAlert>

    <!-- 2×2 大卡片（改进 2：方案 B） -->
    <NGrid :cols="2" :x-gap="16" :y-gap="16">
      <NGi>
        <NCard title="同步引擎" style="min-height: 200px">
          <div style="font-size: 28px; font-weight: 600; margin-top: 20px">{{ engineTag.text }}</div>
          <div style="margin-top: 12px">
            <NTag :type="engineTag.type" size="large" :bordered="false">{{ engineTag.text }}</NTag>
          </div>
          <div style="font-size: 13px; color: #888; margin-top: 8px">开启后按同步间隔自动执行；Dry Run 与连接测试在「运行测试」页使用</div>
        </NCard>
      </NGi>
      <NGi>
        <NCard title="上次同步" style="min-height: 200px">
          <div style="font-size: 24px; font-weight: 600; margin-top: 20px">
            {{ status.last_sync ? new Date(status.last_sync).toLocaleString() : '暂无' }}
          </div>
          <div style="font-size: 13px; color: #888; margin-top: 8px">最近一轮同步完成时间</div>
        </NCard>
      </NGi>
      <NGi>
        <NCard title="统计概览" style="min-height: 200px">
          <NSpace vertical size="large" style="margin-top: 20px">
            <div style="font-size: 20px">云资源目标 <b style="font-size: 28px">{{ stats.targets }}</b> 个</div>
            <div style="font-size: 20px">域名规则 <b style="font-size: 28px">{{ stats.rules }}</b> 条</div>
            <div style="font-size: 20px">最近同步 新增 <b style="font-size: 28px">{{ stats.lastAdded }}</b> / 删除 <b style="font-size: 28px">{{ stats.lastDeleted }}</b></div>
          </NSpace>
        </NCard>
      </NGi>
      <NGi>
        <NCard title="操作中心" style="min-height: 200px">
          <NSpace size="large" style="margin-top: 20px">
            <!-- 暂停时「立即同步」置灰 + hover 提示 -->
            <NTooltip v-if="!status.enabled">
              <template #trigger>
                <span>
                  <NButton type="primary" size="large" disabled>立即同步</NButton>
                </span>
              </template>
              请先开启同步引擎
            </NTooltip>
            <NButton v-else type="primary" size="large" @click="triggerSync">立即同步</NButton>
            <NButton :type="status.enabled ? 'warning' : 'success'" size="large" :loading="switching" @click="toggleSync">
              {{ status.enabled ? '暂停同步' : '开启同步' }}
            </NButton>
          </NSpace>
          <div style="font-size: 13px; color: #888; margin-top: 16px">同步全局开关状态与「全局设置」中持久化配置一致</div>
        </NCard>
      </NGi>
    </NGrid>
  </div>
</template>
```

**测试与验收：**
```bash
cd webui/frontend && npm run build
# 手工验证：
# 1. 四卡等高（min-height 200px），大字体大按钮，空间利用率明显提升
# 2. 仪表盘无「运行测试」入口；菜单栏入口正常
# 3. 未配置凭据时顶部出现引导条，「去配置」跳转设置页；配置凭据后刷新不再出现
# 4. 暂停时「立即同步」置灰 + hover 提示
# 5. 统计概览数字随目标/规则增删刷新
```

---

### Step 7：中文云产品名与 Keys 缺失提示（前端）

**目标：** 落实 Design3 §9.2：① 域名规则「适用目标」列显示中文云产品名（与云资源管理页一致，改进 4）；② 目标弹窗选择未配置 Keys 的平台时提示（改进 12，仅提示不阻止保存）。

**前置条件：** Step 5（`useSettings` 凭据状态）

**产出文件与操作：**

#### 7.1 `webui/frontend/src/views/Rules.vue` —— `loadTargets` 中文名

`Rules.vue` 中 `loadTargets` 函数与 import 修改：

```typescript
// import 增加（原有 import 不变）
import { cloudLabelMap } from '../constants'

async function loadTargets() {
  try {
    const data = await request<any[]>('/api/targets')
    targetOptions.value = data.map((t: any) => ({
      label: `${cloudLabelMap[t.cloud_type] || t.cloud_type} / ${t.resource_id}`,
      value: t.id,
    }))
  } catch (e: any) {
    message.error(`加载目标失败: ${e.message}`)
  }
}
```

效果：`tc_lighthouse / lhins-3j99jcrw` → `腾讯云轻量云 / lhins-3j99jcrw`（表格列与弹窗下拉同源，一处修改两处生效）。

#### 7.2 `webui/frontend/src/views/Targets.vue` —— Keys 缺失提示

script 部分修改（`loadTargets` 原名 `load`，与 useSettings 的 `load` 重名，页面函数保持原名、composable 解构重命名）：

```typescript
// import 增加
import { watch } from 'vue'
import { useRouter } from 'vue-router'
import { NAlert } from 'naive-ui'
import { useSettings } from '../composables/useSettings'

// script 内新增
const router = useRouter()
const { refresh: refreshSettings, tcReady, aliReady } = useSettings()
const credWarning = ref('')

// 依据当前表单云类型刷新凭据提示（改进 12：仅提示，不阻止保存）
function updateCredWarning() {
  const ct = form.value.cloud_type
  if (ct.startsWith('tc_')) {
    credWarning.value = tcReady.value ? '' : '腾讯云凭据未配置，请先在「全局设置」中填写 SecretId/SecretKey，否则同步将失败'
  } else if (ct.startsWith('ali_')) {
    credWarning.value = aliReady.value ? '' : '阿里云凭据未配置，请先在「全局设置」中填写 AccessKeyId/AccessKeySecret，否则同步将失败'
  } else {
    credWarning.value = ''
  }
}

// 云类型变化时刷新提示
watch(() => form.value.cloud_type, updateCredWarning)

// 挂载：加载目标 + 凭据状态（原 onMounted(load) 扩展；refreshSettings 强制拉取保证提示新鲜）
onMounted(async () => {
  await Promise.all([load(), refreshSettings()])
  updateCredWarning()
})

// openAdd / openEdit 打开弹窗时触发一次提示刷新（watch 默认不立即执行）
function openAdd() {
  // ... 原有逻辑不变 ...
  updateCredWarning()
}

function openEdit(row: TargetConfig) {
  // ... 原有逻辑不变 ...
  updateCredWarning()
}
```

template 修改（`NForm` 内、云产品选择项上方插入）：

```vue
<NForm :model="form" label-placement="left" label-width="80">
  <!-- Keys 缺失提示（改进 12） -->
  <NAlert v-if="credWarning" type="warning" style="margin-bottom: 12px">
    {{ credWarning }}
    <NButton text type="primary" size="small" style="margin-left: 8px" @click="router.push('/settings')">去设置</NButton>
  </NAlert>
  <NFormItem label="云产品">
    ...
```

**测试与验收：**
```bash
cd webui/frontend && npm run build
# 手工验证：
# 1. 域名规则页「适用目标」列显示「腾讯云轻量云 / lhins-xxx」格式
# 2. 目标弹窗选择 tc_lighthouse（未配腾讯云凭据）→ 出现警告条；保存不被阻止
# 3. 配置凭据后重新打开弹窗 → 警告条消失
```

---

### Step 8：文档同步

**目标：** 更新文档体系链接（AGENTS.md 文档体系表与引用、Design3.md 引用 Build4），使文档层级与构建完成后状态一致。

**前置条件：** Step 1-7 全部验收通过

**产出文件与操作：**

#### 8.1 `AGENTS.md` —— 4 处更新

**① 文件头引用块（L4）**：

```
> 项目设计方向见 [Design1.md](./Design1.md)、[Design2.md](./Design2.md) 与 [Design3.md](./Design3.md)（设计构想，非强制），当前构建方案见 [Build4.md](./Build4.md)，历史构建与问题记录见 [Build1.md](./Build1.md)、[Build2.md](./Build2.md)、[Build3.md](./Build3.md)、[Issue1.md](./Issue1.md)、[Issue2.md](./Issue2.md)、[Issue3.md](./Issue3.md)。
```

**② 一、项目基本信息「文档定位与优先级」（L12）**：

```
- **文档定位与优先级**：编码前先阅读本文件（强要求）。设计构想见 [Design1.md](./Design1.md)、[Design2.md](./Design2.md) 与 [Design3.md](./Design3.md)（非强制，供参考）；详细构建方案见 [Build4.md](./Build4.md)（当前）与 [Build1.md](./Build1.md)、[Build2.md](./Build2.md)、[Build3.md](./Build3.md)（历史归档）；错误与修复记录见 [Issue1.md](./Issue1.md)、[Issue2.md](./Issue2.md)、[Issue3.md](./Issue3.md)（历史归档）
```

**③ 12.1 文档定位与优先级表（L165-166）**：

```
| 设计构想 | [Design1.md](./Design1.md) / [Design2.md](./Design2.md) / [Design3.md](./Design3.md) | 设计大方向、架构构想、决策记录 | 非强制，供参考 |
| 构建方案 | [Build1.md](./Build1.md) / [Build2.md](./Build2.md) / [Build3.md](./Build3.md)（历史归档）、[Build4.md](./Build4.md)（当前） | 详细的分步构建方案与验收命令 | 非强制，执行建议 |
```

**④ 12.2 文档清单表（L182-186 区域）**：新增两行、`Build3.md` 状态改历史归档：

```
| [Design3.md](./Design3.md) | 人类（开发者/用户） | WebUI 体验优化与同步日志修复设计（设计构想） | 活跃 |
| [Build4.md](./Build4.md) | 开发者 | 当前构建方案：WebUI 体验优化 + 同步日志修复（Step 1-8） | 活跃 |
| [Build3.md](./Build3.md) | 开发者 | 同步全局开关 + 运行测试页构建（Step 1-13，已全部验收通过） | 历史归档 |
```

#### 8.2 `Design3.md` —— 文档定位处引用 Build4

`Design3.md` 末尾「实施计划」章节的文档说明改为：

```
**文档**：本文件（Design3.md）为设计构想记录；详细构建方案见 [Build4.md](./Build4.md)。
```

#### 8.3 引用核对

- `README.md`：经 grep 验证无 Build/Design 文档引用，无需修改；
- `Issue3.md`/`Build3.md` 等历史归档文档内部引用保持原样，不修改；
- 桌面端文档 [FutureDesktopDevelop.md](./FutureDesktopDevelop.md) 与本次构建无关。

**测试与验收：**
```bash
# 1. 确认 AGENTS.md 中 Build4/Design3 链接存在且指向正确文件
grep -n "Build4\|Design3" AGENTS.md
# 2. 确认所有 .md 中的文档引用路径有效（文件均存在）
ls -1 *.md | grep -E "Build4|Design3"
# 3. 确认无残留“当前构建方案见 Build3.md”表述
grep -rn "当前构建方案见 \[Build3" *.md || echo "OK: 无残留引用"
```

---

## 五、验收总览（对应 Design3 十二项改进）

| 改进 | 验收要点 | 构建步骤 |
|------|---------|---------|
| 1 明暗主题 | 切换即时生效、刷新/重启保持、Docker 多端一致（DB 存储） | Step 5 |
| 2 仪表盘卡片化 | 2×2 等高大卡片、大字体大按钮、统计概览正确 | Step 6 |
| 3 移除运行测试链接 | 仪表盘无该入口，菜单栏 `/run-test` 正常 | Step 6 |
| 4 中文云产品名 | 「腾讯云轻量云 / lhins-xxx」格式，与云资源管理页一致 | Step 7 |
| 5 日志页重排 | 历史记录最顶部，无实时事件版块 | Step 4 |
| 6 默认展开 | 进入日志页运行日志即展开 | Step 4 |
| 7 计数修复 | 一轮同步后「新增/删除」显示实际变更数（幂等不计） | Step 1 + 4 |
| 8 日志一致性 | WebUI 与 docker compose logs 同格式；打开页面即有历史回放 | Step 2 + 4 |
| 9 错误报告 | failed 标签可点击，弹窗展示时间/目标/域名/error 原文 | Step 1 + 4 |
| 10 清空记录 | 确认后历史清空，其他数据不受影响 | Step 3 + 4 |
| 11 首次引导 | 无凭据时仪表盘顶部引导条，「去配置」跳转；配置后消失 | Step 5 + 6 |
| 12 Keys 提示 | 目标弹窗选择未配凭据平台时提示，保存不阻止；测试连接后端快速失败保留 | Step 5 + 7 |

---

## 六、风险与回归注意

| 风险项 | 等级 | 说明 | 缓解措施 |
|--------|------|------|---------|
| `retrySync` 签名变更影响调用方 | 🟢 低 | 调用点仅 `syncDomainInternal` 一处（已 grep 确认） | 本构建按 Step 1 顺序实施，`go build` 兜底 |
| 环形缓冲回放死锁 | 🟡 中 | 订阅通道容量 < 回放条数时，Subscribe 内同步写可能阻塞 | 通道容量 `logRingSize+256` > 回放 1000 条；回放移出锁外；Step 2 测试覆盖溢出场景 |
| 事件 Data 计数类型 | 🟢 低 | 进程内传递为 `int`，JSON 往返后为 `float64` | `toInt` 类型开关兼容两者 |
| 主题键触发热重载 | 🟢 低 | `theme` 键不被 `LoadConfig()` 使用 | 已确认 `config.go` 不解析该键，无副作用 |
| 删除实时事件版块信息丢失 | 🟢 低 | 事件信息由历史记录 + 运行日志覆盖（Design3 §7.2） | 后端 `/api/sync/events` 端点保留，可复用 |
| 前端 `load` 函数与 useSettings 重名 | 🟢 低 | `Targets.vue` 原 `load` 与 composable 的 `load` 冲突 | 解构重命名 `refresh: refreshSettings`（Step 7 已注明） |

---

## 七、变更记录

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.0 | 2026-08-02 | 依据 Design3.md 生成初始构建方案（Step 1-8） |
| v1.1 | 2026-08-02 | 深度核验修正：① Step 1 测试补 `filterIPv4`（localhost 解析含 ::1，与 syncDomain 实际路径一致）；② Step 2 三个测试修正（中文消息被 TextHandler 加引号、LevelFilter 须经 slog.Logger 走 Enabled 检查）；③ Step 5 `useSettings` 增加 `refresh()` 强制拉取，Step 6 引导 / Step 7 提示改用 `refresh`（解决"设置页配置凭据后返回页面仍显示旧提示"） |
| v1.2 | 2026-08-02 | **构建完成**：Step 1-8 全部验收通过（后端 build/vet/test -race 全绿，前端 npm run build 零错误，全量回归通过） |

