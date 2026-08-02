# Design3.md — WebUI 体验优化与同步日志修复设计

> 本文档描述 WebUI 的 12 项改进构想：全局明暗主题、仪表盘卡片化重设计、同步日志页重排与修复（历史计数、实时日志一致性、错误报告、清空记录）、中文云产品名展示、首次引导与 Keys 缺失提示。
>
> **文档定位：** 本文档属于**设计构想（非强制规定）**。编码约束遵循 [AGENTS.md](./AGENTS.md)（项目唯一强要求文档）；与 AGENTS.md 或用户决策冲突时，以用户确认为准；详细构建方案由后续 Build 文档承载。
>
> **决策记录（用户已确认）**：
>
> - **主题持久化**：后端 DB（`settings` 表），跨浏览器 / Docker 多端一致；
> - **仪表盘布局**：方案 B（2×2 大卡片网格）；
> - **实时事件版块**：删除（历史记录置顶，运行日志默认展开）；
> - **Keys 缺失策略**：仅前端提示，不阻止保存。
>
> **设计原则**：延续 AGENTS.md「简单轻量化」——不引入新依赖（主题用 Naive UI 内置 `darkTheme`）、不新增聚合 API（统计复用现有端点）、schema 零变更（复用 `settings` 键值表与 `sync_logs` 既有字段）。

---

## 一、概述

### 1.1 改进清单

本次共 12 项改进，按性质分四组：

| 组别 | 编号 | 改进项 | 性质 |
|------|------|--------|------|
| UI 体验 | 1 | 页面全局明暗主题切换 | 新功能 |
| UI 体验 | 2 | 仪表盘卡片化重设计（方案 B：2×2） | 重设计 |
| UI 体验 | 3 | 移除仪表盘同步引擎框内「运行测试」超链接 | 微调 |
| UI 体验 | 4 | 域名规则「适用目标」显示中文云产品名 | 修复 |
| 日志页 | 5 | 移除「实时事件」版块，历史记录移至最顶部 | 重排 |
| 日志页 | 6 | 实时运行日志默认展开 | 微调 |
| 日志页 | 7 | 历史记录新增/删除计数始终为 0 | **Bug 修复** |
| 日志页 | 8 | 实时日志与 `docker compose logs` 不一致且显示不全 | **Bug 修复** |
| 日志页 | 9 | 历史记录 failed 展示错误原因（点击弹出错误报告） | 新功能 |
| 日志页 | 10 | 历史记录增加清空按钮 | 新功能 |
| 引导 | 11 | 首次启动提示优先填写各类 Keys | 新功能 |
| 引导 | 12 | 云资源管理选择未配 Keys 平台时提示 | 新功能 |

### 1.2 设计动机

1. **视觉基础缺失**：`App.vue` 的 `NConfigProvider` 未配置 `theme`（恒为亮色），无暗色模式；运行日志 `<pre>` 硬编码深色终端风格，亮色页面中为唯一深色块，主题化后需保持自洽。
2. **仪表盘信息密度低**：现为 3 列小卡片 + `size="small"` 小按钮（`Dashboard.vue`），空间利用率低，关键状态（引擎状态、上次同步）不突出，无统计概览。
3. **两个数据缺陷**：历史记录「新增/删除」列恒为 0（事件链路不携带计数）；实时日志与终端输出不一致且丢日志（格式差异 + 无历史回放 + 通道丢弃），用户排查问题时会误判。
4. **诊断路径不完整**：历史记录 failed 无错误详情展示（DB 已存 `error` 字段未使用）；无法清空历史。
5. **首次使用门槛**：WebUI 启动后无任何引导，新用户不知道需先填 Keys 再配置目标（`main.go` 仅在 stdout 打一行日志）。

---

## 二、现有架构分析

### 2.1 相关文件与现状

| 文件 | 现状要点 |
|------|---------|
| [App.vue](webui/frontend/src/App.vue) | `NConfigProvider` 无 `theme`；侧边栏 200px 固定；无主题切换入口 |
| [Dashboard.vue](webui/frontend/src/views/Dashboard.vue) | 3 列 `NGrid` 小卡片；同步引擎卡片内含「运行测试 →」文本链接（L77-79）；5s 轮询 `GET /api/sync/status` |
| [Rules.vue](webui/frontend/src/views/Rules.vue) | `loadTargets()` 的 `label` 为 `` `${t.cloud_type} / ${t.resource_id}` ``（L45），原始英文云类型；`constants.ts` 已有 `cloudLabelMap` 未使用 |
| [Logs.vue](webui/frontend/src/views/Logs.vue) | 结构：实时事件表格（顶部）→ 历史记录 → 运行日志（`NCollapse` 默认收起）；历史列无 error 展示；logLines 上限 200；订阅 `GET /api/sync/events` 与 `GET /api/logs/stream` 两个 SSE |
| [syncer/syncer.go](syncer/syncer.go) | `syncDomainInternal` 发布 `EventDomainSyncComplete` 时 Data 仅 `{provider, domain}`（L422-426），无计数 |
| [syncer/retry.go](syncer/retry.go) | `retrySync` 仅返回 `error`，diff 执行量（`len(diff.ToAdd)`/`len(diff.ToDelete)`）未对外暴露 |
| [webui/api/logwriter.go](webui/api/logwriter.go) | `StoreLogWriter.OnEvent` 仅写 `Result`，从不写 `Added`/`Deleted`（恒 0） |
| [webui/api/logstream.go](webui/api/logstream.go) | 自定义行格式 `15:04:05 [INFO] msg attrs`（无日期/时区）；订阅 channel 容量 64、满则跳过；无历史回放 |
| [app/logutil.go](app/logutil.go) | `MultiHandler` 将 stdout TextHandler 与 LogBroadcaster 并联，二者格式互不相同 |
| [config/store.go](config/store.go) | `sync_logs` 表已含 `added/deleted/error` 字段；无清空方法 |
| [config/settings 表](config/store.go) | `settings` 键值表可承载 `theme` 键，`GET/PUT /api/settings` 机制现成 |

