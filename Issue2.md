# FWAlizer WebUI 设计/功能问题记录

> 本文档记录 WebUI 层面的设计优化与功能缺陷。
> 问题按编号顺序排列，末尾附汇总表与优先修复建议。

---

## 1. 云资源管理页

### [UI-01] 云产品列显示原始标识符而非中文名称

- **严重度：** 低
- **当前状态：** ✅ 已修复
- **所属模块：** WebUI 前端 / 云资源管理
- **涉及文件：** `webui/frontend/src/views/Targets.vue`

**现象描述：** 表格"云产品"列直接显示 `cloud_type` 原始值（如 `tc_lighthouse`、`ali_swas`），对用户不友好。

**原因分析：** `Targets.vue` L70 列定义为 `{ title: '云产品', key: 'cloud_type' }`，使用 NDataTable 默认的 key 直出渲染，未提供 `render` 函数做映射。而同文件 L12-17 的 `cloudOptions` 已定义了完整的 label↔value 映射表（用于新增/编辑弹窗的 NSelect），但表格列未复用。

**影响范围：** 仅影响可读性，不影响功能。

**推荐修复方案：** 为"云产品"列添加 `render` 函数，复用 `cloudOptions` 映射：

```typescript
// Targets.vue columns 定义中
const cloudLabelMap: Record<string, string> = Object.fromEntries(
  cloudOptions.map(o => [o.value, o.label])
)

const columns = [
  { title: '#', key: 'index', render: (_: any, i: number) => i + 1 },
  {
    title: '云产品', key: 'cloud_type',
    render(row: any) {
      return cloudLabelMap[row.cloud_type] || row.cloud_type
    }
  },
  // ... 其余列不变
]
```

---

## 2. 全局设置页

### [UI-02] DNS 服务器默认值应改为国内公共 DNS 并自动补全端口

- **严重度：** 中
- **当前状态：** ✅ 已修复
- **所属模块：** 配置 / WebUI 前端
- **涉及文件：** `config/store.go`（L331）、`config/env.go`（L30）、`dns/resolver.go`（L32-34）、`webui/frontend/src/views/Settings.vue`（L79）

**现象描述：**
1. 默认 DNS 服务器为 `8.8.8.8:53`（Google DNS），国内环境访问不稳定，应改为 `223.5.5.5`（阿里公共 DNS）
2. 前端 placeholder 提示 `8.8.8.8:53`，要求用户手动输入端口号，增加使用门槛

**原因分析：**
- `config/store.go` L331：`DNS: "8.8.8.8:53"` 硬编码默认值
- `config/env.go` L30：`getOr(kv, "DNS", "8.8.8.8:53")` 硬编码默认值
- `dns/resolver.go` L32-34：`NewResolver` 已有端口自动补全逻辑（`if !hasPort(dnsAddr) { dnsAddr += ":53" }`），但前端和文档未利用此能力
- `Settings.vue` L79：placeholder 为 `"8.8.8.8:53"`，暗示用户必须带端口

**影响范围：** 国内用户首次使用时 DNS 解析可能超时或失败；用户需额外记忆端口号格式。

**推荐修复方案：**

1. **默认值变更**（`store.go` L331 + `env.go` L30）：
```go
// store.go LoadConfig()
DNS: "223.5.5.5",

// env.go ParseEnv()
DNS: getOr(kv, "DNS", "223.5.5.5"),
```

2. **前端 placeholder 更新**（`Settings.vue` L79）：
```vue
<NInput v-model:value="settings.dns" placeholder="223.5.5.5" />
```

3. **后端无需额外改动**：`dns/resolver.go` L33 的 `hasPort()` 检查已能自动为纯 IP 输入补全 `:53`，`LoadConfig()` 和 `ParseEnv()` 存储的值无需带端口。

4. **`.env.example` 同步更新**（L50）：
```
DNS=223.5.5.5
```

---

### [UI-03] 设置表单无默认值预填，用户无法看到当前生效配置

