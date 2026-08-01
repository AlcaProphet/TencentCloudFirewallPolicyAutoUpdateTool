# Design2.md — 同步全局开关与运行测试页设计

> 本文档描述「同步全局开关」「运行测试页（Dry Run + 连接测试）」「页面收敛（高级功能页拆分）」的设计构想，涵盖需求分析、存储方案、核心机制、API 设计、WebUI 交互、兼容性分析及风险评估。
>
> **文档定位：** 本文档属于**设计构想（非强制规定）**。编码约束遵循 [AGENTS.md](./AGENTS.md)（项目唯一强要求文档）；与 AGENTS.md 或用户决策冲突时，以用户确认为准；详细构建方案见 [Build3.md](./Build3.md)。
>
> **版本说明（本版整合）**：在原「同步全局开关」设计基础上，整合 Dry Run 与连接测试功能升级方案——原分布于仪表盘（`Dashboard.vue` 试运行按钮）、高级功能（`Advanced.vue` Dry Run 标签页、连接测试标签页、配置导入导出标签页）的功能统一收敛：
>
> - **运行测试页（`/run-test`）**：合并仪表盘「试运行」与高级功能「Dry Run」「连接测试」为单一页面，内部以 `NTabs` 拆分「Dry Run」「连接测试」两个子标签页；
> - **Dry Run 明细化**：`POST /api/sync/dryrun` 响应结构破坏性变更——`to_add`/`to_delete` 由计数 `int` 改为 `RuleChange[]` 数组（方案 A），一次请求返回全部明细，响应包装为 `{results, warnings}` 支持语义化空状态；展示采用「目标 → 域名 → 规则」三级分组视图；
> - **连接测试增强**：前端 15s 超时（`AbortController`）+ 后端凭据空值快速失败提示；
> - **页面收敛**：高级功能页（`/advanced`）三个标签页全部迁出——Dry Run 与连接测试 → 运行测试页，配置导入导出 → 全局设置页（设置页已内置同功能按钮，删除高级功能页重复实现后以设置页为准），**高级功能页删除**；
> - 同时纳入前端共享逻辑抽离、交互完善（loading/错误处理/防连点）、Dry Run 并发安全（限速、防重入、快照锁）等改进要点，并为全局开关的仪表盘联动预留（`SyncStatus.enabled` 可选字段）。
>
> **深度审查整合（本版补充）**：基于对现有代码的全量检查（启动链路/配置层/API 层/Provider 层/事件总线/全部前端页面），补充四项结论：① 云资源管理页弹窗内的表单级「测试连接」**保留**（第三处调用点，与运行测试页定位互补，见 8.5）；② 热重载需同步开关状态（配置导入/设置写入 `sync_enabled` 后 Syncer 运行时镜像保持一致，见 5.6）；③ 新增前端统一 fetch 封装 `api.ts`（修复全站 5 处"失败误报成功"，见 8.5/12.1）；④ 连接测试 15s 超时增强覆盖全部调用点（含 `Targets.vue` 弹窗）。

---

## 一、概述

### 1.1 功能定义

**功能一：同步全局开关**

为 FWAlizer 增加一个**可持久化状态的同步全局开关**，允许用户按需开启或暂停自动同步。开关状态持久化到 SQLite（WebUI 模式）或环境变量（`.env` 模式），应用重启后自动恢复。

**功能二：运行测试页**

将 Dry Run（试运行）与连接测试从仪表盘、高级功能页中独立出来，作为**统一入口的「运行测试」页面**（路由 `/run-test`，菜单「运行测试」），定位为正式的测试诊断工具，页内以 `NTabs` 提供两个子标签页：

- **Tab 1：Dry Run** —— 结果**逐条规则级明细**（返回每个目标/域名待添加、待删除的具体规则：协议、端口、动作、CIDR、描述），而非仅计数；**一次请求返回全部明细**，不设独立明细端点；展示采用**「目标 → 域名 → 规则」三级分组视图**；空状态语义化（无目标、无规则时返回可解释的 warnings）；暂停状态下始终可用（与全局开关解耦，见 9.1）。
- **Tab 2：连接测试** —— 由高级功能页迁入，保留现有表单形态与错误语义（业务失败返回 HTTP 200 + `success:false`）；新增 15s 前端超时与后端凭据空值快速失败提示；与 Dry Run 一致，**暂停状态下可用**（端点不依赖 Syncer）。云资源管理页弹窗内的表单级「测试连接」（用未保存的表单值验证凭据/资源 ID）**保留**——运行测试页是常驻诊断入口而非唯一入口（见 8.5）。

**功能三：页面收敛（高级功能页拆分）**

高级功能页（`/advanced`）原有三个标签页全部迁出后删除：

| 原标签页 | 去向 |
|---------|------|
| Dry Run | → 运行测试页 Tab 1 |
| 连接测试 | → 运行测试页 Tab 2（云资源管理弹窗内的表单级测试连接**保留**，见 8.5） |
| 配置导入/导出 | → 全局设置页（设置页表单末尾保存行**已内置**「导出配置」「导入配置」按钮，见 8.6） |

### 1.2 核心需求

| 需求 | 说明 |
|------|------|
| **按需启动** | 开关关闭时，Syncer 不执行任何同步（包括定时和手动触发）；Dry Run 不受影响 |
| **状态持久化** | 关闭后重启应用，同步保持暂停；开启后重启，自动恢复同步 |
| **优雅暂停** | 暂停时若正在执行同步，等待当前轮次完成后再进入暂停状态 |
| **即时恢复** | 开启后立即恢复定时同步（等待下一个 ticker 周期，或用户手动触发） |
| **双模式兼容** | WebUI 模式通过 SQLite + API 控制；`.env` 模式通过环境变量控制 |
| **运行测试独立页** | 合并仪表盘「试运行」、高级功能「Dry Run」「连接测试」为统一「运行测试」页（双子标签）；云资源管理弹窗内的表单级测试连接**保留** |
| **明细化预览** | Dry Run 返回逐条规则级明细（`to_add`/`to_delete` 为规则数组），支撑「先验证再执行」闭环 |
| **暂停可用** | 暂停状态下 Dry Run 与连接测试均始终可用（均不依赖 Syncer 主循环） |
| **交互完备** | 运行测试页具备 loading、错误提示、防连点、空状态、超时提示等完整交互 |
| **页面收敛** | 高级功能页删除；配置导入导出由全局设置页统一承载 |

### 1.3 设计动机