### 2.2 问题 7 根因：历史记录计数恒为 0

**数据链路断点**：`retrySync`（执行增删）→ `syncDomainInternal`（发布事件）→ `StoreLogWriter`（写 DB）→ `GetSyncLogs` → 前端表格。

```
retrySync: 执行 diff.ToAdd / diff.ToDelete 后只 return err ── 计数在这里丢失 ✗
    ↓
syncDomainInternal: 发布 EventDomainSyncComplete，Data = {provider, domain} ── 无计数字段 ✗
    ↓
StoreLogWriter.OnEvent: 仅 log.Result = "success" ── 不读计数、不写 DB 的 added/deleted 列 ✗
    ↓
sync_logs.added / deleted 恒为默认值 0
```

三层各自独立缺失，任一层补上都无效，需**串通链路**（详见 §四）。

### 2.3 问题 8 根因：实时日志不一致且显示不全

三个独立原因叠加：

| 原因 | 位置 | 表现 |
|------|------|------|
| **① 格式不一致** | `logstream.go` L69-75 vs `slog.TextHandler` | stdout：`time=2026-08-02T10:00:00+08:00 level=INFO msg=同步完成 provider=...`；WebUI：`10:00:00 [INFO] 同步完成 provider=...`（无日期、无时区、属性无引号转义） |
| **② 无历史回放** | `handleLogStream` 订阅即增量 | 页面打开前的日志不显示；`docker compose logs` 显示全部 → 视觉上"少了一大截" |
| **③ 通道丢弃** | `logstream.go` L79-83（channel 64 满则跳过）；前端 L80（logLines 上限 200 shift 丢弃） | 高频日志（debug、同步过程多属性行）在页面端/浏览器端丢失 |

修复方案见 §五。

---

## 三、全局明暗主题切换（改进 1）

### 3.1 方案：Naive UI `darkTheme` + DB 持久化（用户已确认）

**零新依赖**：Naive UI 原生导出 `darkTheme`，`NConfigProvider` 动态切换即可。

### 3.2 存储

复用 `settings` 表新增键 `theme`（`light` / `dark`，默认 `light`）：

- **写入**：前端切换时 `PUT /api/settings`（`{theme: "dark"}`），现有批量保存机制直接覆盖；
- **读取**：页面加载时 `GET /api/settings`（现有端点已返回全部键，零改动）；
- **副作用**：`theme` 键不被 `LoadConfig()` 使用（`config.go` 不解析），写入后触发的热重载对同步引擎无任何影响；
- **兼容性**：老用户无 `theme` 键 → 前端 `settings.theme || 'light'` 兜底。

### 3.3 前端交互

- **入口**：侧边栏顶部（FWAlizer 标题行）右侧放置主题切换按钮（`NSwitch` 或图标按钮），全站可见、一处切换；
- **逻辑**（`App.vue`）：

```
App.vue:
  isDark = ref(settings.theme === 'dark')
  <NConfigProvider :theme="isDark ? darkTheme : null">
  toggle():
    1. isDark.value = !isDark.value          // 即时生效
    2. PUT /api/settings {theme: ...}        // 持久化（失败仅 message.warn，不回滚 UI）
```

- **暗色适配点**：
  - 运行日志 `<pre>`（Logs.vue）：**保持深色终端风格**（`#1e1e1e` 背景 + `#d4d4d4` 文字），明暗主题下均一致——终端块在暗色下与页面融合，在亮色下作为代码块视觉惯例保留，不做跟随；
  - `NTag`/`NButton`/`NDataTable` 等组件由 Naive UI 主题自动适配，无需逐页处理；
  - 页面内硬编码色值（如 Settings.vue `color: #999` 辅助文字）在暗色下可读性稍降，但 Naive UI 的 `text-color-3` 级灰字在暗色下仍可读，**不逐一排查**（符合不过度防御）。

### 3.4 验收要点

1. 切换后全站（含表格、弹窗、标签）即时切换，刷新后保持；
2. 重启应用后主题保持；
3. Docker 多浏览器访问同一实例时主题一致（DB 存储）。

---

## 四、问题 7 修复设计：历史记录计数（改进 7）

### 4.1 修复方案：打通计数链路（三级改动）

**改动 1 — `syncer/retry.go`：`retrySync` 返回实际执行计数**

```go
// retrySync 返回 (added, deleted, err)
// 计数规则：仅统计最终成功轮次中实际调用云 API 成功的数量；
// 幂等跳过不计（规则已存在/已不存在视为"无变更"，与 AGENTS.md 幂等语义一致）
func (s *Syncer) retrySync(p provider.Provider, rule config.DomainRule, resolved []dns.ResolvedIP) (added, deleted int, err error) {
    ...
    if len(diff.ToDelete) > 0 {
        if err := p.DeleteRules(diff.ToDelete); err != nil {
            if isIdempotentDelete(err) {
                slog.Warn("规则已不存在，跳过", ...)   // 不计 deleted
            } else { ... continue }
        } else {
            deleted += len(diff.ToDelete)              // 成功 → 计数
        }
    }
    if len(diff.ToAdd) > 0 {
        if err := p.CreateRules(diff.ToAdd); err != nil {
            if isIdempotentCreate(err) { ... }         // 不计 added
            else { ... continue }
        } else {
            added += len(diff.ToAdd)                   // 成功 → 计数
        }
    }
    return added, deleted, nil
}
```