- **严重度：** 中
- **当前状态：** ✅ 已修复
- **所属模块：** WebUI 前端 + API
- **涉及文件：** `webui/frontend/src/views/Settings.vue`（L8-10, L72-83）、`webui/api/settings.go`（`handleGetSettings`）、`config/store.go`（`GetSettings`）

**现象描述：** TAG、同步间隔、DNS 服务器、日志级别四个字段仅有 `placeholder` 提示（灰色占位文本），输入框实际为空。用户打开设置页面无法区分"当前使用默认值"和"尚未配置"，也无法看到当前生效值。

**原因分析：**
- `GET /api/settings` → `Store.GetSettings()`（store.go L99-115）仅返回 settings 表中**已显式存储**的 key-value 对
- 首次使用时 settings 表为空 → API 返回 `{}` → 前端 `settings.tag`、`settings.interval` 等均为 `undefined` → 输入框为空（仅显示 placeholder）
- 实际生效的默认值（`auto-dns`、`5m`、`223.5.5.5`、`info`）定义在 `LoadConfig()` 中（store.go L329-334），但 `GetSettings()` 不返回这些默认值

**影响范围：** 用户体验差——无法确认当前配置状态；可能误以为未配置而重复填写。

**推荐修复方案：**

**方案 A（推荐）：后端 `GET /api/settings` 返回合并后的有效配置**

在 `handleGetSettings` 中将默认值与已存储值合并后返回：

```go
func (d *Deps) handleGetSettings(w http.ResponseWriter, r *http.Request) {
    settings, err := d.Store.GetSettings()
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    // 填充默认值（仅当 key 不存在时）
    defaults := map[string]string{
        "tag":       "auto-dns",
        "interval":  "5m",
        "dns":       "223.5.5.5",
        "log_level": "info",
    }
    for k, v := range defaults {
        if _, exists := settings[k]; !exists {
            settings[k] = v
        }
    }
    writeJSON(w, http.StatusOK, settings)
}
```

前端无需改动——`v-model:value="settings.tag"` 自动显示返回的值。

**方案 B：前端 onMounted 时填充默认值**（不推荐，默认值分散在前后端两处维护）

---

### [UI-04] 日志级别应改为下拉选择组件

- **严重度：** 低
- **当前状态：** ✅ 已修复
- **所属模块：** WebUI 前端
- **涉及文件：** `webui/frontend/src/views/Settings.vue`（L81-83）

**现象描述：** 日志级别（`log_level`）当前为 `NInput` 文本输入框，用户可输入任意字符串。若输入无效值（如 `verbose`、`INFO`），后端 `app.InitLogger()` 会 fallback 到默认级别但不报错，用户不知情。

**原因分析：** `Settings.vue` L81-83 使用 `<NInput v-model:value="settings.log_level" placeholder="info" />`，无输入约束。有效值仅为 `debug` / `info` / `warn` / `error` 四种（`app.InitLogger` 中的 switch 分支）。

**影响范围：** 用户可能输入无效值导致日志级别不符合预期。

**推荐修复方案：**

```vue
<!-- Settings.vue：替换 NInput 为 NSelect -->
<NFormItem label="日志级别">
  <NSelect v-model:value="settings.log_level" :options="[
    { label: 'Debug', value: 'debug' },
    { label: 'Info', value: 'info' },
    { label: 'Warn', value: 'warn' },
    { label: 'Error', value: 'error' },
  ]" />
</NFormItem>
```

需在 import 中增加 `NSelect`。

---

## 3. 同步日志页

### [UI-05] 时间戳格式不可读

- **严重度：** 中
- **当前状态：** ✅ 已修复
- **所属模块：** WebUI 前端 / 同步日志
- **涉及文件：** `webui/frontend/src/views/Logs.vue`（L29, L45）、`config/store.go`（SyncLog 结构体 L20）

**现象描述：**
- 历史记录"时间"列显示 Go `time.Time` 的 JSON 默认序列化格式：`2026-07-26T19:51:37.263852+08:00`（ISO 8601 带纳秒精度和时区偏移）
- 实时事件"时间"列同样显示原始 ISO 8601 格式
- 应格式化为 `YYYY-MM-DD HH:mm:ss`，使用 UTC 时间（或明确标注时区）