当前 FWAlizer 启动后 `Syncer.Run()` 立即执行 `syncAll()`（[syncer.go](syncer/syncer.go#L67)），无任何缓冲机制。用户首次配置或维护时缺乏「先验证再执行」的安全窗口。虽然 Dry Run 提供了预览能力，但缺乏在源头阻止写入的手段。（此为全局开关的动机，保留原文）

Dry Run 与测试功能的升级动机（本次整合）：

1. **三处入口分散且体验不对等**：仪表盘「试运行」`dryRun()`、高级功能「Dry Run」`runDryRun()` 调用同一端点 `POST /api/sync/dryrun`、同一 handler、底层复用同一份 `Syncer.DryRun()`，差异仅在前端展示层；仪表盘版无 loading、无错误处理，高级功能版不检查 `res.ok`（后端 400/500 的 `{"error":...}` 会被当结果渲染，且误报成功）。连接测试位于高级功能页第三个标签，与 Dry Run 同属"测试"心智模型，分散在不同入口。
2. **结果仅计数、明细被丢弃**：`provider.Diff` 已算出完整明细（`DiffResult{ToAdd []RuleAction, ToDelete []RuleInfo}`），`Syncer.DryRun()` 只取 `len()`，用户只能看到"加 3 条、删 1 条"，无法确认具体规则；与历史构建文档 [Build1.md](./Build1.md) §12.9「返回 JSON：每个域名/目标的 toAdd 和 toDelete 规则列表」的承诺不一致。
3. **Dry Run 无限速、无防重入、有竞态**：`syncAll()` 每域名间有 `rateLimitInterval` 限速，`DryRun()` 循环无限速（频繁点击会冲击云 API 配额）；无防重入（可并发多次执行）；遍历 `s.providers`/`s.cfg` 无读锁，与热重载并发存在数据竞态。
4. **Dry Run 空状态无语义**：无目标或无规则时返回 `[]`，用户无法区分"没配置"与"无变更"。
5. **配置导入导出重复实现**：`Settings.vue` 保存行与 `Advanced.vue` 配置标签页均实现了「导出配置」「导入配置」（同一对端点），重复维护。
6. **连接测试边界缺陷**：`GET/POST` 云 API 无超时控制（云厂商 API 卡住时前端 loading 无限转圈）；凭据未配置时直接暴露 SDK 原始报错，提示不友好。

---

## 二、现有架构分析

### 2.1 两种运行模式的 Syncer 生命周期

| 模式 | 启动流程 | Syncer 启动 | 可重启 |
|------|---------|-------------|--------|
| `.env` | [app/app.go](app/app.go#L44-L45) → `go s.Run()` | 启动即 `syncAll()` | 否 |
| WebUI | [main.go](main.go#L199) → `go s.Run()` | 启动即 `syncAll()` | 否 |

两种模式的 `Syncer.Run()` 都是**单一 goroutine、一次性运行**（[syncer.go](syncer/syncer.go#L57-L85)）：

```
setRunning(true)
ticker := NewTicker(interval)
syncAll()                    ← 启动即执行，无前置检查
for {
  select {
    case ticker.C    → syncAll()
    case triggerCh   → syncAll()
    case configCh    → 更新配置 + 重置 ticker
    case stopCh      → return (永久退出，stopCh 已关闭，不可重启)
  }
}
```

核心约束：`Stop()` 调用 `close(stopCh)` 是终态的，一个 Syncer 实例只能运行一次。

### 2.2 现有对外接口与缺陷

[webui/api/deps.go](webui/api/deps.go#L12-L17) 暴露的 Syncer 接口：

```go
type Syncer interface {
    Status() syncer.SyncStatus
    TriggerSync()
    DryRun() ([]syncer.DryRunResult, error)
}
```

当前无 `Pause()` / `Resume()` / `SetEnabled()` 方法。

**DryRun 现状缺陷清单**（升级设计依据）：

| 缺陷 | 位置 | 说明 |
|------|------|------|
| 明细丢弃 | [syncer.go](syncer/syncer.go#L183-L184) | `result.ToAdd = len(diff.ToAdd)`，仅保留计数 |
| 无限速 | [syncer.go](syncer/syncer.go#L156-L189) | 循环内无 `rateLimitInterval`（对比 `syncAll()` 的 L214） |
| 无防重入 | `DryRun()` | 可并发多次执行，重复消耗云 API 配额 |
| 数据竞态 | `DryRun()` 遍历 | 读取 `s.providers`/`s.cfg` 无读锁，与 `ReloadProviders()`/`Reload()` 并发不安全 |
| 空状态无语义 | 返回 `[]` | 无目标/无规则时用户无法区分原因 |

**连接测试现状**（[targets.go](webui/api/targets.go#L78-L120)）：

| 维度 | 现状 |
|------|------|
| 端点 | `POST /api/test-connection`，请求体 `{cloud_type, region, resource_id}` |
| 流程 | 解析请求 → `Store.GetSettings()` 读凭据 → `provider.SetCredentials()` → 临时 `NewProvider()` → `GetRules()` → 成功 `{success:true, message:"连接成功，当前 N 条规则"}` |
| 错误语义 | **业务失败返回 HTTP 200 + `{success:false, error}`**（与 Dry Run 的 4xx/5xx + `{error}` 语义不同，迁移时保留各自语义） |
| 依赖 | 不依赖 `d.Syncer`，与暂停状态无关（暂停时可用） |
| 调用点 | 现存两处：`Advanced.vue` 标签页、`Targets.vue` 添加/编辑弹窗（表单级：用未保存的表单值验证，**保留**）；迁入后仍为两处：`RunTest.vue`（常驻诊断入口）+ `Targets.vue` 弹窗 |
| 缺陷 | 无超时控制（云 API 卡住时前端 loading 无限转圈）；凭据未配置时直接暴露 SDK 原始报错；两处调用点均无统一 loading/超时（`Targets.vue` 弹窗仅有文本"测试中..."） |

**配置导入导出现状**：`Settings.vue`（保存行按钮）与 `Advanced.vue`（配置标签页）重复实现同一对端点（`GET /api/config/export`、`POST /api/config/import`）；`Settings.vue` 版导入成功后会刷新设置表单（[Settings.vue](webui/frontend/src/views/Settings.vue#L41-L42)），`Advanced.vue` 版不刷新——**以设置页实现为准，删除高级功能页重复实现**。

**前端 fetch 模式现状（深度审查发现）**：全部 7 个页面裸 `fetch`，仅 `Alerts.vue` 检查 `res.ok`；`Targets.vue`（保存/删除）、`Rules.vue`（保存/删除）、`Settings.vue`（save）在请求失败时仍执行 `message.success`——**后端报错但前端提示成功**（共 5 处误报）。统一封装方案见 8.5 前端实现拆分与 12.1。

### 2.3 热重载链路

WebUI 增删改操作 → `notifyReload()` → [main.go](main.go#L127-L181) ReloadFunc → `s.Reload(cfg)` → `configCh` → Run() 中仅更新配置和重置 ticker，**不触发 syncAll()**。热重载本身是安全的。

---

## 三、持久化存储方案

### 3.1 WebUI 模式：复用 `settings` 表

使用现有 `settings` 表（[store.go](config/store.go#L104-L107)）的 key-value 结构，新增键 `sync_enabled`：

```
settings:
  key           | value
  --------------|-------
  tag           | auto-dns
  interval      | 5m
  sync_enabled  | true      ← 新增
```

**优势**：
- 零 schema 变更，`INSERT OR REPLACE` 语义天然支持
- 已有 `GetSettings()` / `SetSetting()` 直接复用
- [LoadConfig()](config/store.go#L472-L535) 读取 pattern 成熟，仅需新增一行解析
- [Settings.vue](webui/frontend/src/views/Settings.vue) 的 `handlePutSettings` 批量保存机制已覆盖

**向后兼容**：`sync_enabled` 键不存在时（老用户升级），默认视为 `"true"`（保持启动即同步的现有行为）。

**配置导入/导出**：[settings.go](webui/api/settings.go#L51-L91) 的导入导出逻辑遍历 settings map，新增 key 自动覆盖，无需特殊处理。

### 3.2 `.env` 模式：环境变量

通过 `SYNC_ENABLED` 环境变量控制（默认 `true`）。在 [env.go](config/env.go#L28-L89) 的 `ParseEnv()` 中解析，存入 `Config` 结构体。

---

## 四、`Config` 结构体扩展

在 [config/config.go](config/config.go#L75-L91) 的 `Config` 结构体中新增字段：

```go
type Config struct {
    // ... 现有字段不变
    SyncEnabled bool  // 同步开关：true=开启，false=暂停；默认 true
}
```

### 4.1 WebUI 模式读取

[store.go](config/store.go) 的 `LoadConfig()` 中：

```go
cfg.SyncEnabled = true  // 默认值
if v := settings["sync_enabled"]; v != "" {
    cfg.SyncEnabled = v == "true"
}
```

### 4.2 `.env` 模式读取

[env.go](config/env.go) 的 `ParseEnv()` 中：

```go
cfg.SyncEnabled = true  // 默认值
if v := kv["SYNC_ENABLED"]; v != "" {
    cfg.SyncEnabled = v == "true"
}
```

---

## 五、Syncer 暂停/恢复机制设计

### 5.1 设计原则

1. **不重启 Syncer goroutine**：避免重建 Syncer 导致的连锁影响（EventBus 订阅丢失、LogWriter 重新绑定等）
2. **暂停仅阻止新轮次启动**：不中断正在执行的同步轮次，等待其自然完成
3. **最小侵入**：尽量复用现有 channel 机制

### 5.2 核心方案：内部暂停门控（Pause Gate）

#### 数据结构扩展

```go
type Syncer struct {
    // ... 现有字段不变
    syncEnabled bool          // 运行时镜像，启动时从 cfg.SyncEnabled 初始化
    pauseCh     chan struct{} // 接收暂停信号，容量 1
    resumeCh    chan struct{} // 接收恢复信号，容量 1
}
```

#### Run() 主循环改造

```
启动时:
  if s.syncEnabled → syncAll()         // 现有行为
  else → 跳过 syncAll()，进入主循环但暂停 ticker

主循环新增 case:
  case <-s.pauseCh:
    ticker.Stop()
    进入 waitForResume() 子循环
    子循环监听:
      resumeCh  → ticker.Reset() → 回到正常循环
      configCh  → 更新 cfg（不触发同步，ticker 已停）
      stopCh    → return（应用退出）

  case <-s.resumeCh (仅在暂停子循环中):
    ticker.Reset(cfg.Interval)
    回到正常主循环
```

#### 暂停期间行为

| 信号 | 正常运行时 | 暂停时 |
|------|-----------|--------|
| `ticker.C` | 执行 `syncAll()` | 不触发（ticker 已 Stop） |
| `triggerCh` | 执行 `syncAll()` | 忽略（或 API 层拒绝） |
| `configCh` | 更新配置 + Reset ticker | 仅更新配置 |
| `stopCh` | 优雅退出 | 优雅退出 |
| `DryRun()` | 正常 | **正常**（独立于主循环） |
| `POST /api/test-connection` | 正常 | **正常**（不依赖 Syncer） |

### 5.3 公开方法

```go
// Pause 暂停同步（非阻塞）
func (s *Syncer) Pause() {
    s.syncEnabled = false
    select {
    case s.pauseCh <- struct{}{}:
    default:  // 已在暂停中
    }
}

// Resume 恢复同步（非阻塞）
func (s *Syncer) Resume() {
    s.syncEnabled = true
    select {
    case s.resumeCh <- struct{}{}:
    default:  // 已在运行中
    }
}

// IsEnabled 返回当前开关状态
func (s *Syncer) IsEnabled() bool {
    return s.syncEnabled
}
```

### 5.4 正在执行的同步轮次处理

**设计决策：不在中途中断**。理由：

1. `syncAll()` 内部按云厂商并行，使用 `sync.WaitGroup` 等待所有 goroutine（[syncer.go](syncer/syncer.go#L204-L219)）
2. 每个 `syncDomain()` 最多 3 次重试（[retry.go](syncer/retry.go#L18-L75)）
3. 中断需要 context 传播 + 每个 Provider API 调用支持取消，改动范围巨大
4. "等待当前轮次完成"与项目已有优雅退出策略一致
5. 单域名同步通常几十秒内完成，用户感知延迟可接受

用户点击「暂停」后，若当前有同步正在执行，WebUI 应显示"等待当前轮次完成..."状态。

### 5.5 信号发送的竞态处理

快速点击暂停→恢复→暂停可能造成信号堆积。使用容量为 1 的 channel + `select/default` 发送可避免阻塞：

```go
select {
case s.pauseCh <- struct{}{}:
default: // 已有暂停信号在 channel 中，跳过
}
```

前端在 API 请求返回前禁用按钮（loading 状态），进一步降低竞态概率。

### 5.6 热重载与开关状态同步（深度审查整合新增）

**问题**：配置导入（`handleConfigImport` 的 `ClearAllTx` 清空重建 settings 表，`sync_enabled` 随导出文件携带）或 `PUT /api/settings` 直接写入 `sync_enabled` 后，经 `notifyReload()` → `LoadConfig()` → `Reload(newCfg)` **仅更新 `cfg` 字段，`syncEnabled` 运行时镜像不更新**——DB 状态与 Syncer 实际行为不一致（例如导入 `sync_enabled=false` 的配置后，定时同步仍照跑）。

**约定**：热重载时对比新旧 `cfg.SyncEnabled`，变化则同步调用 `Pause()`/`Resume()`（二者幂等，与 5.5 的 `select/default` 发送兼容）；或更简单地在 `Reload()` 内直接镜像 `s.syncEnabled = newCfg.SyncEnabled`（镜像赋值需在 `s.mu` 保护下进行，避免与 `Pause()`/`Resume()` 并发写 `syncEnabled`）。保证「DB → LoadConfig → cfg → 运行时镜像 → 主循环门控」五者一致。

---

## 六、API 设计

### 6.1 新增端点

| 方法 | 路径 | 功能 | 请求体 | 响应 |
|------|------|------|--------|------|
| `POST` | `/api/sync/pause` | 暂停同步 | 无 | `{"message":"同步已暂停"}` |
| `POST` | `/api/sync/resume` | 恢复同步 | 无 | `{"message":"同步已恢复"}` |

### 6.2 处理流程

```
POST /api/sync/pause
  → 1. Store.SetSetting("sync_enabled", "false")
  → 2. Syncer.Pause()          // 非阻塞
  → 3. 返回 200 {"message":"同步已暂停"}

POST /api/sync/resume
  → 1. Store.SetSetting("sync_enabled", "true")
  → 2. Syncer.Resume()         // 非阻塞
  → 3. 返回 200 {"message":"同步已恢复"}
```

**先写 DB 后通知 Syncer**：即使通知失败（如 channel 满），持久化状态已正确写入，重启后会自动读取到正确值。

### 6.3 现有端点改造

| 端点 | 改动 |
|------|------|
| `GET /api/sync/status` | `SyncStatus` 新增 `enabled bool` 字段（后端落地后返回；前端 `types.ts` 以可选字段 `enabled?: boolean` 兼容过渡，未返回时默认 true——B5 预留） |
| `POST /api/sync/trigger` | 暂停时返回 HTTP `409 Conflict` + `{"error":"同步已暂停，请先开启"}` |
| `POST /api/sync/dryrun` | **响应结构升级**（明细化 + 包装对象，详见 6.6）；执行中重复请求返回 `409 {"error":"Dry Run 正在执行中"}`（防重入）；**不检查暂停状态，始终可用** |
| `POST /api/test-connection` | **端点与错误语义不变**（200 + `success:false`）；新增凭据空值快速失败提示（详见 6.7） |
| `GET /api/config/export`、`POST /api/config/import` | **无需改动**（前端入口收敛到全局设置页） |
| `PUT /api/settings` | **无需改动**；`sync_enabled` 被写入后经 `notifyReload()` 走 5.6 开关同步 |

### 6.4 Syncer 接口扩展

[webui/api/deps.go](webui/api/deps.go#L12-L17) 需新增：

```go
type Syncer interface {
    Status() syncer.SyncStatus
    TriggerSync()
    DryRun() ([]syncer.DryRunResult, error)
    Pause()     // 新增
    Resume()    // 新增
}
```

### 6.5 `SyncStatus` 扩展

[syncer/syncer.go](syncer/syncer.go#L122-L127)：

```go
type SyncStatus struct {
    Running  bool       `json:"running"`
    Enabled  bool       `json:"enabled"`    // 新增
    LastSync *time.Time `json:"last_sync"`
}
```

### 6.6 Dry Run API 响应结构设计（升级）

#### 响应包装：空状态语义化（推荐方案，已采纳）

响应从裸数组改为包装对象：

```json
{
  "results": [ ... ],
  "warnings": [ "暂无云资源目标" ]
}
```

- `results`：明细数组（见下）；
- `warnings`：非致命提示数组——`providers` 为空时输出 `"暂无云资源目标，请先在云资源管理页配置"`；`DomainRules` 为空时输出 `"暂无域名规则，请先在域名规则页配置"`；二者都为空时输出两条；
- 单个目标/域名的失败仍落在对应 `result.error` 字段（`omitempty`），不进入 warnings。

#### `DryRunResult` 明细化（方案 A：字段类型变更）

**决策：`to_add`/`to_delete` 由 `number` 直接改为 `RuleChange[]`（破坏性变更）。项目仍处于构建期，不考虑旧版兼容，前端消费方本次统一改造。**

```go
// RuleChange 规则变更摘要（供前端直接渲染）
type RuleChange struct {
    Protocol string `json:"protocol"`
    Port     string `json:"port"`
    Action   string `json:"action"`
    Cidr     string `json:"cidr"` // IPv4 或 IPv6 的 CIDR（如 1.2.3.4/32）
    Desc     string `json:"desc"` // 规则描述（含 [TAG]）
}

// DryRunResult 试运行结果（升级后）
type DryRunResult struct {
    Provider string       `json:"provider"`
    Domain   string       `json:"domain"`
    ToAdd    []RuleChange `json:"to_add"`    // 待添加规则明细
    ToDelete []RuleChange `json:"to_delete"` // 待删除规则明细
    Error    string       `json:"error,omitempty"`
}
```

#### 服务端实现要点

- `provider/common.go` 新增 `RuleChange` 摘要构造函数（从 `config.RuleAction`/`config.RuleInfo` 转换，`Cidr` 取 `CidrBlock`/`Ipv6CidrBlock` 非空者）；
- `syncer.DryRun()` 填充数组（`diff.ToAdd` → `ToAdd`，`diff.ToDelete` → `ToDelete`），计数由前端 `length` 计算；
- **一次返回全部明细**：不新增独立明细端点（单台服务器少量 IP 场景下明细量小；仅将来支持多目标大量规则时再考虑拆分）；
- `handleSyncDryRun` 在 Syncer 为 nil 时返回 `400`，执行失败返回 `500`，执行中返回 `409`。

### 6.7 连接测试端点增强（本次整合新增）

端点、请求体、错误语义（200 + `success:false`）均不变，仅新增**凭据空值快速失败**，避免凭据未配置时暴露 SDK 原始报错：

```go
// handleTestConnection 中 SetCredentials 之前插入（[targets.go](webui/api/targets.go#L78) 内 +4 行；需新增 "strings" import）
if strings.HasPrefix(req.CloudType, "tc_") && (settings["tc_access_id"] == "" || settings["tc_access_key"] == "") {
    writeJSON(w, http.StatusOK, map[string]any{"success": false, "error": "腾讯云凭据未配置，请先在全局设置中填写"})
    return
}
if strings.HasPrefix(req.CloudType, "ali_") && (settings["ali_access_id"] == "" || settings["ali_access_key"] == "") {
    writeJSON(w, http.StatusOK, map[string]any{"success": false, "error": "阿里云凭据未配置，请先在全局设置中填写"})
    return
}
```

---

## 七、启动阶段决策

### 7.1 WebUI 模式

在 [main.go](main.go#L66-L93) 中：

```
LoadConfig() → 包含 SyncEnabled 字段
↓
创建 Provider、Resolver、Syncer
↓
将 Syncer 传入 WebUI（支持 Status/DryRun/Pause/Resume）
↓
go s.Run()   // 始终启动 Run()；启动时按 cfg.SyncEnabled 门控（见 5.2 启动逻辑）
```

**设计决策：`Run()` 始终启动**（本版修正，替代早期"`SyncEnabled=false` 时不启动 Run()"方案）。理由：若跳过启动，`Resume()` 发送的 `resumeCh` 信号将无人消费，暂停状态无法恢复。统一后：

- `SyncEnabled=false` 时，Run() 启动后立即进入 `waitForResume()` 暂停子循环，不执行 `syncAll()`；
- `Status()` 返回 `{running: true, enabled: false}`（引擎存活、开关关闭）；
- 用户点「开启」→ `Resume()` → 子循环收到信号 → 恢复 ticker 并立即执行首次 `syncAll()`（与 5.2 恢复逻辑一致）。

Syncer 实例始终注入 WebUI，因为：
- `DryRun()` 独立于 `Run()` 主循环，始终可用
- `Status()` 需要返回 `{running, enabled, last_sync}`
- 后续用户开启时通过 `Resume()` 恢复同步

### 7.2 `.env` 模式

在 [app/app.go](app/app.go#L44-L45) 中：

```go
go s.Run()   // 始终启动（与 7.1 一致）；启动门控由 Run() 内部按 cfg.SyncEnabled 决定
```

**收益**：统一后 `SYNC_ENABLED=false` 时 Run() 同样存活于暂停子循环，`doneCh` 正常关闭——早期"不启动 Run() 导致 `doneCh` 永不关闭、需特殊处理 `WaitForSignal()`"的问题自动消除，`WaitForSignal()` 无需改动。`.env` 模式无 WebUI 恢复途径，暂停状态重启后按环境变量生效（符合"重启生效"预期）。

---

## 八、WebUI 交互设计

### 8.1 控制位置：仪表盘（Dashboard.vue）

仪表盘是同步引擎状态的主展示页，已有「同步引擎」状态标签和操作按钮。开关放在此处最符合用户心智模型。

**布局调整（本次整合更新）**：

- **移除**「试运行」按钮（`dryRun()` 函数与 `NModal` 弹窗一并删除），Dry Run 由独立「运行测试」页统一承载；
- 「运行测试」入口以次级链接形式放在「同步引擎」状态卡片内；
- 操作卡片保留「立即同步」，新增「暂停/开启」开关按钮。

```
┌──────────────────────────────────────────────────┐
│  仪表盘                                           │
│  ┌───────────┐ ┌───────────┐ ┌────────────────┐  │
│  │ 同步引擎   │ │ 上次同步   │ │   操作          │  │
│  │ ● 运行中   │ │ 2026-...  │ │ [立即同步]      │  │
│  │            │ │           │ │ [暂停/开启]     │  │
│  │ [运行测试] │ │           │ │                 │  │
│  └───────────┘ └───────────┘ └────────────────┘  │
└──────────────────────────────────────────────────┘
```

### 8.2 状态标签变化

| 开关状态 | 引擎状态 | 标签颜色 | 标签文本 |
|---------|---------|---------|---------|
| 开启 | `enabled: true`、`running: true` | 绿色 (success) | 运行中 |
| 关闭 | `enabled: false`、`running: true`（Run() 存活于暂停子循环，见 7.1） | 橙色 (warning) | 已暂停 |

**说明**：`running` 保持现有语义（Run() 主循环存活，由 `setRunning()` 于 Run() 启停时切换），不承载"正在执行一轮同步"的语义；区分「空闲 / 同步中」需借助 SSE 事件流（`sync:start`/`sync:complete`），不作为本版必做项。

### 8.3 按钮状态联动（本次整合更新）

| 开关 | 「立即同步」按钮 | 「运行测试」入口 | 暂停/开启按钮 |
|------|----------------|----------------|-------------|
| 开启 | 可用 | 可用（跳转 `/run-test`） | 显示「暂停」 |
| 关闭 | **置灰（disabled）** | 可用（Dry Run 与连接测试均不受暂停影响） | 显示「开启」 |

「立即同步」按钮置灰时，hover 提示"请先开启同步引擎"。

**B5 预留**：`types.ts` 的 `SyncStatus` 预置可选字段 `enabled?: boolean`（后端未返回时默认 `true`）；Dashboard 操作区渲染以 `status.enabled === false` 为条件数据驱动，全局开关落地时仅需新增开关组件与 pause/resume API 调用，Dry Run 链路零改动。

### 8.4 建议不放在 Settings 页

虽然 `sync_enabled` 是一个配置，但它本质是**运行时控制**而非静态配置。放在 Dashboard 可让用户一眼看到当前状态并即时操作。Settings 页更适合「改完保存、下次生效」的配置项。（配置导入/导出属静态配置，按 8.6 收敛到 Settings 页，与本节不冲突。）

### 8.5 「运行测试」页（Dry Run + 连接测试，本次整合新增）

#### 定位

- 合并原仪表盘「试运行」弹窗、高级功能「Dry Run」与「连接测试」标签页，作为 Dry Run 与连接测试的**常驻诊断入口**，定位为正式测试诊断工具：结果常驻页面、可反复执行、有完整的 loading/成功/失败反馈；
- 页内以 `NTabs` 提供两个子标签页：**Tab 1「Dry Run」**、**Tab 2「连接测试」**；
- **保留云资源管理页（`Targets.vue`）弹窗内的表单级「测试连接」**：它用未保存的表单值验证凭据/资源 ID，与运行测试页定位互补（表单级验证 vs 独立诊断），不随页面收敛删除；两处共享同一端点、同一 `api.ts` 封装与 15s 超时增强。
- 标签页激活状态与路由 query 同步（`?tab=dryrun|connection`），刷新/分享可保留位置；
- 路由 `/run-test`，菜单「运行测试」（位于「同步日志」与「告警配置」之间——高级功能页删除后相邻项更新）。

#### Tab 1：Dry Run（三级分组视图：目标 → 域名 → 规则）

```
运行测试 → Tab 1：Dry Run
├── 操作区
│   ├── [执行 Dry Run] 按钮（loading 态，执行中禁用）
│   ├── 上次执行时间（lastRunAt，未执行为空）
│   └── 统计条：目标数 / 待添加总数 / 待删除总数 / 错误数
├── 结果区（三级分组）
│   └── 目标卡片（Provider，如 腾讯云轻量云 · ap-guangzhou · lhins-xxx）
│       └── 域名列表（rule.Host）
│           ├── 待添加规则明细表（列：协议 / 端口 / 动作 / CIDR / 描述）
│           └── 待删除规则明细表（列同上）
│       （result.error 非空的目标/域名显示错误行，不展开明细）
└── 空状态（按优先级展示）
    ├── 未执行：显示"尚未执行 Dry Run"
    ├── warnings 非空：按 warnings 文案展示（暂无云资源目标 / 暂无域名规则）
    ├── 执行后 results 为空且无 warnings："无待变更规则"
    └── 请求失败：message.error + 结果区错误提示
```

#### Tab 2：连接测试（由高级功能页迁入）

```
运行测试 → Tab 2：连接测试
├── 表单区：云类型选择（NSelect，4 云）+ 地域（NInput）+ 资源 ID（NInput）
├── [测试连接] 按钮（loading 态，执行中禁用，防连点）
└── 结果区：页面内文本展示
    ├── 成功：data.message（如"连接成功，当前 5 条规则"）
    ├── 业务失败：data.error（200 + success:false 语义，保留现有分支）
    ├── 凭据空值：后端快速失败提示（"腾讯云/阿里云凭据未配置，请先在全局设置中填写"）
    └── 超时：前端 AbortController 15s → "连接超时（15 秒），请检查网络或云 API 状态"
```

**连接测试交互规范**：

| 项目 | 规范 |
|------|------|
| 请求 | `POST /api/test-connection`，请求体 `{cloud_type, region, resource_id}` |
| 错误语义 | **沿用现有** `data.success ? data.message : data.error`（200 + `success:false` 是既有约定，不与 Dry Run 的 4xx/5xx 语义强行统一） |
| 超时（增强 1） | `AbortController` + 15s；`AbortError` 单独分支显示超时文案；`finally` 中 `clearTimeout`；**同样覆盖 `Targets.vue` 弹窗内的测试连接**（经 `api.ts` 的 `AbortSignal` 支持共享，不重复实现） |
| 凭据提示（增强 2） | 后端 `targets.go` 凭据空值快速失败（6.7），前端直接展示 `data.error` |
| 表单状态 | 迁移后保留 `testForm`/`testLoading` 形态；`cloudOptions` 改引 `constants.ts` 共享常量（与 Targets.vue 统一） |
| 实现形态 | **内嵌 `RunTest.vue`**（约 40 行，纯表单 + 结果文本，无共享状态），不抽 composable（符合简单轻量化；若未来复杂化再抽） |

#### 前端实现拆分

| 模块 | 说明 |
|------|------|
| `api.ts`（新增） | **统一 fetch 封装**：`res.ok` 检查 + `{"error":...}` 提取 + `AbortSignal` 支持；全站迁移，修复 5 处"失败误报成功"（Targets/Rules/Settings 保存路径） |
| `constants.ts`（新增） | 共享常量：`cloudOptions`（4 云选项）+ `cloudLabelMap`（消除 Targets/Advanced 两处重复定义） |
| `composables/useDryRun.ts`（新增） | Dry Run 共享组合逻辑：`loading` / `results` / `warnings` / `error` / `lastRunAt` / `run()`（基于 `api.ts`） |
| `components/DryRunResults.vue`（新增） | Dry Run 三级分组视图渲染组件（目标卡片 → 域名 → 规则明细表），含空状态分支 |
| `views/RunTest.vue`（新增） | 页面骨架：`NTabs`（Tab 1 Dry Run 引用上述组件；Tab 2 连接测试表单内嵌，基于 `api.ts` + 15s 超时） |
| `views/Dashboard.vue`（修改） | 删除 `dryRun()`、`dryrunResults`、`showDryrun`、`dryrunColumns`、`NModal`；新增「运行测试」链接与开关联动（8.1/8.3） |
| `views/Advanced.vue`（删除） | 三个标签页全部迁出后删除（见 8.7） |
| `views/Targets.vue`（修改） | 弹窗内表单级「测试连接」**保留**：接入 `api.ts` + 15s 超时 + `cloudOptions` 改引 `constants.ts`；保存/删除路径接入 `api.ts`（修复误报成功） |
| `views/Settings.vue`（修改，极小） | 保留现有「导出配置」「导入配置」按钮（已满足收敛目标）；`save()` 接入 `api.ts`（修复误报成功）；按需微调文案 |
| `main.ts` / `App.vue`（修改） | 新增路由 `/run-test` 与菜单「运行测试」；**删除路由 `/advanced` 与菜单「高级功能」** |

### 8.6 配置导入/导出收敛至全局设置页（本次整合新增）

**关键事实**：`Settings.vue` 表单末尾的保存行**已内置**「导出配置」「导入配置」按钮（[Settings.vue](webui/frontend/src/views/Settings.vue#L92-L101)），且导入成功后自动刷新设置表单（L41-L42）；`Advanced.vue` 配置标签页是对同一对端点的**重复实现**（且导入后不刷新表单）。

**方案**：不新增区块——**保留设置页现有按钮行作为唯一入口**（恰在表单末尾，即"整合进全局设置的末尾"），删除 `Advanced.vue` 的配置导入/导出标签页及 `exportConfig`/`handleImport` 重复实现。可选微调：按钮行下增加一行说明文案（"导入会覆盖当前配置"）。

**预期收益**：消除重复实现；导入/导出与设置表单同页，导入后即时刷新所见即所得。

### 8.7 高级功能页删除（本次整合新增）

高级功能页三个标签页全部迁出后页面为空，**删除 `/advanced` 页面**：

| 步骤 | 内容 |
|------|------|
| 1 | `Advanced.vue` 文件删除 |
| 2 | `main.ts` 删除路由 `{ path: '/advanced', ... }` |
| 3 | `App.vue` 菜单删除「高级功能」项 |
| 4 | 既有导航跳转引用（如有）一并清理 |

---

## 九、兼容性分析

### 9.1 Dry Run（本次整合改写）

**暂停兼容（不变）**：`POST /api/sync/dryrun` 不检查暂停状态，始终可用。`DryRun()` 是独立方法，不经过 `Run()` 主循环，直接访问 `s.providers` 和 `s.cfg.DomainRules`，这两个字段在 Syncer 创建后就始终可用。使用场景契合：**先暂停 → 配置规则 → Dry Run 预览 → 确认无误 → 开启同步**。

**响应结构破坏性变更（本次新增，已决策接受）**：`to_add`/`to_delete` 由 `number` 改为 `RuleChange[]`，响应由裸数组改为 `{results, warnings}` 包装对象。项目仍处于构建期，不考虑旧版兼容；前端消费方本次统一改造。

**前端入口合并（本次新增）**：仪表盘「试运行」与高级功能「Dry Run」标签页移除，由「运行测试」页 Tab 1 统一承载；原两处交互缺陷（无 loading、不检查 `res.ok`、误报成功）随合并一并修复。

### 9.2 热重载

**完全兼容（并新增快照语义）**。

热重载通过 `configCh` 传递新配置。正常运行时和暂停等待时均正确响应：

- 正常运行时：更新配置 + Reset ticker（现有行为）
- 暂停等待时：仅更新配置（ticker 已停止，不触发同步）

`ReloadProviders()` 通过直接赋值 `s.providers` 实现，不经过 channel，暂停状态下同样生效。用户可在暂停期间修改 Targets，Dry Run 能立即反映最新配置。

**Dry Run 执行期间热重载（本次新增约定）**：`DryRun()` 开始时在 `s.mu.RLock()` 保护下对 `s.providers`（切片引用）与 `s.cfg`（快照）各做一次快照，整个遍历使用快照；期间发生的重载不影响本次结果，下次 Dry Run 生效。同时消除与 `ReloadProviders()`/`Reload()` 并发遍历的数据竞态。

### 9.3 事件总线（EventBus）

**不需要改造。**

暂停/恢复本身不产生 EventBus 事件（保持简洁）。同步过程中的事件（`EventSyncStart`、`EventDomainSyncComplete`、`EventSyncError`、`EventDNSFailed`）仅在 `syncAll()` 实际执行时产生。暂停期间无同步轮次，自然不会产生这些事件。Dry Run 同样不触发事件（现有行为不变）。

### 9.4 同步日志（StoreLogWriter）

**不需要改造。**

[StoreLogWriter](webui/api/logwriter.go) 订阅 `EventDomainSyncComplete` 和 `EventSyncError`。暂停期间无同步轮次，无事件产生，无日志记录。日志记录中不会出现异常的空白期。

### 9.5 告警通知

暂停期间 DNS 解析和云 API 调用都不会发生（`syncAll()` 被门控阻止），不会产生误报告警。

### 9.6 连接测试与配置导入导出（本次整合新增）

| 项 | 兼容性结论 |
|----|-----------|
| `POST /api/test-connection` | 端点、请求体、错误语义（200 + `success:false`）**全部不变**；仅新增凭据空值快速失败分支（对正常流程无影响）；`Targets.vue` 弹窗内表单级测试连接**保留**，随统一封装与超时增强升级 |
| 连接测试暂停可用 | 端点不依赖 `d.Syncer`，暂停状态下可用（与 Dry Run 一致） |
| 连接测试超时 | 纯前端增强（AbortController），后端无感知 |
| `GET /api/config/export`、`POST /api/config/import` | 端点不变；前端入口从高级功能页收敛到全局设置页（设置页本就具备，删除重复实现） |
| 高级功能页删除 | 纯前端页面收敛；`/advanced` 路由与菜单移除，不影响任何 API |

---

## 十、状态转换图

```
                         ┌─────────────────────┐
                         │     应用启动          │
                         │ 读取 cfg.SyncEnabled  │
                         └──────┬──────────────┘
                                │
                   ┌────────────┴────────────┐
                   ▼                         ▼
          SyncEnabled=true           SyncEnabled=false
                   │                         │
                   ▼                         ▼
           ┌──────────────┐          ┌──────────────┐
           │   正常运行    │          │   暂停状态    │
           │ ticker 驱动   │          │ Run() 暂停子循环 │
           │ 可手动触发    │          │ Dry Run 可用  │
           │ 可热重载      │          │ 可热重载      │
           └───┬──────┬───┘          └──────┬───────┘
               │      │                     │
               │      │  用户点「暂停」       │
               │      │  DB写false+Pause()  │
               │      └────────────────────►│
               │                            │
               │     用户点「开启」           │
               │     DB写true+Resume()      │
               │◄───────────────────────────┘
               │
               │  进程退出 / 异常终止
               ▼
         ┌──────────┐
         │  进程结束  │
         └──────────┘
         下次启动 → 从 DB/ENV 读取 SyncEnabled → 决定初始状态
```

（暂停状态下「运行测试」页的 Dry Run 与连接测试均可用——二者不依赖 `Run()` 主循环。）

---

## 十一、风险评估

| 风险项 | 等级 | 说明 | 缓解措施 |
|--------|------|------|---------|
| 暂停中应用异常退出导致状态不一致 | 🟢 低 | API 先写 DB 后通知 Syncer，DB 状态始终正确。重启后正确读取 | — |
| 暂停/恢复信号竞态（快节奏点击） | 🟡 中 | 快速点击可能造成信号堆积 | channel 容量 1 + `select/default`；前端 loading 状态禁用按钮 |
| 停止信号在暂停子循环中到达 | 🟢 低 | `waitForResume()` 需同时监听 `stopCh`，返回后外层 `Run()` 退出 | 与现有优雅退出行为一致 |
| 暂停期间热重载触发意外同步 | 🟢 低 | `waitForResume()` 中 `configCh` 仅更新配置，不触发 `syncAll()` | — |
| DryRun 响应结构破坏性变更 | 🟢 低 | `to_add`/`to_delete` 改为数组 + 响应包装化 | 项目处于构建期不考虑旧版兼容；前端消费方同步改造 |
| 配置导入/设置写入导致开关状态不一致 | 🟡 中 | 热重载仅更新 cfg，`syncEnabled` 运行时镜像不更新 | `Reload()` 同步 `Pause()`/`Resume()` 或直接镜像（见 5.6） |
| 前端"失败误报成功" | 🟡 中 | 5 处保存/删除路径无 `res.ok` 检查 | 统一 `api.ts` 封装（见 8.5/12.1） |
| DryRun 重复点击并发执行 | 🟡 中 | 连点可并发多次 DryRun | 前端 loading 禁用 + 后端 `TryLock()` 防重入（409） |
| DryRun 无限制冲击云 API | 🟡 中 | 循环无限速，频繁执行可触发云厂商频率限制 | `DryRun()` 内复用 `rateLimitInterval(ct)` 加入间隔 |
| DryRun 与热重载并发竞态 | 🟡 中 | 遍历 `s.providers`/`s.cfg` 无读锁 | 开始前快照 + `RLock` 保护（见 9.2） |
| 连接测试无超时（云 API 卡住） | 🟡 中 | loading 无限转圈，无任何反馈 | 前端 `AbortController` 15s 超时 + 超时文案（见 8.5 Tab 2） |
| 连接测试凭据空值提示不友好 | 🟢 低 | 直接暴露 SDK 原始报错 | 后端凭据空值快速失败（见 6.7） |
| 高级功能页删除遗漏引用 | 🟢 低 | 路由/菜单/跳转引用残留 | 按 8.7 清单逐步核对；WebUI 内部工具，影响面小 |
| 三级分组视图明细渲染性能 | 🟢 低 | 单台服务器少量 IP 场景，明细量小 | 一次请求返回全部明细；分组在前端内存组织 |
| 现有单元测试受影响 | 🟢 低 | 项目测试覆盖主要在 resolver/portconv/tag，Syncer 本身几乎无测试 | — |

---

## 十二、实施计划

### 12.1 改动文件清单

**Dry Run 与测试功能升级（Phase 0，不依赖全局开关）：**

| 文件 | 改动量 | 说明 |
|------|--------|------|
| `syncer/syncer.go` | ~40 行 | `DryRunResult` 明细化 + `DryRun()` 限速/快照锁/防重入标志 |
| `provider/common.go` | ~20 行 | `RuleChange` 摘要构造函数 |
| `webui/api/sync.go` | ~20 行 | `handleSyncDryRun` 响应包装 `{results, warnings}` + 409 防重入分支 |
| `webui/api/targets.go` | +4 行 | 连接测试凭据空值快速失败（见 6.7） |
| `webui/api/deps.go` | 0~3 行 | 无需改（或按接口调整） |
| `webui/frontend/src/types.ts` | ~20 行 | `DryRunResult`/`RuleChange` 更新 + `SyncStatus.enabled?` 可选字段 + 响应包装类型 + `TestConnectionResult`（可选） |
| `webui/frontend/src/api.ts` | 新增 | 统一 fetch 封装（res.ok 检查 + error 提取 + AbortSignal） |
| `webui/frontend/src/constants.ts` | 新增 | `cloudOptions`/`cloudLabelMap` 共享常量 |
| `webui/frontend/src/composables/useDryRun.ts` | 新增 | Dry Run 共享组合逻辑（loading/results/warnings/error/lastRunAt/run()） |
| `webui/frontend/src/components/DryRunResults.vue` | 新增 | 三级分组视图渲染 + 空状态 |
| `webui/frontend/src/views/RunTest.vue` | 新增 | 运行测试页骨架（`NTabs`：Tab 1 Dry Run + Tab 2 连接测试表单内嵌，含 15s 超时增强） |
| `webui/frontend/src/views/Dashboard.vue` | ~20 行 | 删除试运行按钮/弹窗/相关状态；新增「运行测试」链接 + 开关联动预留（B5） |
| `webui/frontend/src/views/Advanced.vue` | **删除** | 三个标签页全部迁出（Dry Run/连接测试 → 运行测试页；配置导入导出 → 设置页） |
| `webui/frontend/src/views/Targets.vue` | ~15 行 | 弹窗测试连接**保留**并接入 `api.ts` + 15s 超时；保存/删除接入 `api.ts`；`cloudOptions` 引共享常量 |
| `webui/frontend/src/views/Settings.vue` | 0~5 行 | 保留现有导入导出按钮（已满足收敛）；`save()` 接入 `api.ts`；按需微调 |
| `webui/frontend/src/main.ts` / `App.vue` | ~4 行 | 新增路由 `/run-test` + 菜单「运行测试」；删除路由 `/advanced` + 菜单「高级功能」 |
| `Design1.md` | 同步 | 特性表（L274「Dry Run \| 执行到 Diff 为止，不实际写入」）与页面列表（L195「高级功能（Dry Run、配置导入/导出、健康检查）」）措辞对齐（注：Design1.md 无 12.9 章节，Dry Run 描述实际位于特性表与页面列表） |

**同步全局开关（Phase 1-5，原设计）：**

| 文件 | 改动量 | 说明 |
|------|--------|------|
| `config/config.go` | +1 字段 | `Config.SyncEnabled bool` |
| `config/store.go` | +2 行 | `LoadConfig()` 读取 `sync_enabled` |
| `config/env.go` | +2 行 | `ParseEnv()` 读取 `SYNC_ENABLED` |
| `syncer/syncer.go` | ~60 行 | 新增 `pauseCh`/`resumeCh` + `Pause()`/`Resume()`/`IsEnabled()` + `Run()` 改造 + `waitForResume()` |
| `webui/api/deps.go` | +2 方法签名 | `Syncer` 接口新增 `Pause()` / `Resume()` |
| `webui/api/sync.go` | ~30 行 | `handleSyncPause` / `handleSyncResume` + 路由注册 + `handleSyncTrigger` 暂停保护 |
| `main.go` | 0~3 行 | 无需条件分支；`go s.Run()` 保持原样（启动门控在 Run() 内，见 7.1） |
| `app/app.go` | 0~3 行 | 无需条件分支与 `WaitForSignal` 适配（7.2 统一启动，死锁风险消除） |
| `webui/frontend/src/views/Dashboard.vue` | ~20 行 | 开关组件 + 暂停/恢复 API 调用 + 按钮状态联动（B5 预留直接生效） |

**合计**：测试功能升级约 90 行（后端：syncer 40 + common 20 + sync 20 + targets 4 + deps 3）+ 新增 5 个前端文件（`api.ts`/`constants.ts`/`useDryRun.ts`/`DryRunResults.vue`/`RunTest.vue`）+ 删除 1 个前端文件 + 修改 4 个页面；全局开关约 95 行（后端）+ 前端约 35 行。零 schema 变更。

### 12.2 实施顺序

| 阶段 | 内容 | 依赖 |
|------|------|------|
| **Phase 0-1** | 前端共享基建：`api.ts` 统一封装 + `constants.ts` + `useDryRun.ts` + 类型更新（`types.ts`） | 无 |
| **Phase 0-2** | 后端 DryRun 升级：明细化 + 响应包装 + 限速 + 快照锁 + 防重入；连接测试凭据空值提示（`targets.go`） | Phase 0-1 的接口约定 |
| **Phase 0-3** | 运行测试页落地：`RunTest.vue`（Dry Run + 连接测试双标签，含 15s 超时）+ `DryRunResults.vue` + 路由/菜单新增 | Phase 0-1, 0-2 |
| **Phase 0-4** | 页面收敛与收尾：`Settings.vue` 导入导出确认 + `save()` 接入 `api.ts` → 删除 `Advanced.vue` + `/advanced` 路由与菜单 → `Targets.vue` 测试连接接入统一封装与超时 → Dashboard 收尾（「运行测试」链接 + 开关联动预留 B5） | Phase 0-3 |
| **Phase 1** | 数据层：`Config.SyncEnabled` + `Store` 读写 + `env.go` 解析 | 无 |
| **Phase 2** | Syncer 核心：暂停门控 + `Pause()`/`Resume()`/`IsEnabled()` | Phase 1 |
| **Phase 3** | API 层：端点 + 路由 + 接口扩展 + Trigger 保护 | Phase 1, 2 |
| **Phase 4** | 启动决策：`main.go`/`app.go` 统一为始终启动 `Run()`（门控在 Run() 内，见 7.1/7.2） | Phase 1, 2 |
| **Phase 5** | 前端：Dashboard 开关 + 状态联动（复用 Phase 0-4 预留） | Phase 3 |

### 12.3 验收标准

**全局开关（原 8 条）：**

1. **默认行为不变**：首次启动（无 `sync_enabled` 键）时，同步立即开始（向后兼容）
2. **暂停持久化**：暂停后 → 重启 → 同步保持暂停
3. **开启持久化**：开启后 → 重启 → 同步自动恢复
4. **暂停阻止写入**：暂停状态下，ticker 和手动触发均不会执行 `syncAll()`
5. **Dry Run 始终可用**：暂停和运行状态下均可执行 Dry Run
6. **优雅暂停**：暂停时若正在同步，等待当前轮次完成后进入暂停
7. **`.env` 模式**：`SYNC_ENABLED=false` 时启动不执行同步（Run() 存活于暂停子循环，`doneCh` 正常关闭）
8. **前端状态一致**：Dashboard 开关状态与后端实际状态实时同步

**Dry Run 与测试功能升级（本次新增）：**

9. **明细化**：运行测试页返回并展示逐条规则明细（协议/端口/动作/CIDR/描述），而非仅计数
10. **三级分组**：Dry Run 结果按「目标 → 域名 → 规则」分组展示，待添加/待删除分列
11. **空状态**：未执行、无目标、无规则、无变更四种空状态均有对应文案（warnings 语义化）
12. **错误可见**：400/409/500 均通过 `message.error` 展示 `data.error`，不再误报成功
13. **防连点**：执行中按钮禁用；后端重复请求返回 409
14. **限速合规**：连续执行 Dry Run 时，同云厂商请求间隔不小于 `rateLimitInterval`
15. **入口合并**：仪表盘、高级功能不再出现 Dry Run 入口，全部由「运行测试」页承载
16. **连接测试迁入**：连接测试在「运行测试」页 Tab 2 可用，表单/结果语义与迁移前一致；暂停状态下可用；`Targets.vue` 弹窗内的表单级测试连接**保留**且行为一致
17. **连接测试超时**：云 API 15s 无响应时前端提示"连接超时（15 秒）"，loading 解除（运行测试页与 `Targets.vue` 弹窗均覆盖）
18. **凭据空值提示**：凭据未配置时连接测试返回"腾讯云/阿里云凭据未配置，请先在全局设置中填写"
19. **页面收敛**：高级功能页删除（`/advanced` 路由、菜单、`Advanced.vue` 文件均移除）；配置导入/导出仅由全局设置页承载，导入后设置表单即时刷新
20. **开关状态一致**：配置导入或设置写入含 `sync_enabled` 变更后，Syncer 实际行为与 DB 状态一致（热重载同步开关，见 5.6）
21. **统一封装**：全站 fetch 经 `api.ts` 封装后，任何非 2xx 响应不再误报成功（重点验证 Targets/Rules/Settings 保存路径）

### 12.4 低优先可选附项（深度审查发现，可作独立小项处理）

| # | 现状 | 问题 | 建议 |
|---|------|------|------|
| A | `env.go` 的 `INTERVAL` 解析失败**返回错误**；`store.go` 的 `LoadConfig` 解析失败**静默 fallback 5m** | 两模式行为不一致；WebUI 用户改了 interval 实际未生效 | 统一校验策略；`Settings.vue` 的 interval 输入加格式校验与提示 |
| B | `server.go` 用 `Handle("/")` 兜底静态文件 | 未知 API 路径返回 HTML，前端 `res.json()` 解析失败 | 随 `api.ts` 统一封装天然免疫（解析失败进 catch），无需专门处理 |
| C | `main.go` 热重载重建 Provider 失败时 `continue` | 部分目标静默丢失（仅 slog.Error） | 可选：ReloadFunc 返回错误并在日志中汇总提示 |
| D | `Settings.vue` 表单缺 `dns_fail_threshold`（后端已支持该 key） | 前端无法配置 DNS 失败阈值 | 设置页补输入框（与 `dns_timeout` 并列） |