**改动 2 — `syncer/syncer.go`：`syncDomainInternal` 将计数写入成功事件**

```go
added, deleted, err := s.retrySync(p, rule, resolved)
if err != nil { ... 发布 EventSyncError（不变）... }
s.bus.Publish(notifier.Event{
    Type:      notifier.EventDomainSyncComplete,
    Timestamp: time.Now(),
    Data:      map[string]any{"provider": p.Name(), "domain": rule.Host, "added": added, "deleted": deleted},
})
```

**改动 3 — `webui/api/logwriter.go`：`StoreLogWriter` 读取计数**

```go
case notifier.EventDomainSyncComplete:
    log.Result = "success"
    // 类型开关兼容 int（进程内传递）与 float64（若经 JSON 再入）
    if v, ok := event.Data["added"]; ok { log.Added = toInt(v) }
    if v, ok := event.Data["deleted"]; ok { log.Deleted = toInt(v) }
```

**无需改动**：`sync_logs` 表结构与 `GetSyncLogs`（`added`/`deleted`/`error` 列均已存在并已返回）。

### 4.2 设计说明

- **重试语义**：重试场景下只统计最后一次成功轮次的执行量（重试会重新 Describe → Diff，前几轮未写入的量不计入），符合"实际生效"语义；
- **幂等语义**：规则已存在/已不存在按 AGENTS.md 视为成功但不计数（count 反映"本次实际变更数"），Dry Run 的 `to_add`/`to_delete` 与历史 `added`/`deleted` 口径一致；
- **历史数据不回填**：旧记录（全 0）无法追溯，仅新记录生效，文档说明即可，不做迁移。

---

## 五、问题 8 修复设计：实时日志一致性（改进 8）

### 5.1 修复方案：环形缓冲回放 + 格式统一 + 通道扩容

**改动 1 — `webui/api/logstream.go`：环形缓冲 + 回放**

```go
// LogBroadcaster 增加：
//   ring  [1000]string   // 环形缓冲：最近 1000 条日志
//   ringPos int          // 写指针
//   ringCount int        // 已写条数

func (b *LogBroadcaster) Handle(...) {
    line := renderLine(r)          // 统一格式（见改动 2）
    // 1. 写入环形缓冲（无锁或独立小锁；回放时需快照）
    b.pushRing(line)
    // 2. 推送订阅者（channel 满则跳过，容量 256）
    ...
}

// Subscribe 订阅时先回放缓冲（顺序 = 时间正序），再进入增量模式
func (b *LogBroadcaster) Subscribe() (<-chan string, func()) {
    ch := make(chan string, 256)
    // 加锁快照环形缓冲 → 逐个写入 ch（不阻塞，回放为同步写，容量足够）
    ...
}
```

**改动 2 — 格式统一（推荐：复用 slog TextHandler 渲染）**

```go
// renderLine 用 slog.TextHandler 渲染单行，与 stdout 逐字符一致
// （TextHandler 输出：time=... level=INFO msg=... key=value）
var buf bytes.Buffer
h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: b.level})
h.Handle(context.Background(), r)
return buf.String()
```

**备选（放弃字符级一致）**：保持简洁格式但补全日期与时区（`2006-01-02 15:04:05 [INFO] msg`）。权衡：TextHandler 行较长但可精确对拍 `docker compose logs`；**推荐 TextHandler 渲染**（本设计采用）。

**改动 3 — 前端 `Logs.vue`：行数上限 200 → 1000**（与后端缓冲一致）。

**改动 4 — channel 容量 64 → 256**：正常一轮同步日志量远小于此，配合回放补偿，丢日志概率趋近于零。

### 5.2 效果对照

| 场景 | 修复前 | 修复后 |
|------|--------|--------|
| 打开日志页看历史日志 | 无（只显示订阅后） | 回放最近 1000 条，与终端输出衔接 |
| 与 `docker compose logs` 对比 | 格式不同、内容缺失 | 逐字符一致（时间戳/级别/属性格式同源） |
| 同步高峰丢日志 | channel 64 满则丢 + 前端 200 行截断 | 256 通道 + 1000 行缓冲 + 1000 行前端 |
| 页面刷新/短暂断线 | 断线期间日志永久丢失 | 刷新后重新回放最近 1000 条 |

---

## 六、仪表盘重设计（改进 2、3）

### 6.1 布局：方案 B（2×2 大卡片网格，用户已确认）

```
┌────────────────────────────────────────────────────────────┐
│  仪表盘                                                     │
│  ┌─────────────────────────┬──────────────────────────────┐ │
│  │ 卡片1：同步引擎          │ 卡片2：上次同步              │ │
│  │   [运行中] (大号标签)     │   2026-08-02 10:30:45        │ │
│  │   引擎状态大字           │   (上次同步时间, 大字)        │ │
│  └─────────────────────────┴──────────────────────────────┘ │
│  ┌─────────────────────────┬──────────────────────────────┐ │
│  │ 卡片3：统计概览          │ 卡片4：操作中心              │ │
│  │   云资源目标    N 个      │   [立即同步]  [暂停/开启]     │ │
│  │   域名规则      N 条      │   (大号按钮)                │ │
│  │   最近同步     +N / -M    │                              │ │
│  └─────────────────────────┴──────────────────────────────┘ │
└────────────────────────────────────────────────────────────┘
```

### 6.2 卡片内容与数据来源（全部复用现有端点，零新 API）

