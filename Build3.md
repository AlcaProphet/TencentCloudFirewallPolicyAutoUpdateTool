# FWAlizer 功能构建计划（Build3：R16 修复 + 同步全局开关 + 运行测试页）

> **文档定位：** 本文档是「R16 修复」「同步全局开关」「运行测试页（Dry Run + 连接测试）」的**当前构建方案**（依据 AGENTS.md §12.1：Build 文档为详细构建方案，非强规则）。
> - 设计构想：[Design2.md](./Design2.md)（设计大方向与构想，供参考；与 AGENTS.md 或用户决策冲突时以用户确认为准）
> - 架构参考：[Design1.md](./Design1.md)（设计构想；本构建完成后需同步其特性表与页面列表的 Dry Run 描述，见 Step 12）
> - 编码指令：[AGENTS.md](./AGENTS.md)（**唯一强要求**：简单轻量化、不过度防御、内部使用导向、中文注释、log/slog、增量添加+精确删除）
> - 问题记录：[Issue3.md](./Issue3.md) §六（第16轮检查 R16-01~06，本构建 Step 1-2 为对应修复方案）；历史构建与问题记录见 [Build1.md](./Build1.md)、[Build2.md](./Build2.md)、[Issue1.md](./Issue1.md)、[Issue2.md](./Issue2.md)（均已归档）
>
> **执行原则（与 Build1/Build2 一致）：**
> - 每一步完成后均可编译、可测试。不跳步、不并行多步。
> - AI 执行指令：每次仅执行一个 Step，完成后运行验收命令，确认通过后再进入下一步。
> - **排序原则：先修复后构建、先安全后优化、先依赖后独立**——Step 1-2 为 R16 修复阶段（必须先于一切构建），Step 3-13 为 Design2 功能构建阶段。
> - 每步的新增逻辑必须配套单元测试（用户决策）。

---

## 一、构建进度追踪

| Step | 内容 | 设计依据 | 状态 |
|------|------|---------|------|
| 1 | R16-01/02 Provider 正确性修复（ICMP Diff 端口归一化 + CVM PolicyIndex fallback） | Issue3 §6.2 | ☐ 未开始 |
| 2 | R16-03~06 运行时与基础设施修复（日志流级别 / cfg 加锁 / 日志清理 / 死代码） | Issue3 §6.2 | ☐ 未开始 |
| 3 | 前端共享基建（`api.ts`/`constants.ts`/`types.ts`/`useDryRun.ts`） | Design2 §8.5 | ☐ 未开始 |
| 4 | 后端 Dry Run 升级（明细化/包装/限速/快照锁/防重入）+ 连接测试凭据提示 | Design2 §6.6/§6.7 | ☐ 未开始 |
| 5 | 运行测试页（`RunTest.vue` + `DryRunResults.vue` + 路由/菜单） | Design2 §8.5 | ☐ 未开始 |
| 6 | 页面收敛（Settings/Targets/Dashboard 接入 `api.ts`，删除 Advanced 页） | Design2 §8.6/§8.7 | ☐ 未开始 |
| 7 | 数据层：`Config.SyncEnabled` + SQLite/`.env` 读写 | Design2 §三/§四 | ☐ 未开始 |
| 8 | Syncer 暂停门控（Pause Gate + 热重载开关同步） | Design2 §五 | ☐ 未开始 |
| 9 | API 层：`pause`/`resume` 端点 + trigger 保护 + `SyncStatus.enabled` | Design2 §6.1-6.5 | ☐ 未开始 |
| 10 | 启动决策统一（`main.go`/`app.go` 始终启动 `Run()`） | Design2 §7 | ☐ 未开始 |
| 11 | Dashboard 同步开关 + 按钮状态联动 | Design2 §8.1-8.3 | ☐ 未开始 |
| 12 | 文档同步（Design1/AGENTS/README/.env.example） | Design2 §12.1 | ☐ 未开始 |
| 13 | 低优先可选附项（Design2 §12.4 A-D，可独立执行） | Design2 §12.4 | ☐ 未开始（可选） |

> 状态标记：☐ 未开始 / ◧ 进行中 / ✅ 验收通过

---

## 二、构建概要（文件清单总览）

| Step | 涉及文件 | 要点 |
|------|---------|------|
| 1 | `provider/common.go`、`provider/tc_cvm.go`、`provider/common_test.go`（测试） | R16-01：`keyOf`/`keyOfAction` 端口归一化（ICMP/ICMPv6 统一按 `ALL` 比较）；R16-02：CVM `GetRules` 无 PolicyIndex 时跳过+WARN（不再用数组索引兜底） |
| 2 | `syncer/syncer.go`、`webui/api/logstream.go`、`config/store.go`、`app/systray_stub.go`、`main.go` | R16-04：`s.cfg` 替换加锁；R16-03：`LogBroadcaster` 按 `log_level` 过滤；R16-05：`AddSyncLog` 仅超限时清理；R16-06：删除 `quitCh`/`QuitCh` 死代码 |
| 3 | `webui/frontend/src/api.ts`（新增）、`constants.ts`（新增）、`types.ts`、`composables/useDryRun.ts`（新增） | 统一 fetch 封装（res.ok 检查 + error 提取 + AbortSignal）；`cloudOptions` 共享常量；`DryRunResult`/`RuleChange`/响应包装/`SyncStatus.enabled?`/`TestConnectionResult` 类型 |
| 4 | `provider/common.go`、`syncer/syncer.go`、`webui/api/sync.go`、`webui/api/targets.go`、测试 | `RuleChange` 摘要构造；`DryRun()` 明细化 + 限速 + 快照锁 + 防重入 + warnings；handler 透传包装；凭据空值快速失败（+4 行 + `strings` import） |
| 5 | `components/DryRunResults.vue`（新增）、`views/RunTest.vue`（新增）、`main.ts`、`App.vue` | `NTabs` 双子标签（Dry Run 三级分组视图 + 连接测试表单含 15s 超时）；路由 `/run-test` + 菜单「运行测试」 |
| 6 | `views/Settings.vue`、`views/Targets.vue`、`views/Rules.vue`、`views/Logs.vue`、`views/Dashboard.vue`、`views/Advanced.vue`（删除）、`main.ts`、`App.vue` | 全站接入 `api.ts`（修复 5 处"失败误报成功"）；Targets 弹窗测试连接保留 + 超时；Dashboard 移除试运行 + 「运行测试」链接 + B5 预留；删除 `/advanced` 路由/菜单/文件 |
| 7 | `config/config.go`、`config/store.go`、`config/env.go`、`config/env_test.go`（测试） | `Config.SyncEnabled`；`LoadConfig()` 读 `sync_enabled`；`ParseEnv()` 读 `SYNC_ENABLED` |
| 8 | `syncer/syncer.go`、`syncer/syncer_test.go`（测试） | `syncEnabled` 镜像 + `pauseCh`/`resumeCh` + `Pause()`/`Resume()`/`IsEnabled()` + `Run()` 启动门控 + `waitForResume()` + `Reload()` 开关同步（5.6） |
| 9 | `webui/api/deps.go`、`webui/api/sync.go`、`syncer/syncer.go` | `Syncer` 接口 +`Pause`/`Resume`；`POST /api/sync/pause|resume`；`handleSyncTrigger` 暂停 409；`SyncStatus.Enabled` |
| 10 | `main.go`、`app/app.go` | 确认始终 `go s.Run()`（门控在 Run() 内）；编译/双模式运行验证 |
| 11 | `views/Dashboard.vue`、`types.ts` | 暂停/开启开关按钮；「立即同步」disabled 联动；三态状态标签 |
| 12 | `Design1.md`、`AGENTS.md`、`README.md`、`.env.example`、`docker-compose.yml.example` | 特性表/页面列表措辞；十一节新增端点/页面/开关说明；SYNC_ENABLED；端口/页面文档 |
| 13 | `config/env.go`/`config/store.go`、`views/Settings.vue`、`main.go` | 附项 A-D（见 Step 13 详述） |

