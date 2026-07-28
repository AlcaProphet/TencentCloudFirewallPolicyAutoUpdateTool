# Design2.md — 同步全局开关功能设计

> 本文档描述「同步全局开关」功能的完整设计方案，涵盖需求分析、存储方案、核心机制、API 设计、WebUI 交互、兼容性分析及风险评估。所有决策以 [Design1.md](./Design1.md) 架构设计为依据，编码约束遵循 [AGENTS.md](./AGENTS.md)。

---

## 一、概述

### 1.1 功能定义

为 FWAlizer 增加一个**可持久化状态的同步全局开关**，允许用户按需开启或暂停自动同步。开关状态持久化到 SQLite（WebUI 模式）或环境变量（`.env` 模式），应用重启后自动恢复。

### 1.2 核心需求

| 需求 | 说明 |
|------|------|
| **按需启动** | 开关关闭时，Syncer 不执行任何同步（包括定时和手动触发）；Dry Run 不受影响 |
| **状态持久化** | 关闭后重启应用，同步保持暂停；开启后重启，自动恢复同步 |
| **优雅暂停** | 暂停时若正在执行同步，等待当前轮次完成后再进入暂停状态 |
| **即时恢复** | 开启后立即恢复定时同步（等待下一个 ticker 周期，或用户手动触发） |
| **双模式兼容** | WebUI 模式通过 SQLite + API 控制；`.env` 模式通过环境变量控制 |

### 1.3 设计动机