| 卡片 | 内容 | 数据来源 |
|------|------|---------|
| 1 同步引擎 | 三态大号 `NTag`（运行中/已暂停/已停止）+ 状态大字 | `GET /api/sync/status`（5s 轮询保留） |
| 2 上次同步 | 大字时间（无则「暂无」） | 同上 `last_sync` |
| 3 统计概览 | 目标数、规则数、最近同步 +N/-M | `GET /api/targets`、`GET /api/rules`（`length`）；`GET /api/sync/logs` 首条（`added`/`deleted`，修复后有效） |
| 4 操作中心 | 「立即同步」（暂停时置灰+tooltip）、「暂停/开启」 | 现有 pause/resume/trigger 端点 |

**视觉规范**：

- 卡片：`NCard` + 统一 `min-height`（约 200px）+ 加大内边距，四卡等高对称；
- 字体：卡片标题 16-18px，状态大字 24-28px，统计数字 24px+（`NStatistic` 或自定义）；
- 按钮：`size="large"`（原 small），操作区留白充足；
- 状态标签：`NTag size="large"` 圆角加大，文字 14-16px。

### 6.3 移除「运行测试」入口（改进 3）

- **现状**：`Dashboard.vue` L77-79 同步引擎卡片内的文本链接「运行测试 →」；
- **处理**：随 2×2 重设计**彻底移除**，仪表盘不再出现运行测试入口；「运行测试」仅保留左侧菜单栏唯一入口（`/run-test`）；
- **理由**：菜单栏入口已存在且语义清晰；用户明确要求移除仪表盘内链接，不另设按钮，避免双入口混淆。

### 6.4 验收要点

1. 四卡等高、空间利用率明显提升（对比原 3 小卡）；
2. 暂停状态下「立即同步」置灰 + hover 提示，与现状行为一致；
3. 仪表盘无任何「运行测试」入口；菜单栏入口正常跳转；
4. 统计概览数字随目标/规则增删自动刷新（进入页面时加载 + 5s 轮询刷新）。

---

## 七、同步日志页重排（改进 5、6）

### 7.1 页面结构（用户已确认：删除实时事件版块）

```
同步日志
├── 历史记录（最顶部）
│   ├── 标题行：[历史记录]  [清空记录] 按钮
│   └── 表格：时间 / 目标 / 域名 / 结果(failed可点击) / 新增 / 删除
└── 实时运行日志（默认展开）
    └── <pre> 终端样式日志流（回放 + 增量）
```

**改动清单（`Logs.vue`）**：

| 项 | 处理 |
|----|------|
| 「实时事件」表格（`eventColumns`/`events`/`formatEventData`/`eventTypeLabels`/`eventTagType`） | **删除** |
| `GET /api/sync/events` SSE 订阅（`es`） | **前端删除**（后端端点保留，向后兼容，见 §十） |
| 历史记录 | 移至页面最顶部（`h3` + 表格） |
| 运行日志 `NCollapse` | 居中放置；`:default-expanded-names="['logs']"` 默认展开（改进 6） |

### 7.2 信息覆盖论证（删除实时事件的安全性）

删除后信息不丢失：

| 原实时事件类型 | 信息去向 |
|---------------|---------|
| `sync:start`（N 目标 M 规则） | 运行日志首行「开始同步 targets= N rules= M」 |
| `domain:sync_complete`（成功） | 历史记录 success 行 + 运行日志「同步完成」 |
| `sync:error` / `dns:failed` | 历史记录 failed 行（error 详情见改进 9）+ 运行日志 WARN/ERROR 行 |

实时事件本质是历史记录与运行日志的**冗余摘要**，删除后信息完整、结构更简洁。

---

## 八、历史记录增强（改进 9、10）

### 8.1 failed 错误报告（改进 9）

**现状**：`SyncLog.Error` 已入库、`GetSyncLogs` 已返回，前端未使用。

**设计**：

- 前端类型补充：`SyncLogEntry` 增加 `error?: string`（后端 JSON 已带，仅类型声明补齐）；
- 结果列渲染：`failed` → `NTag type="error"`，加 `cursor: pointer` + hover 提示「点击查看错误详情」；`success`/其他不可点击；
- 点击 failed → 打开 `NModal`（preset="card"，标题「同步失败详情」）：

```
同步失败详情
  时间：2026-08-02 10:30:45 UTC+08:00
  目标：lhins-3j99jcrw
  域名：api.example.com
  错误原因：
  ┌────────────────────────────────┐
  │ （<pre> 包裹的 error 原文）     │
  └────────────────────────────────┘
```

- 老数据 `error` 为空 → 弹窗内提示「该记录未保存错误详情」；
- 实现形态：列 `render` 中 `h(NTag, { onClick })`，或新增「详情」操作列——**采用标签点击**（交互最直观，不增加列宽）。

### 8.2 清空记录按钮（改进 10）

**后端**：

```go
// config/store.go 新增
func (s *Store) ClearSyncLogs() error {
    _, err := s.db.Exec("DELETE FROM sync_logs")
    return err
}

// webui/api/sync.go 新增 handler + deps.go 路由注册
mux.HandleFunc("DELETE /api/sync/logs", d.handleClearSyncLogs)
```

**前端**：历史记录标题行右侧「清空记录」按钮（`size="small" type="error"`）+ `NPopconfirm` 二次确认（文案：「将清空全部同步历史记录，此操作不可恢复」），确认后 `DELETE /api/sync/logs` 并刷新表格。

**边界**：仅清 `sync_logs` 表，不影响 targets/rules/settings；清空与实时事件/日志流无关（事件总线、日志广播均为内存态，不受影响）。

---

## 九、首次引导与 Keys 提示（改进 11、12）