---

## 三、构建顺序依赖图

```
Step 1 (R16-01/02: Provider 正确性修复) ── 修复阶段，必须最先
  └─ Step 2 (R16-03~06: 运行时与基础设施修复) ── 依赖 Step 1（同阶段串行）
       └─ Step 3 (前端共享基建: api.ts + constants.ts + types.ts + useDryRun.ts)
            ├─ Step 4 (后端 DryRun 升级 + 凭据提示) ── 依赖 Step 3 的接口约定
            │    └─ Step 5 (运行测试页) ── 依赖 Step 3, 4
            │         └─ Step 6 (页面收敛: 全站 api.ts + 删 Advanced) ── 依赖 Step 3, 5
            └─ Step 7 (数据层: SyncEnabled) ── 无依赖
                 └─ Step 8 (Syncer 暂停门控) ── 依赖 Step 7
                      └─ Step 9 (API: pause/resume) ── 依赖 Step 7, 8
                           └─ Step 10 (启动决策) ── 依赖 Step 8
                                └─ Step 11 (Dashboard 开关) ── 依赖 Step 9, 10（复用 Step 6 的 B5 预留）
Step 12 (文档同步) ── 依赖 Step 4-11 全部完成后
Step 13 (附项 A-D) ── 任意位置，各自独立
```

**关键约束：**
- **Step 1-2（R16 修复阶段）必须先于一切构建完成**，每步完成后 `go build && go vet && go test -race` 全绿才可继续
- Step 3-11 按编号线性串行（先测试功能后全局开关；Phase 0 全部完成后再进入 Phase 1-5）
- Step 7-10 构成全局开关主干，必须严格串行（数据层 → 门控 → API → 启动）
- Step 13 各项互相独立，可穿插执行，不作为主链路验收条件

---

## 四、分步构建计划

---

### Step 1：R16-01/02 Provider 正确性修复（修复阶段）

**目标：** 修复 ICMP 规则 Diff 永不收敛（高）与 CVM PolicyIndex fallback 误删风险（中），保证 Diff/删除语义正确后再进入任何构建。

**前置条件：** 无

**产出文件与操作：**

#### 1.1 `provider/common.go` —— R16-01：端口比较归一化

```go
// normalizePortForCompare 端口比较归一化：
// 协议为 ICMP/ICMPv6 时，-1/-1、ALL、空串三者等价（desired 侧为云格式、existing 侧为归一化格式，避免永不收敛）
func normalizePortForCompare(protocol, port string) string {
	proto := strings.ToUpper(protocol)
	if proto == "ICMP" || proto == "ICMPV6" {
		return "ALL"
	}
	if strings.EqualFold(port, "-1/-1") {
		return "ALL"
	}
	return strings.ToUpper(port)
}
```

`keyOf` 与 `keyOfAction` 的 `port` 字段改为：
```go
port: normalizePortForCompare(r.Protocol, r.Port),
// 以及 keyOfAction: port: normalizePortForCompare(r.Protocol, r.Port),
```

要点：
- 归一化仅用于**比较**，不影响 CreateRules/GetRules 的请求/返回格式（两端对称，不改 API 行为）
- 覆盖四云：SWAS/ECS（`-1/-1` vs `ALL`，确定缺陷）、CVM（`ALL` vs 空串）、Lighthouse（`ALL` vs `ALL`，防御）
- **回归验证**：现有 `TestDiff_*` 全部保持通过（非 ICMP 规则比较路径不变）

#### 1.2 `provider/tc_cvm.go` —— R16-02：PolicyIndex fallback 改为跳过+WARN

```go
// GetRules 中：不再使用 Ingress 数组索引作为 PolicyIndex 兜底
for _, r := range policySet.Ingress {
	if r.PolicyIndex == nil {
		slog.Warn("CVM 规则缺少 PolicyIndex，跳过该规则（避免误删）", "description", strVal(r.PolicyDescription))
		continue
	}
	info := config.RuleInfo{
		Protocol:    strings.ToUpper(strVal(r.Protocol)),
		Port:        strVal(r.Port),
		CidrBlock:   strVal(r.CidrBlock),
		Ipv6CidrBlock: strVal(r.Ipv6CidrBlock),
		Action:      strings.ToUpper(strVal(r.Action)),
		Description: strVal(r.PolicyDescription),
		PolicyIndex: strconv.FormatInt(*r.PolicyIndex, 10),
	}
	rules = append(rules, info)
}
```

要点：CVM 的 PolicyIndex 是安全组**全方向全局索引**，Ingress 数组索引与之不一致，删除时按错误索引定位可能误删 Egress 规则；缺失时跳过该规则（不参与本工具删除），主路径不受影响。

#### 1.3 单元测试

`provider/common_test.go` 新增：
- `TestDiff_ICMPPortNormalize`：desired 端口 `-1/-1`（SWAS 场景）对 existing 端口 `ALL` → Diff 应为空（无 toAdd/toDelete）
- `TestDiff_ICMPCVM`：desired `ALL` 对 existing 空串（CVM 场景）→ 空 Diff
- `TestDiff_NonICMPUnchanged`：TCP 规则端口比较行为不变（回归）

**验收：**
```bash
go build ./... && go vet ./...
go test ./provider/... -v
```

---

### Step 2：R16-03~06 运行时与基础设施修复（修复阶段）

**目标：** 修复 cfg 并发竞态（中）、日志流级别不一致（中）、日志清理 O(n)（低）、死代码（低）。

**前置条件：** Step 1 完成

**产出文件与操作：**

#### 2.1 `syncer/syncer.go` —— R16-04：`s.cfg` 替换加锁

```go
// Run() 主循环 configCh 分支（当前 L78 附近）：
case newCfg := <-s.configCh:
	slog.Info("配置热重载")
	s.mu.Lock()
	s.cfg = newCfg // 加锁替换：消除与 syncAll()/DryRun() 并发读取的竞态
	s.mu.Unlock()
	ticker.Reset(newCfg.Interval)
```