当前 FWAlizer 启动后 `Syncer.Run()` 立即执行 `syncAll()`（[syncer.go](syncer/syncer.go#L67)），无任何缓冲机制。用户首次配置或维护时缺乏「先验证再执行」的安全窗口。虽然 Dry Run 提供了预览能力，但缺乏在源头阻止写入的手段。

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

### 2.2 现有对外接口

[webui/api/deps.go](webui/api/deps.go#L12-L17) 暴露的 Syncer 接口：

```go
type Syncer interface {
    Status() syncer.SyncStatus
    TriggerSync()
    DryRun() ([]syncer.DryRunResult, error)
}
```

当前无 `Pause()` / `Resume()` / `SetEnabled()` 方法。

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
| `GET /api/sync/status` | `SyncStatus` 新增 `enabled bool` 字段 |
| `POST /api/sync/trigger` | 暂停时返回 HTTP `409 Conflict` + `{"error":"同步已暂停，请先开启"}` |
| `POST /api/sync/dryrun` | **无需改动** — 始终可用 |

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

---

## 七、启动阶段决策

### 7.1 WebUI 模式

在 [main.go](main.go#L66-L93) 中：

```
LoadConfig() → 包含 SyncEnabled 字段
↓
创建 Provider、Resolver、Syncer（始终创建，即使暂停）
↓
将 Syncer 传入 WebUI（支持 Status/DryRun/Pause/Resume）
↓
if cfg.SyncEnabled:
    go s.Run()      // 正常启动，立即 syncAll()
else:
    不启动 Run() goroutine
    Syncer 实例存在但 Run() 未执行
    slog.Info("同步已暂停，请通过 WebUI 开启")
```

Syncer 实例即使 `Run()` 未启动也需注入 WebUI，因为：
- `DryRun()` 独立于 `Run()` 主循环，始终可用
- `Status()` 需要返回 `{running: false, enabled: false}`
- 后续用户开启时通过 `Resume()` 触发首次 `syncAll()`

### 7.2 `.env` 模式

在 [app/app.go](app/app.go#L44-L45) 中：

```go
if cfg.SyncEnabled {
    go s.Run()
} else {
    slog.Info("同步已禁用（SYNC_ENABLED=false），进程将持续运行但不执行同步")
    // 等待信号，但不启动同步循环
    syncer.WaitForSignal(s)
}
```

**注意**：当 `SYNC_ENABLED=false` 时，`s.Run()` 未启动导致 `doneCh` 永不关闭。需特殊处理 `WaitForSignal()` 使其在无 Run() 时也能被 Ctrl+C 中断。

---

## 八、WebUI 交互设计

### 8.1 控制位置：仪表盘（Dashboard.vue）

仪表盘是同步引擎状态的主展示页，已有「同步引擎」状态标签和操作按钮。开关放在此处最符合用户心智模型。

**布局调整**：

```
┌──────────────────────────────────────────────────┐
│  仪表盘                                           │
│  ┌───────────┐ ┌───────────┐ ┌────────────────┐  │
│  │ 同步引擎   │ │ 上次同步   │ │   操作          │  │
│  │ ● 运行中   │ │ 2026-...  │ │ [立即同步]      │  │
│  │            │ │           │ │ [试运行]        │  │
│  │ [⏸ 暂停]  │ │           │ │                 │  │
│  └───────────┘ └───────────┘ └────────────────┘  │
└──────────────────────────────────────────────────┘
```

### 8.2 状态标签变化

| 开关状态 | 引擎状态 | 标签颜色 | 标签文本 |
|---------|---------|---------|---------|
| 开启 + 空闲 | `running: true` | 绿色 (success) | 运行中 |
| 开启 + 同步中 | `running: true` | 蓝色 (info) | 同步中 |
| 关闭 | `running: false` | 橙色 (warning) | 已暂停 |

### 8.3 按钮状态联动

| 开关 | 「立即同步」按钮 | 「试运行」按钮 | 暂停/开启按钮 |
|------|----------------|-------------|-------------|
| 开启 | 可用 | 可用 | 显示「暂停」 |
| 关闭 | **置灰（disabled）** | 可用 | 显示「开启」 |

「立即同步」按钮置灰时，hover 提示"请先开启同步引擎"。

### 8.4 建议不放在 Settings 页

虽然 `sync_enabled` 是一个配置，但它本质是**运行时控制**而非静态配置。放在 Dashboard 可让用户一眼看到当前状态并即时操作。Settings 页更适合「改完保存、下次生效」的配置项。

---

## 九、兼容性分析

### 9.1 Dry Run

**完全兼容，无需改动。**

[DryRun()](syncer/syncer.go#L156-L189) 是独立方法，不经过 `Run()` 主循环，不检查暂停状态。它直接访问 `s.providers` 和 `s.cfg.DomainRules`，这两个字段在 Syncer 创建后就始终可用。

使用场景契合：**先暂停 → 配置规则 → Dry Run 预览 → 确认无误 → 开启同步**。

### 9.2 热重载

**完全兼容。**

热重载通过 `configCh` 传递新配置。正常运行时和暂停等待时均正确响应：

- 正常运行时：更新配置 + Reset ticker（现有行为）
- 暂停等待时：仅更新配置（ticker 已停止，不触发同步）

`ReloadProviders()` 通过直接赋值 `s.providers` 实现，不经过 channel，暂停状态下同样生效。用户可在暂停期间修改 Targets，Dry Run 能立即反映最新配置。

### 9.3 事件总线（EventBus）

**不需要改造。**

暂停/恢复本身不产生 EventBus 事件（保持简洁）。同步过程中的事件（`EventSyncStart`、`EventDomainSyncComplete`、`EventSyncError`、`EventDNSFailed`）仅在 `syncAll()` 实际执行时产生。暂停期间无同步轮次，自然不会产生这些事件。

### 9.4 同步日志（StoreLogWriter）

**不需要改造。**

[StoreLogWriter](webui/api/logwriter.go) 订阅 `EventDomainSyncComplete` 和 `EventSyncError`。暂停期间无同步轮次，无事件产生，无日志记录。日志记录中不会出现异常的空白期。

### 9.5 告警通知

暂停期间 DNS 解析和云 API 调用都不会发生（`syncAll()` 被门控阻止），不会产生误报告警。

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
           │ ticker 驱动   │          │ Run() 未启动  │
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

---

## 十一、风险评估

| 风险项 | 等级 | 说明 | 缓解措施 |
|--------|------|------|---------|
| 暂停中应用异常退出导致状态不一致 | 🟢 低 | API 先写 DB 后通知 Syncer，DB 状态始终正确。重启后正确读取 | — |
| 暂停/恢复信号竞态（快节奏点击） | 🟡 中 | 快速点击可能造成信号堆积 | channel 容量 1 + `select/default`；前端 loading 状态禁用按钮 |
| 停止信号在暂停子循环中到达 | 🟢 低 | `waitForResume()` 需同时监听 `stopCh`，返回后外层 `Run()` 退出 | 与现有优雅退出行为一致 |
| 暂停期间热重载触发意外同步 | 🟢 低 | `waitForResume()` 中 `configCh` 仅更新配置，不触发 `syncAll()` | — |
| `.env` 模式 `syncEnabled=false` 时 `WaitForSignal` 死锁 | 🟡 中 | `s.Run()` 未启动 → `doneCh` 永不关闭 | `WaitForSignal()` 需特殊处理：直接监听信号，不等待 `doneCh` |
| 现有单元测试受影响 | 🟢 低 | 项目测试覆盖主要在 resolver/portconv/tag，Syncer 本身几乎无测试 | — |

---

## 十二、实施计划

### 12.1 改动文件清单

| 文件 | 改动量 | 说明 |
|------|--------|------|
| `config/config.go` | +1 字段 | `Config.SyncEnabled bool` |
| `config/store.go` | +2 行 | `LoadConfig()` 读取 `sync_enabled` |
| `config/env.go` | +2 行 | `ParseEnv()` 读取 `SYNC_ENABLED` |
| `syncer/syncer.go` | ~60 行 | 新增 `pauseCh`/`resumeCh` + `Pause()`/`Resume()`/`IsEnabled()` + `Run()` 改造 + `waitForResume()` |
| `webui/api/deps.go` | +2 方法签名 | `Syncer` 接口新增 `Pause()` / `Resume()` |
| `webui/api/sync.go` | ~30 行 | `handleSyncPause` / `handleSyncResume` + 路由注册 + `handleSyncTrigger` 暂停保护 |
| `main.go` | ~5 行 | 启动时检查 `cfg.SyncEnabled` 决定是否 `go s.Run()` |
| `app/app.go` | ~5 行 | `.env` 模式同样检查 + `WaitForSignal` 适配 |
| `webui/frontend/src/views/Dashboard.vue` | ~35 行 | 开关组件 + 暂停/恢复 API 调用 + 按钮状态联动 |
| `webui/frontend/src/types.ts` | +1 字段 | `SyncStatus.enabled: boolean` |

**合计**：后端约 105 行，前端约 35 行，总计约 **140 行**。零 schema 变更，零新文件。

### 12.2 实施顺序

| 阶段 | 内容 | 依赖 |
|------|------|------|
| **Phase 1** | 数据层：`Config.SyncEnabled` + `Store` 读写 + `env.go` 解析 | 无 |
| **Phase 2** | Syncer 核心：暂停门控 + `Pause()`/`Resume()`/`IsEnabled()` | Phase 1 |
| **Phase 3** | API 层：端点 + 路由 + 接口扩展 + Trigger 保护 | Phase 1, 2 |
| **Phase 4** | 启动决策：`main.go` + `app.go` 的条件启动 | Phase 1, 2 |
| **Phase 5** | 前端：Dashboard 开关 + 状态联动 | Phase 3 |

### 12.3 验收标准

1. **默认行为不变**：首次启动（无 `sync_enabled` 键）时，同步立即开始（向后兼容）
2. **暂停持久化**：暂停后 → 重启 → 同步保持暂停
3. **开启持久化**：开启后 → 重启 → 同步自动恢复
4. **暂停阻止写入**：暂停状态下，ticker 和手动触发均不会执行 `syncAll()`
5. **Dry Run 始终可用**：暂停和运行状态下均可执行 Dry Run
6. **优雅暂停**：暂停时若正在同步，等待当前轮次完成后进入暂停
7. **`.env` 模式**：`SYNC_ENABLED=false` 时启动不执行同步
8. **前端状态一致**：Dashboard 开关状态与后端实际状态实时同步