### 9.1 首次启动引导（改进 11）

**判定条件**（复用现有数据，零新端点）：`GET /api/settings` 返回的四个凭据键 `tc_access_id`、`tc_access_key`、`ali_access_id`、`ali_access_key` 全部为空 → 未配置凭据。

**设计**：

- 位置：**仪表盘顶部** `NAlert`（`type="warning"`，`bordered`），首屏可见、不打扰其他页面；
- 文案：「首次使用：请先在「全局设置」中填写云厂商 API 密钥（SecretId/SecretKey），再添加云资源目标与域名规则」；
- 操作：`[去配置]` 按钮（跳转 `/settings`）+ `[知道了]` 关闭按钮；
- 展示逻辑：
  - Dashboard 挂载时检查一次（可合并进现有 `fetchStatus` 轮询之外的独立 `loadSettings`）；
  - 凭据已配置 → 不显示（状态驱动，天然自愈）；
  - 用户关闭（`知道了`）→ 本次会话隐藏（前端 ref 置 false）；刷新后若仍未配置则再次出现（不引入"永久已读"状态，保持简单且不遗漏）；
- 与改进 12 共用同一份 settings 缓存（组件外抽 `useSettings.ts` composable 或模块级缓存，避免两处重复请求——符合简单轻量化，两个页面共用一个小 composable 即可）。

### 9.2 云资源管理 Keys 缺失提示（改进 12，用户已确认：仅提示不阻止）

**设计**（`Targets.vue`）：

- 挂载时 `GET /api/settings` 缓存凭据状态：`tcReady`（`tc_access_id`+`tc_access_key` 均非空）、`aliReady`；
- 添加/编辑弹窗内，`watch(form.cloud_type)`：所选平台对应凭据缺失时，在弹窗表单下方显示 `NAlert type="warning"`（或 `NText` 红色小字）：
  - `tc_lighthouse` / `tc_cvm` 且 `!tcReady` →「腾讯云凭据未配置，请先在「全局设置」中填写 SecretId/SecretKey，否则同步将失败」；
  - `ali_swas` / `ali_ecs` 且 `!aliReady` → 对应阿里云文案；
- **不阻止保存**（现有流程不变），凭据缺失仅提示；
- 可选增强：提示条内嵌「去设置」文本链接（`router.push('/settings')`）；
- 测试连接已有后端凭据空值快速失败（返回 `success:false` + 提示文案），保留不动——三层提示（表单内提醒 / 保存不阻止 / 测试连接后端拦截）各司其职。

---

## 十、API 变更汇总

| 端点 | 变更 | 说明 |
|------|------|------|
| `DELETE /api/sync/logs` | **新增** | 清空历史记录（改进 10） |
| `GET /api/settings` | 无 | `theme` 键自动返回；凭据字段已返回（改进 1、11、12 复用） |
| `PUT /api/settings` | 无 | `theme` 写入走现有机制 |
| `GET /api/logs/stream` | **行为增强** | 订阅时回放最近 1000 条；行格式与 stdout 一致（改进 8） |
| `GET /api/sync/events` | 无（保留） | 前端不再消费，端点保留向后兼容（改进 5） |
| `GET /api/sync/logs` | 无 | `error` 字段已返回（改进 9 前端消费） |
| 事件 `domain:sync_complete` | **数据字段扩展** | Data 增加 `added`/`deleted`（改进 7，兼容旧订阅者） |
| `POST /api/targets`、`PUT /api/targets/{id}` | 无 | 凭据检查在前端提示层，不阻止（改进 12） |

**Schema 变更：零。** `settings` 表新增键（`theme`）；`sync_logs` 表既有字段全部复用。

---

## 十一、兼容性分析

| 项 | 结论 |
|----|------|
| 主题 `theme` 键 | 老用户无该键 → 前端 `'light'` 兜底；`LoadConfig()` 不解析该键，热重载无副作用 |
| 事件 Data 扩展 | `EventDomainSyncComplete` 增加 `added`/`deleted`，旧订阅者（告警 Notifier 不订阅该类型）无感知 |
| `retrySync` 签名变更 | 返回 `(added, deleted int, err error)`，调用点仅 `syncDomainInternal` 一处（已确认无其他调用方） |
| 历史记录数据 | 表结构不变；旧记录计数为 0 不回填；`error` 为空时弹窗提示 |
| `/api/sync/events` 保留 | 前端停止消费，端点与 EventBus 机制不动，未来可复用 |
| 日志格式统一 | 纯展示层变更；`MultiHandler` 结构不变，stdout 输出不受影响 |
| 删除实时事件版块 | 纯前端删减；信息由历史记录 + 运行日志覆盖（见 7.2） |

---

## 十二、风险评估

| 风险项 | 等级 | 说明 | 缓解措施 |
|--------|------|------|---------|
| 环形缓冲内存占用 | 🟢 低 | 1000 条 × 约 200 字节 ≈ 200KB | 容量可配置常量，内部工具可接受 |
| 通道满仍丢日志 | 🟡 中 | 极端高频日志可能丢增量 | 容量 64→256 + 回放补偿；前端上限 1000 |
| 计数与重试/幂等交互 | 🟡 中 | 计数口径错误会误导用户 | 只统计最终成功轮次实际 API 成功量；幂等跳过不计，与 Dry Run 口径一致 |
| 主题写入失败 | 🟢 低 | `PUT /api/settings` 失败 | 仅 `message.warn`，不回滚 UI；下次进入页面重新从 DB 读取 |
| 首屏多请求 | 🟢 低 | Dashboard 新增 targets/rules/settings 请求 | 复用 5s 轮询周期合并加载，量小可忽略 |
| 删除实时事件后诊断降级 | 🟢 低 | 事件摘要信息冗余 | 历史记录 + 运行日志已覆盖（7.2） |