> 衔接说明：本步为 Step 4（DryRun 快照 RLock）与 Step 8（Pause Gate 改造）的读写锁配对基础；后续 Step 8 再次改造此分支时**保持锁结构不变**。

#### 2.2 `webui/api/logstream.go` + `main.go` —— R16-03：LogBroadcaster 级别过滤

```go
// LogBroadcaster 增加级别字段：
type LogBroadcaster struct {
	mu    sync.RWMutex
	subs  map[int]chan string
	next  int
	level slog.Level // 新增：日志流级别（与 cfg.LogLevel 一致）
}

// NewLogBroadcaster 创建日志广播器（level: debug/info/warn/error 字符串）
func NewLogBroadcaster(level string) *LogBroadcaster {
	return &LogBroadcaster{subs: make(map[int]chan string), level: parseLevel(level)}
}

func (b *LogBroadcaster) Enabled(_ context.Context, level slog.Level) bool {
	return level >= b.level // 原为恒 >= Debug
}

// parseLevel 与 app.InitLogger 的级别解析一致（debug/info/warn/error，默认 info）
```

`main.go` L73 调用处同步更新：`logBroadcaster := webapi.NewLogBroadcaster(cfg.LogLevel)`。

#### 2.3 `config/store.go` —— R16-05：AddSyncLog 清理优化

```go
// AddSyncLog 中：仅当超过保留上限时执行清理
if err := s.db.Exec("INSERT INTO sync_logs ..."); err != nil { return err }
var count int
if err := s.db.QueryRow("SELECT COUNT(*) FROM sync_logs").Scan(&count); err == nil && count > 1000 {
	_, err = s.db.Exec("DELETE FROM sync_logs WHERE id NOT IN (SELECT id FROM sync_logs ORDER BY id DESC LIMIT 1000)")
}
```

#### 2.4 `app/systray_stub.go` —— R16-06：死代码清理

删除包级 `quitCh` 与 `QuitCh()`（桌面端搁置后无引用），保留 `RunSystray` 空实现（维持 `//go:build !desktop` 与 `desktop/systray.go` 的标签配对）。

**测试与验收：**
```bash
go build ./... && go vet ./... && go test ./... -race
# 手动验证：LOG_LEVEL=info 启动 WebUI 后日志流不含 debug 日志
# 验证 AddSyncLog：连续写入 1100 条日志后行数 ≤ 1000
```

---

### Step 3：前端共享基建

**目标：** 建立统一请求层与共享常量/类型，消除全站裸 `fetch` 与重复定义。

**前置条件：** Step 2 完成

**产出文件：**

#### 3.1 `webui/frontend/src/api.ts`（新增，`api.ts` 位于 `src/` 根）—— 统一 fetch 封装

```typescript
// 统一请求封装：res.ok 检查 + error 提取 + 可选超时
export class RequestError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

export async function request<T>(url: string, opts: RequestInit = {}, timeoutMs?: number): Promise<T> {
  const controller = new AbortController()
  const timer = timeoutMs ? setTimeout(() => controller.abort(), timeoutMs) : null
  try {
    const res = await fetch(url, { ...opts, signal: controller.signal })
    let data: any = null
    try { data = await res.json() } catch { /* 非 JSON 响应（如 SPA 兜底 HTML） */ }
    if (!res.ok) {
      throw new RequestError(res.status, data?.error || `请求失败 (${res.status})`)
    }
    return data as T
  } catch (e: any) {
    if (e instanceof RequestError) throw e
    if (e?.name === 'AbortError') throw new Error('请求超时')
    throw e
  } finally {
    if (timer) clearTimeout(timer)
  }
}
```

要点：所有非 2xx 统一抛 `RequestError`（修复"失败误报成功"）；`timeoutMs` 可选（连接测试传 `15000`）；未知 API 路径返回 HTML 时解析失败进 catch。

#### 3.2 `webui/frontend/src/constants.ts`（新增）—— 共享常量

```typescript
import type { SelectOption } from 'naive-ui'

export const cloudOptions: SelectOption[] = [
  { label: '腾讯云轻量云', value: 'tc_lighthouse' },
  { label: '腾讯云CVM', value: 'tc_cvm' },
  { label: '阿里云轻量云', value: 'ali_swas' },
  { label: '阿里云ECS', value: 'ali_ecs' },
]

export const cloudLabelMap: Record<string, string> = Object.fromEntries(
  cloudOptions.map((o) => [String(o.value), o.label])
)
```

（消除 `Targets.vue` 与 `Advanced.vue` 的两处重复定义。）

#### 3.3 `webui/frontend/src/types.ts` —— 更新

```typescript
// 新增/修改（保留现有接口不变）
export interface RuleChange {
  protocol: string
  port: string
  action: string
  cidr: string      // IPv4 或 IPv6 的 CIDR
  desc: string      // 规则描述（含 [TAG]）
}

export interface DryRunResult {          // 破坏性变更：to_add/to_delete 由 number 改为数组
  provider: string
  domain: string
  to_add: RuleChange[]
  to_delete: RuleChange[]
  error?: string
}

export interface DryRunResponse {        // 响应包装（Step 4 后端配套）
  results: DryRunResult[]
  warnings: string[]
}

export interface SyncStatus {            // enabled 为可选字段（B5 预留，后端未返回时默认 true）
  running: boolean
  last_sync: string | null
  enabled?: boolean
}

export interface TestConnectionResult {  // 可选
  success: boolean
  message?: string
  error?: string
}
```

#### 3.4 `webui/frontend/src/composables/useDryRun.ts`（新增，`composables/` 为新建目录）—— Dry Run 共享组合逻辑

```typescript
import { ref } from 'vue'
import { request } from '../api'
import type { DryRunResult } from '../types'

export function useDryRun() {
  const loading = ref(false)
  const results = ref<DryRunResult[]>([])
  const warnings = ref<string[]>([])
  const error = ref('')
  const lastRunAt = ref<Date | null>(null)

  async function run() {
    loading.value = true
    error.value = ''
    results.value = []
    warnings.value = []
    try {
      const data = await request<{ results: DryRunResult[]; warnings: string[] }>(
        '/api/sync/dryrun', { method: 'POST' }
      )
      results.value = data.results || []
      warnings.value = data.warnings || []
      lastRunAt.value = new Date()
    } catch (e: any) {
      error.value = e.message
      throw e // 由页面 message.error 展示
    } finally {
      loading.value = false
    }
  }
  return { loading, results, warnings, error, lastRunAt, run }
}
```

**测试与验收：**
```bash
cd webui/frontend && npm run build    # vue-tsc 类型检查 + vite 构建，零错误
cd ../.. && go build ./...            # 后端不受影响，保持可编译
```

---

### Step 4：后端 Dry Run 升级 + 连接测试凭据提示