**原因分析：**
- `Logs.vue` L29：历史记录列 `{ title: '时间', key: 'timestamp' }` 无 render 函数，直出原始字符串
- `Logs.vue` L45：实时事件列 `{ title: '时间', key: 'timestamp' }` 同样无 render 函数
- 后端 `SyncLog.Timestamp` 为 `time.Time` 类型（store.go L20），JSON 序列化为 RFC 3339 格式
- SSE 事件 `Event.Timestamp` 同为 `time.Time`（bus.go L23），序列化格式一致

**影响范围：** 时间信息难以快速阅读，尤其纳秒精度和时区偏移对运维场景无意义。

**推荐修复方案：**

前端统一添加时间格式化 render 函数：

```typescript
// Logs.vue
function formatTime(ts: string): string {
  if (!ts) return '-'
  const d = new Date(ts)
  return d.toISOString().replace('T', ' ').substring(0, 19) + ' UTC'
}

const columns = [
  { title: '时间 (UTC)', key: 'timestamp', render: (row: any) => formatTime(row.timestamp) },
  // ... 其余列
]

const eventColumns = [
  { title: '时间 (UTC)', key: 'timestamp', render: (row: any) => formatTime(row.timestamp) },
  // ... 其余列
]
```

---

### [UI-06] 历史记录中“目标”和“域名”字段始终为空

- **严重度：** 高
- **当前状态：** ✅ 已修复
- **所属模块：** Syncer 事件发布 / LogWriter / 同步日志
- **涉及文件：** `syncer/syncer.go`（L216-220）、`webui/api/logwriter.go`（L16-37）、`notifier/bus.go`（L21-25）

**现象描述：** `GET /api/sync/logs` 返回的日志条目中 `target` 和 `domain` 字段始终为空字符串，"历史记录"表格的"目标"和"域名"两列永远无数据。

**原因分析（完整数据流追踪）：**

1. **事件发布端**（`syncer.go`）：
   - `sync:complete` 事件（L216-220）：`Data = {"duration": "..."}`——**不包含 `provider` 和 `domain` key**。这是一个全局级别的"本轮同步全部完成"事件，不携带单域名信息。
   - `sync:error` 事件（L255-259）：`Data = {"provider": p.Name(), "domain": rule.Host, "error": ...}`——**包含两个 key**。

2. **事件订阅端**（`logwriter.go` L16-37）：
   - 订阅了 `EventSyncComplete` 和 `EventSyncError` 两种事件
   - L18-22：尝试从 `event.Data["provider"]` 和 `event.Data["domain"]` 提取值
   - 对于 `sync:complete`：Data 中无这两个 key → 类型断言失败 → `log.Target` 和 `log.Domain` 保持空字符串
   - 对于 `sync:error`：Data 中有这两个 key → 正确提取

3. **结论**：每轮同步成功时写入的日志（占绝大多数）target/domain 必然为空；仅同步失败时才有值。

**影响范围：** 同步日志功能几乎无用——用户无法从历史记录中得知哪个目标、哪个域名被同步。

**推荐修复方案：**

在 `syncDomain()` 成功路径发布**逐域名事件**，携带 provider + domain + added/deleted 计数：

```go
// syncer.go syncDomain() 成功路径（L262 附近）
slog.Info("同步完成", "provider", p.Name(), "domain", rule.Host)
s.bus.Publish(notifier.Event{
    Type:      notifier.EventSyncComplete,
    Timestamp: time.Now(),
    Data:      map[string]any{"provider": p.Name(), "domain": rule.Host},
})
```

同时将 `syncAll()` L216-220 的全局完成事件改为 `EventSyncStart` 的配对事件（或新增 `EventSyncAllComplete` 类型），避免 logwriter 为全局事件写入空日志：

```go
// logwriter.go：仅处理携带 provider 的事件
case notifier.EventSyncComplete:
    if log.Target == "" {
        return nil // 全局完成事件，不写入日志
    }
    log.Result = "success"
```