---

## 十三、实施计划

### 13.1 改动文件清单

**后端（3 个包，约 70 行）：**

| 文件 | 改动 | 说明 |
|------|------|------|
| `syncer/retry.go` | ~15 行 | `retrySync` 返回 `(added, deleted, err)`，成功分支计数 |
| `syncer/syncer.go` | ~4 行 | 成功事件 Data 增加 `added`/`deleted` |
| `webui/api/logwriter.go` | ~8 行 | 读取计数写入 DB（含 int/float64 类型开关） |
| `config/store.go` | ~5 行 | `ClearSyncLogs()` |
| `webui/api/sync.go` | ~10 行 | `handleClearSyncLogs` |
| `webui/api/deps.go` | +1 行 | 路由注册 `DELETE /api/sync/logs` |
| `webui/api/logstream.go` | ~40 行 | 环形缓冲 + 回放 + TextHandler 渲染 + 通道 256 |

**前端（7 个文件）：**

| 文件 | 改动 | 说明 |
|------|------|------|
| `App.vue` | ~25 行 | `darkTheme` 动态切换 + 侧边栏主题开关 + `GET /api/settings` 读主题 + PUT 持久化 |
| `views/Dashboard.vue` | 重写 | 2×2 大卡片（引擎状态/上次同步/统计概览/操作中心）+ 移除运行测试链接 + 首次引导 NAlert |
| `views/Rules.vue` | ~3 行 | `loadTargets` label 用 `cloudLabelMap` 转中文 |
| `views/Logs.vue` | ~50 行 | 删实时事件版块与事件 SSE；历史记录置顶；运行日志默认展开；failed 点击弹窗；清空按钮；logLines 上限 1000 |
| `views/Targets.vue` | ~20 行 | settings 缓存 + `watch(cloud_type)` 凭据缺失提示条 |
| `types.ts` | ~3 行 | `SyncLogEntry.error?` + 主题相关类型（如需） |
| `composables/useSettings.ts`（新增） | ~25 行 | 凭据状态/主题读取共享逻辑（Dashboard 与 Targets 复用） |

**文档**：本文件（Design3.md）为设计构想记录；详细构建方案见 [Build4.md](./Build4.md)。

### 13.2 实施顺序建议

| 阶段 | 内容 | 依赖 |
|------|------|------|
| **P1 数据修复** | 计数链路（retry.go → syncer.go → logwriter.go）+ 日志广播增强（环形缓冲/格式/回放） | 无 |
| **P2 日志页** | 历史记录置顶、删实时事件、默认展开、failed 弹窗、清空按钮（前端） | P1 的计数/回放生效后验收 |
| **P3 主题** | `useSettings.ts` + App.vue 主题切换 + 暗色适配 | 无 |
| **P4 仪表盘** | 2×2 重设计 + 移除运行测试链接 + 首次引导 | P3（useSettings 复用） |
| **P5 细节** | Rules.vue 中文云产品名 + Targets.vue Keys 提示 | P3（useSettings 复用） |

### 13.3 验收标准（对应 12 项改进）

1. **主题**：切换即时生效、刷新/重启保持、Docker 多端一致；
2. **仪表盘**：2×2 等高大卡片、大字体大按钮、统计概览正确（目标数/规则数/最近同步 ±N）；
3. **运行测试入口**：仪表盘无该入口，菜单栏 `/run-test` 正常；
4. **中文云产品名**：规则表「适用目标」显示「腾讯云轻量云 / lhins-xxx」格式，与云资源管理列一致；
5. **日志页结构**：历史记录最顶部，无实时事件版块；
6. **默认展开**：进入日志页运行日志即展开；
7. **计数**：完成一轮同步后，历史记录「新增/删除」显示实际变更数（幂等跳过不计）；
8. **日志一致性**：WebUI 实时日志与 `docker compose logs` 同格式；打开页面即可看到最近日志（回放）；
9. **错误报告**：failed 标签可点击，弹窗展示时间/目标/域名/error 原文；
10. **清空**：确认后历史记录清空，其他数据不受影响；
11. **首次引导**：无凭据时仪表盘顶部出现引导条，「去配置」跳转设置页；配置凭据后消失；
12. **Keys 提示**：目标弹窗选择未配凭据平台时显示提示条，保存不被阻止；测试连接仍返回后端快速失败提示。

### 13.4 可选附项（不属本次 12 项，低优先）

| # | 项 | 说明 |
|---|----|------|
| A | DNS 失败写入历史记录 | 现 `EventDNSFailed` 不落库；可将其写入 `sync_logs`（`result=failed`、`error`=DNS 错误、`domain` 已含），提升故障可见性——需在 `main.go` 增加 `EventDNSFailed` 订阅 |
| B | 日志页「暂停输出」按钮 | 高日志量时前端可暂停渲染（性能考虑），内部工具场景可选 |
| C | 主题设置项加入全局设置页 | 目前主题仅侧边栏切换；如需在设置页展示/管理，复用同一 `theme` 键即可 |

---

## 十四、WebUI 细节改进研究（后续改进，用户已确认决策）

> 本节为 Build4 构建完成后追加的 4 项 WebUI 细节改进研究（2026-08-02），均已完成现状确认与方案取舍确认（用户决策项以 **✅ 已确认** 标注），待后续 Build 文档承接实施。**实施前不修改任何代码。**

### 14.1 改进 A：仪表盘·同步引擎状态显示去重

#### 现状确认