**目标：** `POST /api/sync/dryrun` 响应结构升级（方案 A：明细化 + 包装对象），补齐限速/快照锁/防重入；连接测试凭据空值快速失败。

**前置条件：** Step 3 完成（`types.ts` 契约先行）

**产出文件与操作：**

#### 4.1 `provider/common.go` —— `RuleChange` 摘要类型与构造函数

```go
// RuleChange 规则变更摘要（供前端直接渲染）
type RuleChange struct {
	Protocol string `json:"protocol"`
	Port     string `json:"port"`
	Action   string `json:"action"`
	Cidr     string `json:"cidr"` // IPv4 或 IPv6 的 CIDR（如 1.2.3.4/32）
	Desc     string `json:"desc"` // 规则描述（含 [TAG]）
}

// RuleChangeFromAction 从期望规则构造摘要（to_add）
func RuleChangeFromAction(a config.RuleAction) RuleChange {
	cidr := a.CidrBlock
	if cidr == "" {
		cidr = a.Ipv6CidrBlock
	}
	return RuleChange{Protocol: a.Protocol, Port: a.Port, Action: a.Action, Cidr: cidr, Desc: a.Description}
}

// RuleChangeFromInfo 从云端规则构造摘要（to_delete）
func RuleChangeFromInfo(r config.RuleInfo) RuleChange {
	cidr := r.CidrBlock
	if cidr == "" {
		cidr = r.Ipv6CidrBlock
	}
	return RuleChange{Protocol: r.Protocol, Port: r.Port, Action: r.Action, Cidr: cidr, Desc: r.Description}
}
```

#### 4.2 `syncer/syncer.go` —— `DryRun()` 升级（明细化 + 限速 + 快照锁 + 防重入 + warnings）

```go
// ErrDryRunInProgress 防重入冲突错误
// 注：syncer.go 需新增 import "errors"
var ErrDryRunInProgress = errors.New("Dry Run 正在执行中")

// DryRunResponse 试运行响应（包装对象：空状态语义化）
type DryRunResponse struct {
	Results  []DryRunResult `json:"results"`
	Warnings []string       `json:"warnings"`
}

// DryRunResult 试运行结果（升级后：to_add/to_delete 为规则数组）
type DryRunResult struct {
	Provider string               `json:"provider"`
	Domain   string               `json:"domain"`
	ToAdd    []provider.RuleChange `json:"to_add"`
	ToDelete []provider.RuleChange `json:"to_delete"`
	Error    string               `json:"error,omitempty"`
}

// Syncer 结构体新增字段：dryRunMu sync.Mutex（防重入）

// DryRun 试运行：DNS 解析 + Diff，不写入不触发事件
func (s *Syncer) DryRun() (DryRunResponse, error) {
	if !s.dryRunMu.TryLock() {
		return DryRunResponse{}, ErrDryRunInProgress
	}
	defer s.dryRunMu.Unlock()

	// 快照：RLock 保护 providers/cfg（cfg 替换写锁已在 Step 2 实施，读写配对）
	s.mu.RLock()
	providers := s.providers
	cfg := s.cfg
	s.mu.RUnlock()

	resp := DryRunResponse{Results: []DryRunResult{}}
	if len(providers) == 0 {
		resp.Warnings = append(resp.Warnings, "暂无云资源目标，请先在云资源管理页配置")
	}
	if len(cfg.DomainRules) == 0 {
		resp.Warnings = append(resp.Warnings, "暂无域名规则，请先在域名规则页配置")
	}
	for _, p := range providers {
		rules := filterRulesForTarget(cfg.DomainRules, p.TargetIndex())
		for _, rule := range rules {
			result := DryRunResult{Provider: p.Name(), Domain: rule.Host}
			resolved, err := s.resolver.Resolve(context.Background(), rule.Host)
			if err != nil {
				result.Error = err.Error()
				resp.Results = append(resp.Results, result)
				continue
			}
			if !rule.EnableIPv6 {
				resolved = filterIPv4(resolved)
			}
			allRules, err := p.GetRules()
			if err != nil {
				result.Error = err.Error()
				resp.Results = append(resp.Results, result)
				continue
			}
			owned := provider.OwnedRules(allRules, cfg.Tag)
			desc := truncateDesc(tag.Format(cfg.Tag, rule.Comment), p.CloudType())
			diff := provider.Diff(resolved, rule, desc, owned, p)
			for _, a := range diff.ToAdd {
				result.ToAdd = append(result.ToAdd, provider.RuleChangeFromAction(a))
			}
			for _, r := range diff.ToDelete {
				result.ToDelete = append(result.ToDelete, provider.RuleChangeFromInfo(r))
			}
			resp.Results = append(resp.Results, result)
			time.Sleep(rateLimitInterval(p.CloudType())) // 限速：与 syncAll 一致
		}
	}
	return resp, nil
}
```

要点：
- **限速**：循环内复用 `rateLimitInterval(p.CloudType())`（AGENTS.md §七）
- **快照锁**：`s.mu.RLock()` 读快照，与 Step 2 的 cfg 写锁配对（竞态消除）
- **防重入**：`dryRunMu.TryLock()`，冲突返回 `ErrDryRunInProgress` → handler 转 409
- **warnings**：Syncer 层生成，handler 只透传（Design2 §6.6）
- 空结果时 `Results` 初始化为 `[]`（保证 JSON 输出 `[]`）

#### 4.3 `webui/api/sync.go` —— `handleSyncDryRun` 透传包装 + 409

```go
// 注：webui/api/sync.go 需新增 import "errors"（errors.Is 使用）
func (d *Deps) handleSyncDryRun(w http.ResponseWriter, r *http.Request) {
	if d.Syncer == nil {
		writeError(w, http.StatusBadRequest, "同步引擎未启动，请先配置目标和规则")
		return
	}
	resp, err := d.Syncer.DryRun()
	if err != nil {
		if errors.Is(err, syncer.ErrDryRunInProgress) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
```

同步更新 `webui/api/deps.go` 的 `Syncer` 接口签名：`DryRun() (syncer.DryRunResponse, error)`。

#### 4.4 `webui/api/targets.go` —— 凭据空值快速失败（+4 行 + import）

```go
// handleTestConnection 中 SetCredentials 之前插入（需新增 "strings" import）
if strings.HasPrefix(req.CloudType, "tc_") && (settings["tc_access_id"] == "" || settings["tc_access_key"] == "") {
	writeJSON(w, http.StatusOK, map[string]any{"success": false, "error": "腾讯云凭据未配置，请先在全局设置中填写"})
	return
}
if strings.HasPrefix(req.CloudType, "ali_") && (settings["ali_access_id"] == "" || settings["ali_access_key"] == "") {
	writeJSON(w, http.StatusOK, map[string]any{"success": false, "error": "阿里云凭据未配置，请先在全局设置中填写"})
	return
}
```

#### 4.5 单元测试