---

### [UI-07] 实时事件文案不可读

- **严重度：** 中
- **当前状态：** ✅ 已修复
- **所属模块：** WebUI 前端 / 同步日志
- **涉及文件：** `webui/frontend/src/views/Logs.vue`（L44-48）、`webui/api/sync.go`（L61-76）、`notifier/bus.go`（L12-18）

**现象描述：** 实时事件表格"类型"列直接显示原始事件类型字符串（`sync:start`、`sync:complete`、`sync:error`、`dns:failed`），"详情"列显示 `JSON.stringify(row.data)` 原始 JSON（如 `{"targets":3,"rules":5}`），对非开发用户完全不可读。

**原因分析：**
- `Logs.vue` L46：`{ title: '类型', key: 'type' }` 无 render 函数
- `Logs.vue` L47：`{ title: '详情', key: 'data', render: (row: any) => JSON.stringify(row.data) }` 直接序列化
- SSE 端点（`sync.go` L67）将 `notifier.Event` 原样 JSON 序列化推送，不做文案转换
- 事件类型定义（`bus.go` L13-17）为技术标识符，非人类可读文案

**影响范围：** 实时事件区域对普通用户无实际价值，仅开发者可解读。

**推荐修复方案：**

前端添加事件类型映射和智能详情渲染：

```typescript
// Logs.vue
const eventTypeLabels: Record<string, string> = {
  'sync:start': '同步开始',
  'sync:complete': '同步完成',
  'sync:error': '同步失败',
  'dns:failed': 'DNS解析失败',
  'rule:changed': '规则变更',
}

function formatEventData(row: any): string {
  const d = row.data || {}
  switch (row.type) {
    case 'sync:start':
      return `${d.targets ?? 0} 个目标，${d.rules ?? 0} 条规则`
    case 'sync:complete':
      return d.domain ? `${d.provider} / ${d.domain}` : `耗时 ${d.duration ?? '-'}`
    case 'sync:error':
      return `${d.provider} / ${d.domain}：${d.error ?? '未知错误'}`
    case 'dns:failed':
      return `${d.domain}：${d.error ?? '解析超时'}`
    default:
      return JSON.stringify(d)
  }
}

const eventColumns = [
  { title: '时间 (UTC)', key: 'timestamp', render: (row: any) => formatTime(row.timestamp) },
  {
    title: '事件', key: 'type',
    render(row: any) {
      return h(NTag, {
        size: 'small',
        type: row.type === 'sync:error' || row.type === 'dns:failed' ? 'error'
            : row.type === 'sync:complete' ? 'success' : 'info'
      }, { default: () => eventTypeLabels[row.type] || row.type })
    }
  },
  { title: '详情', key: 'data', render: (row: any) => formatEventData(row) },
]
```

---

### [UI-08] 缺少实时日志流输出模块

- **严重度：** 中
- **当前状态：** ✅ 已修复
- **所属模块：** WebUI 前端 + 后端
- **涉及文件：** `webui/server.go`、`webui/api/`（新增端点）、`webui/frontend/src/views/Logs.vue`

**现象描述：** 同步日志页面仅有"实时事件"（结构化事件）和"历史记录"（SQLite 日志），缺少后端 `slog` 日志的实时流式输出。用户排查问题时需切换到终端查看 `docker logs` 或进程 stdout，不便。

**原因分析：** 当前架构中 `slog` 日志输出到 stdout（AGENTS.md 规定），未接入 WebUI 推送通道。EventBus 仅推送结构化业务事件（5 种 EventType），不覆盖 Debug/Info 级别的运行日志。

**影响范围：** 用户排查同步异常时需离开 WebUI 查看终端日志，体验割裂。

**推荐修复方案：**

**方案 A（推荐）：SSE + 自定义 slog.Handler**

理由：项目已有 SSE 基础设施（`/api/sync/events`），复用相同模式最简单；符合"简单轻量化"原则；无需引入 WebSocket 依赖。