- **DOM 位置**：`Dashboard.vue` 同步引擎卡片（Build4 Step 6 产出）：

```
NCard「同步引擎」
├── 大字 28px：{{ engineTag.text }}（运行中/已暂停/已停止，无背景色）
├── NTag size="large" :type="engineTag.type"（带颜色背景 success/warning/default）← 重复
└── 灰色说明文字 13px（辅助说明，非状态，保留）
```

- **数据来源**：`engineTag` computed（`status.running`/`status.enabled` 三态 → `type: success/warning/default`）；
- **渲染条件**：无条件渲染，两处均输出同一 `engineTag.text` → 状态文本重复显示两次。

#### 推荐方案（✅ 已确认：移除 NTag + 大字上色）

- 移除 `NTag` 元素，保留 28px 大字作为**唯一状态展示**；
- `engineTag` computed 扩展返回 `{ text, color }`：运行中 `#18a058`（绿）、已暂停 `#f0a020`（橙）、已停止 `#808080`（灰），大字通过 `:style="{ color: engineTag.color }"` 上色；
- 颜色硬编码（与 naive-ui 主题色近似），明暗主题下均可读，不引入 `useThemeVars`（符合不过度防御）；
- 灰色说明文字（第三行）保留不动。

#### 影响范围

- 仅 `Dashboard.vue`：`engineTag` computed 扩展 + 同步引擎卡片模板；
- 不影响「操作中心」卡片（暂停/开启开关、立即同步按钮）与「统计概览」卡片；
- 三态判断逻辑（`running`/`enabled`）不变。

#### 备选方案（已对比，未采纳）

| 备选 | 说明 | 未采纳理由 |
|------|------|-----------|
| 保留 NTag 移除大字 | NTag 自带颜色背景语义 | 用户已确认移除带色小标签、保留大字主状态 |
| 两处均保留但去背景 | NTag 改 text 型（无背景） | 仍属重复展示，不满足去重目标 |

#### 风险

| 风险 | 等级 | 缓解 |
|------|------|------|
| 大字硬编码色值在暗色主题下可读性 | 🟢 低 | 三个色值在明暗背景均可达对比度要求（内部工具从简，不引入主题变量） |
| 移除 NTag 后状态辨识度下降 | 🟢 低 | 大字 28px + 状态色补偿，辨识度不降反升 |

---

### 14.2 改进 B：全站 "Dry Run" 更名为「模拟测试」+ 说明文字

#### 现状确认（全站 grep 梳理）

**前端用户可见文案（4 处，需更名）：**

| 位置 | 现文案 | 更名后 |
|------|--------|--------|
| `RunTest.vue` L67 | `<NTabPane name="dryrun" tab="Dry Run">` | `tab="模拟测试"` |
| `RunTest.vue` L70 | 按钮「执行 Dry Run」 | 「执行模拟测试」 |
| `RunTest.vue` L22/L24 | message「Dry Run 完成 / Dry Run 失败」 | 「模拟测试完成 / 模拟测试失败」 |
| `DryRunResults.vue` L61 | 「尚未执行 Dry Run，点击上方「执行 Dry Run」开始」 | 「尚未执行模拟测试，点击上方「执行模拟测试」开始」 |
| `Dashboard.vue` L115 | 「…Dry Run 与连接测试在「运行测试」页使用」 | 「…模拟测试与连接测试在「运行测试」页使用」 |

**无需更改（非用户可见）：**

- API 端点 `POST /api/sync/dryrun`（内部契约，保留）；
- 代码标识符与内部值：`DryRunResult`/`DryRunResponse`（types.ts）、`useDryRun.ts`、`DryRunResults.vue` 组件名、`runDryRun()` 函数名、`route.query.tab='dryrun'` 与 `name="dryrun"`；
- 侧边栏菜单「运行测试」（页面名本身无 Dry Run 字样）。

#### 推荐方案（✅ 已确认：前端 UI + 活跃文档；说明文字放标签页副标题）

1. **前端更名**：按上表 5 处用户可见文案统一替换为「模拟测试」；
2. **说明文字（标签页副标题）**：「模拟测试」Tab 内、操作区上方常驻一行小字：

```
「模拟测试仅生成变更预览，不实际写入云防火墙规则」
（实现：font-size 12px、color #888、margin-bottom 12px）
```

3. **活跃文档同步更名**（仅用户面向描述，技术概念描述保留）：

| 文档 | 位置 | 处理 |
|------|------|------|
| `README.md` | L85（SYNC_ENABLED 说明）、L145（运行测试页描述） | Dry Run → 模拟测试 |
| `AGENTS.md` | L93、L154（同步开关/运行测试页描述） | Dry Run → 模拟测试（API 名 `DryRunResponse` 保留） |
| `Build4.md` | UI 参考代码中的用户文案（如 Dashboard 说明文字） | 与前端代码同步 |
| `Design3.md` | 本文档 | 本节省略技术性 Dry Run 引用；若后续实施，相关描述同步 |
| `Build1-3`/`Issue1-3`/`Design1-2` | 历史归档 | **保持原样**（✅ 已确认，避免历史失真） |

#### 影响范围

- 前端 3 文件（`RunTest.vue`/`DryRunResults.vue`/`Dashboard.vue`）；
- 活跃文档 3 份（README/AGENTS/Build4）；
- 后端零改动；API、路由、组件名、类型名全部保留。

#### 备选方案（已对比，未采纳）

| 备选 | 说明 | 未采纳理由 |
|------|------|-----------|
| 仅改前端 UI | 文档与 UI 不一致 | 用户已确认同步活跃文档 |
| 全部文档统一更名 | 含历史归档 | 历史失真、改动量大 |