`provider/common_test.go` 新增：
- `TestRuleChangeFromAction`：IPv4/IPv6 的 Cidr 取值正确、Desc 透传
- `TestRuleChangeFromInfo`：同上（to_delete 路径）

`syncer/syncer_test.go` 新增（新建文件，stub Provider + 真实 Resolver 解析 `localhost`）：
- `TestDryRun_EmptyConfig`：无 providers/无规则 → `Warnings` 两条、`Results` 为空数组
- `TestDryRun_Detail`：单域名规则 → 断言 `ToAdd`/`ToDelete` 明细数组长度与字段
- `TestDryRun_Concurrent`（可选）：并发调用第二个返回 `ErrDryRunInProgress`

**验收：**
```bash
go build ./...
go vet ./...
go test ./provider/... ./syncer/... -v
```

---

### Step 5：运行测试页（Dry Run + 连接测试双标签）

**目标：** 独立的「运行测试」页面，统一承载 Dry Run（三级分组视图）与连接测试（含 15s 超时）。

**前置条件：** Step 3、Step 4 完成

**产出文件与操作：**

#### 5.1 `webui/frontend/src/components/DryRunResults.vue`（新增，`components/` 为新建目录）—— 三级分组视图 + 空状态

结构（依据 Design2 §8.5 Tab 1）：
- Props：`results: DryRunResult[]`、`warnings: string[]`、`hasRun: boolean`
- 空状态优先级：
  1. `!hasRun` → "尚未执行 Dry Run"
  2. `warnings.length > 0` → 按 warnings 文案逐条展示（`NAlert type="warning"`）
  3. `results.length === 0` → "无待变更规则"
- 分组渲染（三级：目标 → 域名 → 规则）：
  - 按 `provider` 分组 → 目标卡片（`NCard`，标题 = provider 名称）
  - 组内按 `domain` 列表 → 每个域名一个区块（`h4`）
  - 每区块两张明细表（`NDataTable size="small"`）：待添加（`to_add`）/ 待删除（`to_delete`），列：协议/端口/动作/CIDR/描述
  - `result.error` 非空 → `NAlert type="error"` 错误行，不展开明细
- 统计条（内部计算）：目标数 / 待添加总数 / 待删除总数 / 错误数

#### 5.2 `webui/frontend/src/views/RunTest.vue`（新增）—— 页面骨架

```
NTabs（激活状态与路由 query 同步：?tab=dryrun|connection）
├── NTabPane "dryrun"（tab="Dry Run"）
│   ├── 操作区：NButton「执行 Dry Run」(:loading) + 上次执行时间 + 统计条
│   └── DryRunResults 组件（数据来自 useDryRun()）
│       成功：message.success('Dry Run 完成')；失败：message.error(`Dry Run 失败: ${e.message}`)
└── NTabPane "connection"（tab="连接测试"）
    ├── 表单：NSelect(cloudOptions) + NInput(region) + NInput(resource_id)
    ├── NButton「测试连接」(:loading="testLoading"，执行中禁用防连点)
    └── 结果区：data.success ? data.message : data.error（200+success:false 语义保留）
         超时：request('/api/test-connection', {...}, 15000) 抛"请求超时"→ "连接超时（15 秒）"
```

要点：连接测试表单**内嵌本文件**（约 40 行），不抽 composable（Design2 §8.5）；`cloudOptions` 从 `constants.ts` 引用；Tab 切换用 `useRoute`/`router.replace` 同步 `?tab=` query。

#### 5.3 `main.ts` / `App.vue` —— 路由与菜单

```typescript
// main.ts 新增
{ path: '/run-test', component: () => import('./views/RunTest.vue') },
```

```typescript
// App.vue 菜单新增（位于「同步日志」与「告警配置」之间）
{ label: '运行测试', key: '/run-test' },
```

**测试与验收：**
```bash
cd webui/frontend && npm run build
cd ../.. && go build ./... && go vet ./... && go test ./...
# 手动验证：启动 WebUI，访问 /#/run-test，执行 Dry Run 与连接测试
```

---

### Step 6：页面收敛（全站接入 `api.ts` + 删除高级功能页）

**目标：** 修复全站 5 处"失败误报成功"；Targets 弹窗测试连接保留并增强；Dashboard 移除试运行；删除 Advanced 页面。

**前置条件：** Step 3、Step 5 完成

**产出文件与操作：**

#### 6.1 `views/Settings.vue`
- `save()` 改走 `request('/api/settings', { method: 'PUT', body: ... })`，失败 `message.error(e.message)`（修复误报成功）
- 保留现有「导出配置」「导入配置」按钮（唯一入口，Design2 §8.6）；可选微调：按钮行下加"导入会覆盖当前配置"说明

#### 6.2 `views/Targets.vue`
- `load`/`saveTarget`/`deleteTarget`/`testConnection` 全部改走 `request()`
- `testConnection` 使用 `request(url, {...}, 15000)` 支持超时；`AbortError` 分支显示"连接超时（15 秒）"
- `cloudOptions`/`cloudLabelMap` 改引 `constants.ts`
- 弹窗内表单级「测试连接」**保留**（用未保存的表单值验证，Design2 §8.5 决策）

#### 6.3 `views/Rules.vue`
- `load`/`saveRule`/`deleteRule` 改走 `request()`（修复保存/删除失败误报成功）

#### 6.4 `views/Logs.vue`
- `load` 改走 `request()`（保持 SSE 逻辑不变）

#### 6.5 `views/Dashboard.vue`（B5 预留）
- 删除：`dryRun()`、`dryrunResults`、`showDryrun`、`dryrunColumns`、`NModal` 及相关 import
- 新增：「运行测试」次级链接（放「同步引擎」状态卡片内）
- **B5 预留**：操作区渲染数据驱动——`status.enabled === false` 时「立即同步」`disabled` + hover 提示"请先开启同步引擎"（后端未返回 `enabled` 时默认 true，行为不变）

#### 6.6 删除高级功能页
- 删除 `views/Advanced.vue`；`main.ts` 删除路由 `/advanced`；`App.vue` 菜单删除「高级功能」；全库检索引用并清理

**测试与验收：**
```bash
cd webui/frontend && npm run build
cd ../.. && go build ./... && go vet ./... && go test ./...
grep -rn "advanced" webui/frontend/src || echo "无残留引用"
```

---

### Step 7：数据层 —— `Config.SyncEnabled` 持久化

**目标：** `sync_enabled` 键（SQLite）与 `SYNC_ENABLED` 环境变量（.env）读写。

**前置条件：** 无（按顺序执行）

**产出文件与操作：**

#### 7.1 `config/config.go`

```go
type Config struct {
	// ... 现有字段不变
	SyncEnabled bool // 同步开关：true=开启，false=暂停；默认 true
}
```

#### 7.2 `config/store.go` —— `LoadConfig()` 新增解析

```go
cfg.SyncEnabled = true // 默认值（向后兼容：老用户无该键时保持启动即同步）
if v := settings["sync_enabled"]; v != "" {
	cfg.SyncEnabled = v == "true"
}
```