```go
// 新增 webui/api/logstream.go
type LogBroadcaster struct {
    mu   sync.RWMutex
    subs map[int]chan []byte
    next int
}

// slog.Handler 实现：将日志格式化后广播到所有订阅 channel
func (b *LogBroadcaster) Handle(ctx context.Context, r slog.Record) error {
    line := fmt.Sprintf("%s [%s] %s\n",
        r.Time.Format("15:04:05"), r.Level, r.Message)
    b.mu.RLock()
    defer b.mu.RUnlock()
    for _, ch := range b.subs {
        select {
        case ch <- []byte(line):
        default: // 满则跳过
        }
    }
    return nil
}

// SSE 端点
func (d *Deps) handleLogStream(w http.ResponseWriter, r *http.Request) {
    // 类似 handleSyncEvents 的 SSE 模式
    // Content-Type: text/event-stream
    // 循环读取订阅 channel，写入 "data: ...\n\n"
}
```

前端在 Logs.vue 新增"运行日志"区域（可折叠面板），使用 `EventSource('/api/logs/stream')` 接收并追加到 `<pre>` 块中，限制最多显示 200 行。

**方案 B（不推荐）：WebSocket**

需引入额外依赖（`gorilla/websocket` 或 `nhooyr.io/websocket`），增加复杂度；对于单向日志流场景，SSE 已完全满足需求且更简单。

---

## 4. 启动日志

### [UI-09] 启动日志应显示完整访问地址

- **严重度：** 低
- **当前状态：** ✅ 已修复
- **所属模块：** WebUI 服务器
- **涉及文件：** `webui/server.go`（L47）

**现象描述：** 程序启动后 console 输出为：
```
msg="WebUI 启动" addr=127.0.0.1:9090
```
用户需自行拼接 `http://` 前缀才能访问。应改为：
```
msg="WebUI 启动" 访问地址=http://127.0.0.1:9090
```

**原因分析：** `server.go` L46-47：
```go
addr := fmt.Sprintf("127.0.0.1:%d", s.port)
slog.Info("WebUI 启动", "addr", addr)
```
slog 的 key 为 `addr`，值为纯 `host:port` 格式，不含协议前缀，且 key 名称不够直观。

**影响范围：** 仅影响启动日志可读性。

**推荐修复方案：**

```go
// server.go Start()
addr := fmt.Sprintf("127.0.0.1:%d", s.port)
slog.Info("WebUI 启动", "访问地址", "http://"+addr)
return http.ListenAndServe(addr, s.mux)
```

---

# 汇总表

## 按严重度统计

| 严重度 | 数量 | 编号 |
|--------|------|------|
| 高 | 1 | [UI-06] |
| 中 | 5 | [UI-02]、[UI-03]、[UI-05]、[UI-07]、[UI-08] |
| 低 | 3 | [UI-01]、[UI-04]、[UI-09] |
| **合计** | **9** | |

## 优先修复建议

1. **[UI-06] 同步日志 target/domain 为空**（最高优先级）：这是功能性缺陷，同步日志核心字段无数据导致功能形同虚设。修复需在 syncer 成功路径发布逐域名事件 + logwriter 过滤全局事件，改动集中在 `syncer.go` 和 `logwriter.go` 两个文件。

2. **[UI-03] 设置表单无默认值预填 + [UI-02] DNS 默认值**（高优先级）：两者可合并修复——`handleGetSettings` 返回合并默认值时一并将 DNS 默认值改为 `223.5.5.5`。改动集中在 `settings.go` + `store.go` + `env.go`。

3. **[UI-05] 时间格式化 + [UI-07] 事件文案**（中优先级）：两者均为 Logs.vue 前端渲染层改动，可在一次变更中完成，不涉及后端。

4. **[UI-01] 云产品中文映射 + [UI-04] 日志级别下拉 + [UI-09] 启动日志**（低优先级）：三项均为单文件小改动，可批量处理。

5. **[UI-08] 实时日志流**（待规划）：涉及新增后端 Handler + SSE 端点 + 前端面板，工作量较大，建议作为 v1.1 功能增强项。