#### 风险

| 风险 | 等级 | 缓解 |
|------|------|------|
| 更名遗漏（grep 不全） | 🟢 低 | 实施前以 grep "Dry Run|DryRun" 全量核对；API/标识符刻意保留处加注释说明 |
| 文档与代码术语不一致 | 🟢 低 | 文档保留 `DryRunResponse` 等 API 名（与代码一致），仅用户面向描述更名 |

---

### 14.3 改进 C：主题切换按钮可读性增强

#### 现状确认

- `App.vue` 侧边栏标题行右侧：纯 `NSwitch`（无文字、无图标、无提示），用户难以辨识其功能；
- naive-ui 不内置图标集（项目不引入新依赖）。

#### 推荐方案（✅ 已确认：图标 + tooltip）

- 开关左侧显示随主题切换的 emoji 图标：`theme === 'dark' ? '🌙' : '☀️'`（零依赖 Unicode）；
- 整体包裹 `NTooltip`，hover 提示「切换明暗主题」；
- 布局整合：标题行 `display: flex; justify-content: space-between; align-items: center` 已存在，图标 + 开关以内联 span 组合（`display:flex; gap:4px`）替换原 NSwitch，不改变标题行结构；
- 图标与开关同步切换（`:value="theme === 'dark'"` 同一数据源）。

```
FWAlizer        ☀️ [开关]   ← 亮色时
FWAlizer        🌙 [开关]   ← 暗色时
（hover 任意元素显示「切换明暗主题」）
```

#### 影响范围

- 仅 `App.vue` 侧边栏标题行；
- 主题状态与持久化逻辑（`useSettings.setTheme`）不变。

#### 备选方案（已对比，未采纳）

| 方案 | 说明 | 未采纳理由 |
|------|------|-----------|
| ① 文字标签「深色/浅色」 | 最明确 | 侧边栏 200px 宽度占用较大 |
| ③ 仅 hover 提示 | 改动最小 | 无悬停时仍无辨识度 |
| ① 图标方案（✅ 已选） | 🌙/☀️ emoji + tooltip | 直观、零依赖、占用小 |

#### 风险

| 风险 | 等级 | 缓解 |
|------|------|------|
| emoji 跨平台渲染差异 | 🟢 低 | ☀/🌙 为 Unicode 标准字符，主流系统均支持 |
| tooltip 在窄侧边栏遮挡 | 🟢 低 | NTooltip 默认向上/向下弹出，不影响布局 |

---

### 14.4 改进 D：移除实时运行日志的折叠/展开按钮

#### 现状确认

- `Logs.vue` L105-110（Build4 Step 4 产出）：`NCollapse :default-expanded-names="['logs']"` + `NCollapseItem title="运行日志（实时）" name="logs"` 包裹 `<pre>` 日志面板；
- 当前默认展开（Build4 已设），但折叠按钮仍可收起日志区，用户需手动保持展开。

#### 推荐方案：移除折叠控件，日志区常驻平铺

```vue
<!-- 原 NCollapse/NCollapseItem 包装移除，替换为： -->
<h3 style="margin-top: 16px">运行日志（实时）</h3>
<pre style="max-height: 300px; overflow-y: auto; background: #1e1e1e; color: #d4d4d4; ...">{{ logLines.join('\n') || '等待日志输出...' }}</pre>
```

- 标题样式与「历史记录」h3 对齐（统一视觉）；
- 移除 `NCollapse`/`NCollapseItem` import（避免未使用告警）。

#### 影响评估

| 维度 | 评估 |
|------|------|
| 页面布局 | 日志区常驻占位（`max-height: 300px` + 内部滚动），页面整体变长但可滚动，不影响历史记录表格 |
| 长时间运行 | `logLines` 上限 1000 行（内存受控）+ pre 内部滚动，行数持续增长不影响布局；日志始终可见（用户明确需求） |
| 明暗主题 | pre 保持深色终端风格，不随主题变化（与 Build4 一致） |
| SSE 逻辑 | 不变（`/api/logs/stream` 订阅与回放逻辑零改动） |

#### 影响范围

- 仅 `Logs.vue` 模板（NCollapse 包装移除）+ import 清理；
- 后端零改动。

#### 备选方案（已对比）

| 备选 | 说明 | 未采纳理由 |
|------|------|-----------|
| 保留折叠但默认展开 | 现状 | 不满足「移除折叠按钮」需求 |
| 常驻展开 + 「暂停输出」按钮 | 高日志量时手动暂停渲染 | 超出本次范围，列为可选附项（同 §13.4-B） |

#### 风险

| 风险 | 等级 | 缓解 |
|------|------|------|
| 页面纵向空间增加 | 🟢 低 | 300px 固定高度 + 滚动，其他区块不受挤压 |
| 日志量超前端渲染性能 | 🟢 低 | 1000 行上限 + 虚拟滚动需求低（内部工具），可接受 |

---

### 14.5 汇总（后续实施建议）

| 改进 | 涉及文件 | 依赖 | 备注 |
|------|---------|------|------|
| A 状态去重 | `Dashboard.vue` | 无 | 纯前端微调 |
| B Dry Run 更名 | `RunTest.vue`/`DryRunResults.vue`/`Dashboard.vue`/README/AGENTS/Build4 | 无 | 说明文字副标题 |
| C 主题按钮 | `App.vue` | 无 | emoji + tooltip |
| D 日志常驻 | `Logs.vue` | 无 | 移除 NCollapse |

四项均无相互依赖、无后端改动，可合并为一次小型前端构建（后续 Build 文档承接）；实施后统一执行 `npm run build` 验收。