#### 7.3 `config/env.go` —— `ParseEnv()` 新增解析

```go
cfg.SyncEnabled = true // 默认值
if v := kv["SYNC_ENABLED"]; v != "" {
	cfg.SyncEnabled = v == "true"
}
```

#### 7.4 单元测试（`config/env_test.go`）

新增用例：`SYNC_ENABLED=true` → true；`=false` → false；未设置 → 默认 true；非法值（如 `abc`）→ false（`v == "true"` 语义）。

**验收：**
```bash
go build ./... && go vet ./...
go test ./config/... -v
```

---

### Step 8：Syncer 暂停门控（Pause Gate + 热重载开关同步）

**目标：** 暂停/恢复机制（不重启 goroutine），并保证热重载后 DB 状态与运行时镜像一致（Design2 §5.6）。

**前置条件：** Step 7 完成

**产出文件与操作：**

#### 8.1 `syncer/syncer.go` —— 结构扩展

```go
type Syncer struct {
	// ... 现有字段不变
	syncEnabled bool          // 运行时镜像，启动时从 cfg.SyncEnabled 初始化
	pauseCh     chan struct{} // 接收暂停信号，容量 1
	resumeCh    chan struct{} // 接收恢复信号，容量 1
}
```

`New()` 中初始化：`syncEnabled: cfg.SyncEnabled`、`pauseCh/resumeCh: make(chan struct{}, 1)`。

#### 8.2 `Run()` 主循环改造（启动门控 + 暂停子循环）

```go
func (s *Syncer) Run() {
	defer close(s.doneCh)
	s.setRunning(true)
	defer s.setRunning(false)

	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	// 启动门控：开关关闭时跳过首次 syncAll，进入暂停等待（Design2 §7.1）
	// 注意：waitForResume 返回后必须继续进入主循环，不可 return（否则恢复后定时同步失效）
	if !s.syncEnabled {
		slog.Info("同步已暂停（SyncEnabled=false），等待开启")
		s.waitForResume(ticker)
	} else {
		s.syncAll()
	}

	for {
		select {
		case <-ticker.C:
			s.syncAll()
		case <-s.triggerCh:
			slog.Info("手动触发同步")
			s.syncAll()
		case newCfg := <-s.configCh:
			slog.Info("配置热重载")
			s.mu.Lock()
			s.cfg = newCfg // 保持 Step 2 引入的锁结构
			s.mu.Unlock()
			ticker.Reset(newCfg.Interval)
			// 5.6 开关同步：热重载变更 sync_enabled 时同步门控
			if newCfg.SyncEnabled != s.syncEnabled {
				if newCfg.SyncEnabled {
					slog.Info("热重载开启同步")
					s.syncEnabled = true
					s.syncAll() // 立即执行首次
				} else {
					slog.Info("热重载暂停同步")
					s.syncEnabled = false
					s.pauseGate(ticker)
				}
			}
		case <-s.pauseCh:
			s.pauseGate(ticker)
		case <-s.stopCh:
			slog.Info("同步引擎停止")
			return
		}
	}
}

// pauseGate 暂停门控：停止 ticker，进入等待子循环（resume/热重载开启/stop 均返回后回到主循环）
func (s *Syncer) pauseGate(ticker *time.Ticker) {
	ticker.Stop()
	s.waitForResume(ticker)
}

// waitForResume 暂停等待子循环（Run 启动时暂停与运行中暂停共用）
// 返回条件：收到 resumeCh / 热重载开启（configCh 携带 SyncEnabled=true）→ 已恢复 ticker 并执行首次 syncAll；
// 收到 stopCh → 直接返回（外层 Run 退出）
func (s *Syncer) waitForResume(ticker *time.Ticker) {
	for {
		select {
		case <-s.resumeCh:
			slog.Info("同步恢复")
			s.syncEnabled = true
			ticker.Reset(s.cfg.Interval)
			s.syncAll() // 恢复后立即执行首次
			return
		case newCfg := <-s.configCh:
			s.mu.Lock()
			s.cfg = newCfg
			s.mu.Unlock()
			if newCfg.SyncEnabled {
				s.syncEnabled = true
				ticker.Reset(newCfg.Interval)
				s.syncAll()
				return
			}
			// SyncEnabled 仍为 false：继续等待
		case <-s.stopCh:
			return
		}
	}
}
```

> **设计说明（合理化完善）：** Design2 §5.2 原设计为"暂停子循环 + 主循环分置"；本 Step 整合为单一 `waitForResume()`（启动暂停与运行中暂停共用），并在暂停子循环内同步监听 `configCh`（热重载开启可立即恢复），实现 §5.6 开关同步；`Run()` 启动时若 `syncEnabled=false` 直接进入 `waitForResume()` 且不执行首次 syncAll()（对应 Design2 §7.1 "始终启动 Run()"决策，`.env` 模式 `doneCh` 正常关闭，`WaitForSignal` 无需特殊处理）。

#### 8.3 公开方法

```go
// Pause 暂停同步（非阻塞）
func (s *Syncer) Pause() {
	s.syncEnabled = false
	select {
	case s.pauseCh <- struct{}{}:
	default: // 已在暂停中
	}
}

// Resume 恢复同步（非阻塞）
func (s *Syncer) Resume() {
	s.syncEnabled = true
	select {
	case s.resumeCh <- struct{}{}:
	default: // 已在运行中
	}
}

// IsEnabled 返回当前开关状态
func (s *Syncer) IsEnabled() bool { return s.syncEnabled }
```

> **并发说明：** `Pause()`/`Resume()` 直接赋值 `syncEnabled`（信号由主循环串行消费），与 §5.5 的容量 1 channel + `select/default` 兼容，可接受（"不过度防御"）；若 `-race` 报告，可将 `syncEnabled` 改为 `atomic.Bool`。

#### 8.4 单元测试（`syncer/syncer_test.go` 新建）

stub Provider（实现 `provider.Provider` 接口，`GetRules` 返回空）+ 真实 `dns.Resolver`（解析 `localhost`）：
- `TestRun_DisabledStartup`：`SyncEnabled=false` 创建 → `Run()` 后不执行同步（stub 计数 0）→ `Resume()` 后执行一次 → `Stop()`
- `TestPauseResume_Flow`：正常运行 → `Pause()` → 计数停止 → `Resume()` → 恢复
- `TestReload_SyncEnabledSync`：`Reload(SyncEnabled=false)` → 门控生效；再 `Reload(SyncEnabled=true)` → 恢复
- `TestIsEnabled`：状态镜像正确

**验收：**
```bash
go build ./... && go vet ./...
go test ./syncer/... -v -race
```

---

### Step 9：API 层 —— pause/resume 端点 + trigger 保护 + SyncStatus.enabled

**目标：** 新增 `POST /api/sync/pause|resume`；暂停时 trigger 返回 409；`SyncStatus` 带 `enabled`。

**前置条件：** Step 7、Step 8 完成

**产出文件与操作：**

#### 9.1 `webui/api/deps.go` —— Syncer 接口扩展

```go
type Syncer interface {
	Status() syncer.SyncStatus
	TriggerSync()
	DryRun() (syncer.DryRunResponse, error)
	Pause()  // 新增
	Resume() // 新增
}
```

#### 9.2 `syncer/syncer.go` —— `SyncStatus` 扩展

```go
type SyncStatus struct {
	Running  bool       `json:"running"`
	Enabled  bool       `json:"enabled"`    // 新增
	LastSync *time.Time `json:"last_sync"`
}
```

`Status()` 返回时填充 `Enabled: s.IsEnabled()`。

#### 9.3 `webui/api/sync.go` —— 新增 handler + 路由 + trigger 保护

```go
func (d *Deps) handleSyncPause(w http.ResponseWriter, r *http.Request) {
	if d.Syncer == nil {
		writeError(w, http.StatusBadRequest, "同步引擎未启动")
		return
	}
	if err := d.Store.SetSetting("sync_enabled", "false"); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	d.Syncer.Pause() // 先写 DB 后通知 Syncer（Design2 §6.2）
	writeJSON(w, http.StatusOK, map[string]string{"message": "同步已暂停"})
}

func (d *Deps) handleSyncResume(w http.ResponseWriter, r *http.Request) {
	if d.Syncer == nil {
		writeError(w, http.StatusBadRequest, "同步引擎未启动")
		return
	}
	if err := d.Store.SetSetting("sync_enabled", "true"); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	d.Syncer.Resume()
	writeJSON(w, http.StatusOK, map[string]string{"message": "同步已恢复"})
}

// handleSyncTrigger 改造：暂停时返回 409
func (d *Deps) handleSyncTrigger(w http.ResponseWriter, r *http.Request) {
	if d.Syncer == nil {
		writeError(w, http.StatusBadRequest, "同步引擎未启动，请先配置目标和规则")
		return
	}
	if !d.Syncer.Status().Enabled {
		writeError(w, http.StatusConflict, "同步已暂停，请先开启")
		return
	}
	d.Syncer.TriggerSync()
	writeJSON(w, http.StatusAccepted, map[string]string{"message": "同步已触发"})
}
```

`deps.go` 路由注册新增：
```go
mux.HandleFunc("POST /api/sync/pause", d.handleSyncPause)
mux.HandleFunc("POST /api/sync/resume", d.handleSyncResume)
```

#### 9.4 单元测试（`webui/api/sync_test.go` 新建，推荐）

stub Syncer（实现接口）+ `httptest`：
- `TestHandleSyncTrigger_Paused`：stub `Status().Enabled=false` → 409 + `{"error":"同步已暂停，请先开启"}`
- `TestHandleSyncPause`：真实 Store（临时 SQLite）→ 200 且 settings 表写入 `sync_enabled=false`

**验收：**
```bash
go build ./... && go vet ./... && go test ./...
# 手动：curl -X POST /api/sync/pause 与 /api/sync/resume 验证响应
```

---

### Step 10：启动决策统一（始终启动 `Run()`）

**目标：** 确认/落地 Design2 §7.1/§7.2 决策——`main.go` 与 `app.go` **不做条件分支**，`go s.Run()` 保持原样，启动门控由 Step 8 的 `Run()` 内部完成。

**前置条件：** Step 8 完成

**产出文件与操作：**

- `main.go`：`go s.Run()` 保持不变（零改动或仅注释说明：启动门控在 `Run()` 内，见 Step 8.2）
- `app/app.go`：`go s.Run()` 保持不变（`.env` 模式同样受益：`SYNC_ENABLED=false` 时 `Run()` 存活于 `waitForResume()`，`doneCh` 正常关闭，`WaitForSignal()` 无需适配）
- 验证 `main.go` WebUI 流程：`LoadConfig()`（含 `SyncEnabled`）→ `syncer.New(cfg, ...)`（镜像初始化）→ `srv.SetSyncer(s, ...)` → `go s.Run()`

**测试与验收：**
```bash
go build ./... && go vet ./... && go test ./...
# 双模式运行验证：
# 1) FWALIZER_MODE=webui ./fwalizer → 正常启动
# 2) 临时 .env 加 SYNC_ENABLED=false → 日志出现"同步已暂停"，Ctrl+C 正常退出（无死锁）
# 3) SYNC_ENABLED=true → 启动后立即执行首次 syncAll
```

---

### Step 11：Dashboard 同步开关 + 按钮状态联动

**目标：** 仪表盘暂停/开启开关、三态状态标签、「立即同步」联动（Design2 §8.1-8.3，复用 Step 6 的 B5 预留）。

**前置条件：** Step 9、Step 10 完成

**产出文件与操作：**

#### `views/Dashboard.vue` 改造

```typescript
const enabled = ref<boolean>(true)          // 默认 true（后端未返回 enabled 时兼容）
const switching = ref(false)                // 开关请求 loading

// fetchStatus 更新：status.enabled !== undefined 时同步 enabled
async function toggleSync() {
  switching.value = true
  try {
    await request(`/api/sync/${enabled.value ? 'pause' : 'resume'}`, { method: 'POST' })
    enabled.value = !enabled.value
    message.success(enabled.value ? '同步已开启' : '同步已暂停')
  } catch (e: any) {
    message.error(`操作失败: ${e.message}`)
  } finally {
    switching.value = false
    fetchStatus()
  }
}
```

UI 变更：
- 「同步引擎」状态标签三态（Design2 §8.2）：`enabled === true` → 绿色「运行中」；`enabled === false` → 橙色「已暂停」（`running` 保持引擎存活语义）
- 操作卡片：新增 `NButton`「暂停」/「开启」（`:loading="switching"`）；「立即同步」在 `enabled === false` 时 `disabled` + tooltip"请先开启同步引擎"
- 保留「运行测试」链接（Step 6 已加）

`types.ts`：`SyncStatus.enabled` 由可选升为必填（后端已返回；如需兼容旧后端可保留 `?`）

**测试与验收：**
```bash
cd webui/frontend && npm run build
cd ../.. && go build ./... && go vet ./... && go test ./...
# 手动验证：暂停→标签变「已暂停」+ 立即同步置灰；暂停下 trigger 返回 409；开启→立即同步一次；重启后状态保持
```

---

### Step 12：文档同步

**目标：** 全部代码完成后，同步相关文档，保证文档-代码一致性。

**前置条件：** Step 4-11 全部验收通过

**产出文件与操作：**

| 文件 | 修改内容 |
|------|---------|
| `Design1.md` | **特性表**（L274「Dry Run \| 执行到 Diff 为止，不实际写入」）措辞对齐：改为"返回 `to_add`/`to_delete` 规则明细列表（含协议/端口/动作/CIDR/描述），响应为 `{results, warnings}` 包装"；**页面列表**（L195「高级功能（Dry Run、配置导入/导出、健康检查）」）同步：高级功能页已删除，Dry Run/连接测试归入「运行测试」页，配置导入/导出归入「全局设置」页（注：Design1.md 无 12.9 章节，Dry Run 描述实际位于特性表与页面列表两处） |
| `AGENTS.md` §十一（代码规范） | 补充：`POST /api/sync/pause|resume` 端点、`/run-test` 运行测试页、`SYNC_ENABLED`/`sync_enabled` 开关、`SyncStatus.enabled` 字段；§五 同步调度补充"暂停时 ticker/trigger 均不触发，Dry Run 与连接测试不受影响" |
| `README.md` | WebUI 页面列表增加「运行测试」、移除「高级功能」描述（如提及）；功能特性增加"同步全局开关（可暂停/恢复）"；.env 配置表增加 `SYNC_ENABLED` |
| `.env.example` | 可选配置区增加注释行：`# SYNC_ENABLED=true           # 同步开关（默认 true；false 时启动不执行同步，Dry Run 不受影响）` |
| `docker-compose.yml.example` | 如含环境变量透传列表，补充 `SYNC_ENABLED`（默认不设置即可） |
| `Issue3.md` | R16-01~06 修复完成后，将 §六 状态更新为 ✅（含修复证据） |

**测试与验收：**
```bash
grep -n "run-test\|运行测试" README.md Design1.md AGENTS.md || echo "需确认新页面已写入文档"
grep -n "SYNC_ENABLED" .env.example README.md || echo "需确认开关已写入文档"
```

---

### Step 13：低优先可选附项（Design2 §12.4，可独立执行）

> 每项相互独立，不阻塞主链路验收；按需选择执行。

#### 附项 A：`INTERVAL` 解析校验统一
- **现状**：`env.go` 解析失败返回错误；`store.go` `LoadConfig` 静默 fallback 5m
- **方案**：`LoadConfig` 中 `interval` 解析失败时 `slog.Warn` 并保留默认值（不打断启动）；`Settings.vue` interval 输入加 `placeholder="5m"` 与保存前格式校验（正则 `^\d+(ms|s|m|h)$`）
- **验收**：`go build ./...`；设置页输入非法间隔保存时提示错误

#### 附项 B：未知 API 路径返回 HTML（验证项，无代码改动）
- **现状**：`server.go` 的 `Handle("/")` 兜底，未知 `/api/*` 返回 SPA
- **结论**：Step 3 的 `api.ts` 已天然免疫（`res.json()` 失败进 catch），无需改动
- **验收**：`curl http://127.0.0.1:60200/api/not-exist` 返回 HTML 但不影响前端行为

#### 附项 C：热重载 Provider 重建失败提示
- **现状**：`main.go` ReloadFunc 中 `NewProvider` 失败 `continue`，仅 slog.Error
- **方案**：ReloadFunc 收集失败目标列表，失败数量 > 0 时 `slog.Error("部分目标重建失败", "failed", n)`
- **验收**：`go build ./...`；构造非法目标后触发热重载，日志含失败汇总

#### 附项 D：`Settings.vue` 补 `dns_fail_threshold` 输入框
- **现状**：后端 `LoadConfig` 已支持该 key，前端表单缺失
- **方案**：`settings.go` 的 defaults map 增加 `"dns_fail_threshold": "5"`；`Settings.vue` 在 `dns_timeout` 下新增输入框
- **验收**：`npm run build`；设置页可读写该配置

---

## 五、关键约束与验收汇总

**编码约束（全程适用）：**
- 遵循 [AGENTS.md](./AGENTS.md)：不引入外部框架（HTTP 继续用 `net/http`）、不引入 cron 库、中文注释、`log/slog`、所有 error 必须处理
- **防火墙规则操作约束不变**：绝不使用全量覆盖类 API；仅增量添加 + 精确删除；删除"已不存在"视为成功、添加"已存在" WARN 跳过
- 前端不新增重型状态管理库（沿用 Vue 3 `ref`/`composable` 即可）

**每步通用验收（随各 Step 附带的具体命令之外）：**
```bash
go build ./... && go vet ./... && go test ./...
cd webui/frontend && npm ci && npm run build && cd ../..
```

**全量回归（所有 Step 完成后）：**
```bash
go build ./... && go vet ./... && go test ./... -race
cd webui/frontend && npm ci && npm run build && cd ../..
CGO_ENABLED=0 go build -tags docker -ldflags="-s -w -X github.com/alcaprophet/fwalizer/version.Version=dev" .
```

**验收条目 ↔ Step 对应关系：**

| 验收条目 | 对应 Build3 Step |
|------------------------|------------------|
| Issue3 R16-01（ICMP 归一化） | Step 1 |
| Issue3 R16-02（CVM PolicyIndex） | Step 1 |
| Issue3 R16-03（日志流级别） | Step 2 |
| Issue3 R16-04（cfg 竞态） | Step 2（+Step 4/8 锁配对） |
| Issue3 R16-05（日志清理） | Step 2 |
| Issue3 R16-06（死代码） | Step 2 |
| Design2 §12.3 1-8（全局开关 8 条） | Step 7、8、9、10、11 |
| Design2 §12.3 9-14（Dry Run 明细/分组/空状态/错误/防连点/限速） | Step 4、5 |
| Design2 §12.3 15（入口合并） | Step 5、6 |
| Design2 §12.3 16-18（连接测试） | Step 4、5、6 |
| Design2 §12.3 19（页面收敛） | Step 6 |
| Design2 §12.3 20（开关状态一致） | Step 8（5.6 实现） |
| Design2 §12.3 21（统一封装） | Step 3、6 |

---

## 六、风险提示（执行时注意）

1. **Step 1（ICMP 归一化）涉及 Diff 行为变更**：修改 `keyOf`/`keyOfAction` 会影响所有协议的比较路径——必须先行运行现有 `provider` 全部测试（`TestDiff_*` 系列）确认非 ICMP 路径零回归，再补 ICMP 新用例
2. **Step 2 与 Step 4/8 的锁配对**：cfg 读写锁（写锁 Step 2 引入、读锁 Step 4 快照、Step 8 主循环改造）贯穿三步，每步改动后运行 `-race`
3. **Step 8 是核心复杂度所在**：`Run()` 主循环重构涉及并发路径，务必测试先行并运行 `-race`
4. **Step 4 的 `DryRun()` 签名变更**影响 `deps.go` 接口与 Step 3 的前端契约，前后端必须同一轮完成，避免中间态
5. **Step 6 删除 Advanced 页前**先 `grep -rn advanced` 确认无残留跳转引用
6. **Step 11 的 `enabled` 语义**：`running` 保持"引擎存活"语义，前端三态判断以 `enabled` 为准（Design2 §8.2 说明）
7. **Step 13 附项 C** 涉及 ReloadFunc（`main.go`），改动时注意告警订阅重建顺序不受影响
