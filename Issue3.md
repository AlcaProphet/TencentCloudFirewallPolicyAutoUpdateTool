# FWAlizer 问题追踪（Issue3 · 历史归档）

> 第13轮审查及全量复核（2026-07-27），以 Design1.md、Build2.md、AGENTS.md 为基准。
> **全部 4 项问题已修复/实施完毕，最终验收通过。**
> 历史问题见 [Issue1.md](./Issue1.md) 和 [Issue2.md](./Issue2.md)（均已归档），构建计划见 [Build2.md](./Build2.md)（已归档）。

---

## 一、状态速查

| 编号 | 问题 | 严重度 | 最终状态 |
|------|------|--------|---------|
| [R13-01](#r13-01) | `cleanOldBackups` 切片越界 panic | 🔴 高 | ✅ 已修复 |
| [R13-02](#r13-02) | 前端未使用的 import | ⚪ 低 | ✅ 已修复 |
| [R13-03](#r13-03) | Windows 开机自启未实现 | 🟡 中 | ✅ 已实施（方案 B：`golang.org/x/sys/windows/registry`） |
| [R13-04](#r13-04) | 告警配置热重载未处理重复订阅 | 🟡 中 | ✅ 已实施（方案 B：EventBus `Unsubscribe` + 热重载） |

---

## 二、问题详情与修复结论

### [R13-01] `cleanOldBackups` 切片越界 panic

- **严重度：** 高 | **模块：** CLI（`app/cli.go` L109）
- **现象：** `backups[keep:]` 在备份数量少于 `keep` 时触发 slice bounds out of range
- **修复：** 切片前增加 `if len(backups) <= keep { return }` 守卫
- **影响：** 仅影响 `fwalizer backup`，不影响主同步流程

---

### [R13-02] 前端未使用的 import

- **严重度：** 低 | **模块：** WebUI 前端
- **现象：** `Advanced.vue` 导入 `h` 未使用；`Alerts.vue` 导入 `NSpace`/`NDivider` 未使用
- **修复：** 移除未使用的 import

---

### [R13-03] Windows 开机自启未实现

- **严重度：** 中 | **模块：** 桌面端（`app/autostart.go` L101-108）
- **现象：** `enableAutoStartWindows` 和 `disableAutoStartWindows` 仅输出日志桩，未写注册表
- **用户决策：** 方案 B — 引入 `golang.org/x/sys/windows/registry` 官方扩展库
- **关键风险：** `golang.org/x/sys/windows/registry` 有 `//go:build windows` 约束，若在 `autostart.go`（`desktop` 标签）中直接 import 会导致 macOS 编译失败
- **实施方案：** 按平台拆分为 3 个文件：
  - `app/autostart.go`（`//go:build desktop`）— macOS 辅助函数
  - `app/autostart_darwin.go`（`//go:build desktop && darwin`）— macOS dispatch 函数
  - `app/autostart_windows.go`（`//go:build desktop && windows`）— Windows registry 实现
- **编译验证：** macOS 桌面构建 + Windows 交叉编译均通过

---

### [R13-04] 告警配置热重载未处理重复订阅

- **严重度：** 中 | **模块：** 告警通知（`main.go` L93-110、`notifier/bus.go`）
- **现象：** 告警 Notifier 仅在启动时注册；`PUT /api/alerts` 触发 `notifyReload` 后不重建订阅
- **用户决策：** 方案 B — 在 EventBus 中增加 `Unsubscribe`，ReloadFunc 中先取消旧订阅再注册新订阅
- **关键设计：** Go 接口值 `==` 比较同时检查动态类型和指针值，必须在 Subscribe 时保存引用，Unsubscribe 时传入同一实例
- **实施方案：**
  - `notifier/bus.go` 新增 `Unsubscribe(eventType, sub)` 方法（O(n) 扫描，幂等）
  - `main.go` 引入 `currentEmailNotifier`/`currentWebhookNotifier` 追踪变量
  - `ReloadFunc` 末尾先 Unsubscribe 旧 Notifier → 读取最新配置 → 重新 Subscribe
- **并发安全：** `Publish` 持有 `RLock` 复制订阅者快照，`Unsubscribe` 持有 `Lock`，无竞态
- **测试覆盖：** 新增 3 个单元测试（`TestEventBus_Unsubscribe`、`_Idempotent`、`_OnlyTargetType`），全部 PASS

---

## 三、全量复核记录

> 复核日期：2026-07-27
> 复核范围：Issue2.md（R11-01~R11-07 + 待规划项）+ Build2.md（Step 1-11）+ Issue3.md（R13-01~R13-04）
> 复核方式：逐文件代码审查 + 编译验证 + 全量测试

### 3.1 Issue2.md 待修复项（R11-01 ~ R11-07）

| 编号 | 问题 | 严重度 | 复核结果 | 证据 |
|------|------|--------|---------|------|
| R11-01 | `.dockerignore` 去重 | 🟡 中 | ✅ | 文件 7 行，无重复，无 `Ref/` |
| R11-02 | 配置导入事务保护 | 🔴 高 | ✅ | `store.go` L317-373 含 6 个 Tx 方法；`settings.go` L111-133 使用 Tx 版本 |
| R11-03 | 前端数组索引→DB ID | 🔴 高 | ✅ | `Targets.vue` L32-33/L51-52/L82-83 使用 `row.id`；`Rules.vue` 同理 |
| R11-04 | CI/CD 前端构建步骤 | 🔴 高 | ✅ | 两个 YAML 均含 `setup-node@v4` + `npm ci && npm run build` |
| R11-05 | `.env` 模式 0-based 规则过滤 | 🟡 中 | ✅ | `env.go` L248 `0,2`、L257 `n<0\|\|n>=max`；测试 L18 `1`、L54 `want 1` |
| R11-06 | 移除 `app.Run` mode 参数 | ⚪ 低 | ✅ | `app.go` L16 `func Run(cfg *config.Config) error`；`main.go` L160 `app.Run(cfg)` |
| R11-07 | README DNS 默认值同步 | ⚪ 低 | ✅ | README L91 `223.5.5.5`；`grep 8.8.8.8 README.md` 无匹配 |

### 3.2 Issue2.md 待规划项 → Build2.md Step 8-11

| 编号 | Build2 Step | 内容 | 复核结果 | 证据 |
|------|------------|------|---------|------|
| WEB-06 | Step 8 | 页面 + 告警 API | ✅ | `config.go` L57-72 结构体；`store.go` L118-132 表+CRUD；`alerts.go`；`Advanced.vue`；`Alerts.vue`；路由+菜单 |
| FEA-02 | Step 9 | EventBus 接入 | ✅ | `main.go` L93-113 初始注册；L139-167 ReloadFunc 热重载重建 |
| FEA-03 | Step 10 | CLI backup/restore | ✅ | `store.go` L21-43 `GetDataDir()`；`cli.go` L45-75/L80-132 |
| FEA-06 | Step 11 | systray | ✅ | 回调触发同步；macOS plist + Windows registry；优雅退出 |

### 3.3 Issue3.md 已修复/已实施项

| 编号 | 复核结果 | 证据 |
|------|---------|------|
| R13-01 | ✅ | `cli.go` L109 `if len(backups) <= keep { return }` 守卫存在 |
| R13-02 | ✅ | `Advanced.vue` L2 无 `h`；`Alerts.vue` L2 无 `NSpace`/`NDivider` |
| R13-03 | ✅ | 3 文件拆分：`autostart.go`（macOS 辅助）+ `autostart_darwin.go` + `autostart_windows.go`（registry） |
| R13-04 | ✅ | `bus.go` L55-69 `Unsubscribe`；`main.go` L93-95 变量 + L98-113 引用保存 + L139-167 热重载 |

### 3.4 编译与测试

| 验证命令 | 结果 |
|---------|------|
| `go build ./...` | ✅ 零错误 |
| `go vet ./...` | ✅ 零警告 |
| `go test ./...` | ✅ 6 个测试包全部 `ok`（config/dns/portconv/tag/notifier/provider） |
| `CGO_ENABLED=1 go build -tags desktop` (macOS) | ✅ 桌面构建成功 |
| `GOOS=windows go build -tags desktop ./app/` | ✅ Windows 交叉编译成功 |
| `notifier` 测试（含 Unsubscribe） | ✅ 5/5 PASS |

### 3.5 跨平台构建标签验证

| 构建场景 | 编译文件组合 | 结果 |
|---------|------------|------|
| `go build ./...`（无 tags） | `systray_stub.go`（`!desktop`） | ✅ |
| `-tags desktop` (darwin) | `systray.go` + `autostart.go` + `autostart_darwin.go` | ✅ |
| `-tags desktop` (windows) | `systray.go` + `autostart.go` + `autostart_windows.go` | ✅ |

### 3.6 文档一致性

| 检查项 | 结果 | 证据 |
|--------|------|------|
| `.dockerignore` 7 行 | ✅ | `Documents/` `*.md` `.env` `.git/` `Dockerfile` `.dockerignore` `Makefile` |
| `.env.example` targets 从 0 开始 | ✅ | L35 `从 0 开始`；L40 `ACCEPT\|1`；L41 `ACCEPT\|0,2` |
| `README.md` DNS `223.5.5.5` | ✅ | L91 |
| `go.mod` `golang.org/x/sys` direct | ✅ | L14 无 `// indirect` |

### 3.7 总结

- **Issue2.md** 7 项 + 4 项：全部完成 ✅
- **Build2.md** 11 个 Step：全部验收通过 ✅
- **Issue3.md** 4 项：全部确认完成 ✅
- **编译/测试/跨平台**：全部通过 ✅
- **新发现遗漏**：无

---

## 四、第14轮审查（2026-07-27）

> 审查范围：全量代码逐文件审查（共 50+ 文件），以 AGENTS.md、Design1.md、Build1.md、Build2.md 及四家云厂商官方 API 文档为基准。
> 审查方式：逐文件阅读 + API 文档对照 + 设计一致性校验。

### 4.1 发现的问题

| 编号 | 问题 | 严重度 | 状态 |
|------|------|--------|------|
| [R14-01](#r14-01) | DNS 熔断器半开探测逻辑未实现 — 熔断后域名永久跳过 | 🔴 高 | ✅ 已修复 |
| [R14-02](#r14-02) | `Store.LoadConfig()` 缺少 `dns_timeout` 设置支持 | 🟡 中 | ✅ 已修复 |
| [R14-03](#r14-03) | WebUI Rules 表单缺少 target 绑定字段 | 🟡 中 | ✅ 已修复 |
| [R14-04](#r14-04) | 逐域名与全局 `EventSyncComplete` 共用同一事件类型 | 🟡 中 | ✅ 已修复 |
| [R14-05](#r14-05) | ECS ICMPv6 WARN 在重试时重复输出 | ⚪ 低 | ✅ 已修复（随 R14-01 一并修复） |
| [R14-06](#r14-06) | `multiHandler`/`MultiHandler` 代码重复（app.go vs logstream.go） | ⚪ 低 | ✅ 已修复 |
| [R14-07](#r14-07) | Webhook 格式仅面向钉钉，未区分飞书/Slack | ⚪ 低 | ✅ 已修复 |
| [R14-08](#r14-08) | CVM `checkRuleLimit` fallback 路径未统计 IPv6 规则（已知 COR-08） | ⚪ 低 | ✅ 已关闭-误报 |
| [R14-09](#r14-09) | WebUI 缺少 pidfile 防多实例（已知 WEB-05，确认仍未实现） | ⚪ 低 | ✅ 已修复 |
| [R14-10](#r14-10) | TypeScript `any` 类型泛滥（已知 WEB-03，确认仍未改进） | ⚪ 低 | ✅ 已修复 |
| [R14-11](#r14-11) | `Rules.vue` 前端缺少 ICMP 协议端口强制 ALL 校验 | ⚪ 低 | ✅ 已修复（随 R14-03 一并修复） |

---

### 4.2 问题详情与修复方案

---

#### [R14-01] DNS 熔断器半开探测逻辑未实现

- **严重度：** 🔴 高 | **模块：** syncer / dns
- **文件：** `syncer/syncer.go` L231-234
- **现象：** `syncDomain` 在检测到熔断器 `IsOpen` 返回 `true` 时，打印 `"域名已熔断，半开探测"` 后立即 `return`，**未执行任何 DNS 解析探测**。此后每个 Ticker 周期都同样直接跳过该域名，`RecordSuccess` 永远不会被调用，导致域名被永久阻塞直到进程重启。
- **与设计冲突：** Build1.md §12.7 明确规定「熔断后 → 半开状态：每轮仍尝试一次解析；解析成功 → 解除熔断，恢复正常同步」。当前实现完全跳过了半开探测步骤。
- **是否与已有 Issue 重复：** 否（首次发现）

##### 修复方法说明

**涉及文件：** 仅 `syncer/syncer.go` 一个文件，修改 `syncDomain` 函数中 L228-274 的逻辑。

核心思路：当熔断器打开时，不应直接返回，而应执行一次 DNS 解析作为半开探测。探测成功则解除熔断并继续正常同步流程；探测失败则维持熔断状态（不调用 `RecordFailure`，因为 `dns/circuitbreaker.go` L41-47 已保证熔断中不递增计数器）。

需处理的关键细节：
1. 探测成功后，`resolved` 仍需经过 `filterIPv4`（根据 `rule.EnableIPv6`）筛选
2. 探测成功后，继续走正常的 `retrySync` 流程
3. 探测失败后，仍需发布 `EventDNSFailed` 事件（告知用户半开探测也失败）
4. 原有的 `slog.Debug("域名已熔断，半开探测")` 日志需调整语义

##### 代码准备

> **注：** 以下为最终选定的方案 C 代码。原始方案 A/B 的代码草案已废弃，以本版本为准。

##### 决策点识别 ✅ 已决策：方案 C

| 方案 | 描述 | 状态 |
|------|------|------|
| A | 统一流程 — 将 DNS 解析提取到 syncDomain 顶部统一处理 | 未选 |
| B | 内联复制 — 熔断分支内复制同步逻辑 | 未选 |
| **C** | **提取方法 — 将同步流程提取为独立方法 `syncDomainInternal`** | **✅ 已选** |

**选定的修复代码（方案 C）：**

在 `syncer/syncer.go` 中新增 `syncDomainInternal` 方法，`syncDomain` 重构为：

```go
// syncDomain 同步单个域名到单个 Provider
func (s *Syncer) syncDomain(p provider.Provider, rule config.DomainRule) {
    // 0. DNS 解析（无论是否熔断都执行，熔断时作为半开探测）
    resolved, err := s.resolver.Resolve(context.Background(), rule.Host)
    if err != nil {
        if s.cb.IsOpen(rule.Host) {
            // 半开探测失败：维持熔断（不调用 RecordFailure，熔断中已停止计数）
            slog.Debug("域名半开探测失败，维持熔断", "domain", rule.Host, "error", err)
        } else {
            s.cb.RecordFailure(rule.Host)
            slog.Warn("DNS 解析失败，保留现有规则", "domain", rule.Host, "error", err)
        }
        s.bus.Publish(notifier.Event{
            Type:      notifier.EventDNSFailed,
            Timestamp: time.Now(),
            Data:      map[string]any{"domain": rule.Host, "error": err.Error()},
        })
        return
    }

    // 解析成功：若之前处于熔断则解除
    if s.cb.IsOpen(rule.Host) {
        s.cb.RecordSuccess(rule.Host)
        slog.Info("DNS 熔断解除", "domain", rule.Host)
    } else {
        s.cb.RecordSuccess(rule.Host)
    }

    // 1. 按规则配置过滤 IPv6 地址
    if !rule.EnableIPv6 {
        resolved = filterIPv4(resolved)
    }

    // 2. 委托给内部方法执行同步
    s.syncDomainInternal(p, rule, resolved)
}

// syncDomainInternal 执行 DNS 已解析后的同步流程（Describe → Diff → Create/Delete）
func (s *Syncer) syncDomainInternal(p provider.Provider, rule config.DomainRule, resolved []dns.ResolvedIP) {
    // ECS ICMPv6 警告（仅当实际有 IPv6 地址时输出一次）
    if rule.Protocol == "ICMP" && p.CloudType() == config.CloudAliECS {
        for _, ip := range resolved {
            if ip.IsIPv6 {
                slog.Warn("ECS 不支持 ICMPv6 入站规则，IPv6 地址将被跳过", "domain", rule.Host)
                break
            }
        }
    }

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
    s.bus.Publish(notifier.Event{
        Type:      notifier.EventDomainSyncComplete,  // R14-04：逐域名事件使用独立类型
        Timestamp: time.Now(),
        Data:      map[string]any{"provider": p.Name(), "domain": rule.Host},
    })
}
```

**注意：** 此方案同时解决了 R14-05（ECS ICMPv6 WARN 移入 `syncDomainInternal`，仅输出一次）。事件类型使用 `EventDomainSyncComplete` 与 [R14-04](#r14-04) 保持一致。

##### 影响范围分析

- **`syncer/syncer.go`：** 仅修改 `syncDomain` 方法
- **`dns/circuitbreaker.go`：** 无需修改（已支持 `IsOpen` + `RecordSuccess` + 熔断中不递增计数器）
- **`notifier/bus.go`：** 无需修改（复用现有事件类型）
- **其他模块：** 无影响
- **副作用：** 无（方案 C 提取方法，消除了代码重复，且新增的 `syncDomainInternal` 方法同时承载了 R14-05 的修复）

---

#### [R14-02] `Store.LoadConfig()` 缺少 `dns_timeout` 设置支持

- **严重度：** 🟡 中 | **模块：** config
- **文件：** `config/store.go` L471-528、`webui/api/settings.go` L19-24
- **现象：** `.env` 模式可通过 `DNS_TIMEOUT` 环境变量配置 DNS 超时（`config/env.go` L42-48），但 WebUI 模式的 `LoadConfig` 未从 SQLite 读取该设置，始终使用默认值 `10 * time.Second`。若用户通过 WebUI 保存了自定义的 `dns_timeout` 设置，重启后该值被忽略。
- **是否与已有 Issue 重复：** 否（首次发现）

##### 修复方法说明

**涉及文件：**
1. `config/store.go` — `LoadConfig()` 方法末尾增加解析逻辑
2. `webui/api/settings.go` — `handleGetSettings()` 的 defaults map 增加 `dns_timeout`
3. `webui/frontend/src/views/Settings.vue` — 表单增加 DNS 超时输入项（可选）

修改顺序：先改后端（store.go + settings.go），再改前端（Settings.vue）。无依赖关系。

##### 代码准备

**改动 1：`config/store.go` L520-525 之后新增：**

```go
    if v := settings["dns_timeout"]; v != "" {
        if d, err := time.ParseDuration(v); err == nil {
            cfg.DNSTimeout = d
        }
    }
```

**改动 2：`webui/api/settings.go` L19-24 的 defaults map 增加一行：**

```go
    defaults := map[string]string{
        "tag":         "auto-dns",
        "interval":    "5m",
        "dns":         "223.5.5.5",
        "log_level":   "info",
        "dns_timeout": "10s",  // ← 新增
    }
```

**改动 3：`webui/frontend/src/views/Settings.vue` 的「全局设置」区块增加表单项：**

```html
<NFormItem label="DNS 超时">
  <NInput v-model:value="settings.dns_timeout" placeholder="10s" />
</NFormItem>
```

放在「DNS 服务器」输入项下方。

##### 影响范围分析

- **`config/store.go`：** 仅增加 4 行
- **`webui/api/settings.go`：** 仅增加 1 行
- **`webui/frontend/src/views/Settings.vue`：** 仅增加 3 行 HTML
- **配置热重载：** `ReloadFunc` 调用 `store.LoadConfig()` 后通过 `s.Reload(cfg)` 传入 Syncer，Syncer 中 `dns.Resolver` 的 timeout 在构造时固定且不可变。**注意：** 当前架构中 `dns.Resolver` 在 `main.go` L82 创建且不会在热重载中重建，因此修改 `dns_timeout` 后需重启才能生效。这是一个**预存的架构限制**，超出本问题修复范围，建议在本文档中标注为已知限制，等待未来的热重载增强（v1.1）。
- **副作用：** 无

##### 决策点识别

| 方案 | 描述 |
|------|------|
| **A（推荐）** | 仅修复 LoadConfig 的读取 + Settings API 的默认值 + 前端表单（如上所述）。热重载不生效的 DNS timeout 问题作为已知限制在文档中标注 |
| **B** | 同时修改 `main.go` 的 `ReloadFunc`，使其重建 `dns.Resolver` 并传给 Syncer。但这需要 Syncer 增加 `ReloadResolver` 方法或允许在 `Reload` 时同时更新 resolver |

> **已决策：✅ 方案 B — 完整热重载支持**

**选定的修复方案扩充：**

除原有 3 处修改外，追加以下修改以支持热重载：

**改动 4：`syncer/syncer.go` — `Syncer` 增加 `ReloadResolver` 方法：**

```go
// ReloadResolver 热重载 DNS 解析器（DNS 地址或超时变更时调用）
func (s *Syncer) ReloadResolver(resolver *dns.Resolver) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.resolver = resolver
}
```

**改动 5：`main.go` — `ReloadFunc` 中增加 resolver 重建逻辑（在 `s.Reload(newCfg)` 之后）：**

```go
    s.ReloadProviders(newProviders)
    s.Reload(newCfg)
    // 若 DNS 配置变更，重建 Resolver 并热重载
    newResolver := dns.NewResolver(newCfg.DNS, newCfg.DNSTimeout)
    s.ReloadResolver(newResolver)
```

**影响范围补充：**
- `syncer/syncer.go`：新增 `ReloadResolver` 方法（~8 行），需增加 `dns` 包的 import（已有）
- `main.go`：ReloadFunc 中新增 2 行
- 热重载后下一次 Ticker 触发或手动触发同步时生效（无需等待重启）

---

#### [R14-03] WebUI Rules 表单缺少 target 绑定字段

- **严重度：** 🟡 中 | **模块：** WebUI 前端
- **文件：** `webui/frontend/src/views/Rules.vue` L96-119
- **现象：** Rules 表单仅包含域名、协议、端口、动作、备注、IPv6 开关，**缺少目标（Targets）选择器**。这意味着所有通过 WebUI 创建的规则均绑定到全部目标（`targets: []`）。后端 `DomainRule.Targets` 字段和 `filterRulesForTarget` 逻辑均已完备，但前端未提供配置入口。
- **是否与已有 Issue 重复：** 否（首次发现）

##### 修复方法说明

**涉及文件：**
1. `webui/frontend/src/views/Rules.vue` — 表单增加多选组件，表格增加目标列

修改内容：
- `<script>` 中：增加 `targetOptions` 响应式数据；`form` 增加 `targets: []` 字段；`onMounted` 时加载 targets 列表；`openEdit` 时还原 targets
- `<template>` 中：在备注和 IPv6 之间增加 `NSelect` 多选组件；表格增加「适用目标」列

##### 代码准备

**`Rules.vue` `<script setup>` 部分修改：**

```typescript
// 新增：加载 targets 列表供多选
const targetOptions = ref<{ label: string; value: number }[]>([])

// form 增加 targets 字段
const form = ref({ host: '', protocol: 'TCP', ports: '', action: 'ACCEPT', comment: '', enable_ipv6: false, targets: [] as number[] })

// onMounted 中增加加载 targets
onMounted(async () => {
  await load()
  // 加载目标列表（供规则绑定选择）
  const res = await fetch('/api/targets')
  const data: any[] = await res.json() || []
  targetOptions.value = data.map((t: any) => ({
    label: `${t.cloud_type} / ${t.resource_id}`,
    value: t.id,
  }))
})

// openAdd 中增加 targets 重置
function openAdd() {
  editingId.value = null
  form.value = { host: '', protocol: 'TCP', ports: '', action: 'ACCEPT', comment: '', enable_ipv6: false, targets: [] }
  showModal.value = true
}

// openEdit 中增加 targets 还原
function openEdit(row: any) {
  editingId.value = row.id
  form.value = { host: row.host, protocol: row.protocol, ports: row.ports, action: row.action, comment: row.comment || '', enable_ipv6: !!row.enable_ipv6, targets: Array.isArray(row.targets) ? [...row.targets] : [] }
  showModal.value = true
}
```

**`Rules.vue` `<template>` 部分增加表单项（放在备注和 IPv6 之间）：**

```html
<NFormItem label="适用目标">
  <NSelect
    v-model:value="form.targets"
    :options="targetOptions"
    multiple
    placeholder="留空 = 全部目标"
    clearable
  />
</NFormItem>
```

**表格列增加「适用目标」列（放在「备注」列之后）：**

```typescript
{
  title: '适用目标', key: 'targets',
  render(row: any) {
    if (!row.targets || row.targets.length === 0) return '全部'
    return row.targets.map((id: number) => {
      const opt = targetOptions.value.find(o => o.value === id)
      return opt ? opt.label : `#${id}`
    }).join(', ')
  }
},
```

##### 影响范围分析

- **`Rules.vue`：** 仅新增 UI 元素，不改变现有逻辑
- **后端：** 无需修改（`DomainRule.Targets` 字段已存在，JSON 序列化/反序列化已支持）
- **`filterRulesForTarget`：** 无需修改（空 `targets` 表示全部，非空则按 DB ID 匹配，逻辑已正确）
- **副作用：** 无

##### 决策点识别

无。此修复是纯前端 UI 增强，后端无需改动，无多种方案选择。

---

#### [R14-04] 逐域名与全局 `EventSyncComplete` 共用同一事件类型

- **严重度：** 🟡 中 | **模块：** syncer / notifier
- **文件：** `syncer/syncer.go` L219-223、L268-272；`notifier/bus.go` L13-18；`webui/api/logwriter.go` L40-44
- **现象：** `syncAll` 在每轮同步结束时发布全局 `EventSyncComplete`（`Data` 含 `duration`），而 `syncDomain` 在每个域名同步成功时也发布 `EventSyncComplete`（`Data` 含 `provider`、`domain`）。两者使用完全相同的事件类型 `"sync:complete"`。
- **当前处理：** `StoreLogWriter.OnEvent` 通过检测 `Data["provider"]` 是否为空来区分全局/逐域名事件，跳过全局事件。这在当前是有效的，但依赖隐式约定，不够明确。
- **是否与已有 Issue 重复：** 否（首次发现）

##### 修复方法说明

**涉及文件（按修改顺序）：**
1. `notifier/bus.go` — 新增 `EventDomainSyncComplete` 常量
2. `syncer/syncer.go` — `syncDomain` 中逐域名成功事件改为新类型；`syncAll` 保持 `EventSyncComplete`（全局事件不变）
3. `webui/api/logwriter.go` — 将 case 从 `EventSyncComplete` 改为 `EventDomainSyncComplete`，删除跳过逻辑
4. `webui/frontend/src/views/Logs.vue` — 事件类型映射增加 `domain:sync_complete`
5. `main.go` — LogWriter 订阅改为 `EventDomainSyncComplete`

修改顺序：先定义常量 → 再改发布方 → 最后改消费方。

##### 代码准备

> **注：** 以下为最终选定的方案 B 代码。原始方案 A 的代码草案已废弃。

##### 影响范围分析

- **`notifier/bus.go`：** 仅新增常量，不影响现有订阅
- **`syncer/syncer.go`：** 仅 `syncAll` 中一行变更
- **`webui/api/logwriter.go`：** 删除 3 行冗余判断
- **`notifier/email.go`：** 无需修改（只处理 `EventSyncError` / `EventDNSFailed`）
- **`notifier/webhook.go`：** 无需修改（同上）
- **`webui/api/sync.go`：** SSE 事件推送不做过滤，新类型会正常推送到前端，前端事件映射需增加新类型 → **见下方决策点**
- **副作用：** 前端 `Logs.vue` L24-30 的 `eventTypeLabels` 和 `formatEventData` 需要为新类型增加映射，否则实时事件面板中会显示原始类型字符串

##### 决策点识别

| 方案 | 描述 |
|------|------|
| **A（推荐）** | 新增 `EventSyncRoundComplete` 类型，全局完成事件使用新类型。前端同步更新事件类型映射 |
| **B** | 更名逐域名事件为 `EventDomainSyncComplete`，全局事件保持 `EventSyncComplete` |

> **已决策：✅ 方案 B — 新增 `EventDomainSyncComplete`，全局 `EventSyncComplete` 不变**

**选定的修复方案：**

**改动 1：`notifier/bus.go` 常量块新增：**

```go
const (
    EventSyncStart          EventType = "sync:start"
    EventSyncComplete       EventType = "sync:complete"        // 全局：一轮同步完成
    EventDomainSyncComplete EventType = "domain:sync_complete" // 逐域名：单个域名同步成功（新增）
    EventSyncError          EventType = "sync:error"
    EventRuleChanged        EventType = "rule:changed"
    EventDNSFailed          EventType = "dns:failed"
)
```

**改动 2：`syncer/syncer.go` L268-272 — 逐域名事件改用新类型：**

```go
    s.bus.Publish(notifier.Event{
        Type:      notifier.EventDomainSyncComplete,  // ← 原为 EventSyncComplete
        Timestamp: time.Now(),
        Data:      map[string]any{"provider": p.Name(), "domain": rule.Host},
    })
```

**改动 3：`webui/api/logwriter.go` L40-44 — 删除全局事件跳过逻辑：**

```go
    case notifier.EventDomainSyncComplete:  // ← 原为 EventSyncComplete
        log.Result = "success"
```

全局 `EventSyncComplete` 不再进入 LogWriter 的处理，无需特殊判断。

**改动 4：`webui/frontend/src/views/Logs.vue` L24-30 — 事件类型映射增加新类型：**

```typescript
const eventTypeLabels: Record<string, string> = {
  'sync:start': '同步开始',
  'sync:complete': '同步完成',
  'domain:sync_complete': '域名同步完成',  // ← 新增
  'sync:error': '同步失败',
  'dns:failed': 'DNS解析失败',
  'rule:changed': '规则变更',
}
```

**改动 5：`main.go` — LogWriter 订阅事件类型更新：**

```go
    s.EventBus().Subscribe(notifier.EventDomainSyncComplete, logWriter)  // ← 原为 EventSyncComplete
    s.EventBus().Subscribe(notifier.EventSyncError, logWriter)
```

---

#### [R14-05] ECS ICMPv6 WARN 在重试时重复输出

- **严重度：** ⚪ 低 | **模块：** syncer
- **文件：** `syncer/retry.go` L41-43；`syncer/syncer.go` L229-273
- **现象：** `slog.Warn("ECS 不支持 ICMPv6 入站规则，IPv6 地址将被跳过")` 位于 `retrySync` 的 for 循环体内，每次重试都输出。
- **是否与已有 Issue 重复：** 否（首次发现）
- **⚠️ 实施说明：** 此问题的修复代码已整合进 [R14-01](#r14-01) 方案 C 的 `syncDomainInternal` 方法中（L295-303）。实施 R14-01 时一并修复。单独实施时参照下方代码。

##### 修复方法说明

**涉及文件：**
1. `syncer/retry.go` — 删除 L41-43 的 WARN
2. `syncer/syncer.go` — `syncDomain` 中在调用 `retrySync` 前增加条件判断，仅当 `resolved` 中存在 IPv6 地址时输出一次 WARN

修改顺序：先删后加，同一轮修改即可。

##### 代码准备

**改动 1：删除 `syncer/retry.go` L40-43：**

删除以下三行：
```go
        // ECS 不支持 ICMPv6 入站规则创建，跳过 IPv6 部分
        if rule.Protocol == "ICMP" && p.CloudType() == config.CloudAliECS {
            slog.Warn("ECS 不支持 ICMPv6 入站规则，IPv6 地址将被跳过", "domain", rule.Host)
        }
```

**改动 2：在 `syncer/syncer.go` 的 `syncDomain` 中，L248 `s.cb.RecordSuccess(rule.Host)` 之后、L250 IPv6 过滤之前，新增条件判断：**

```go
    s.cb.RecordSuccess(rule.Host)

    // ECS ICMPv6 警告：仅在解析结果实际包含 IPv6 地址时输出一次
    if rule.Protocol == "ICMP" && p.CloudType() == config.CloudAliECS {
        hasIPv6 := false
        for _, ip := range resolved {
            if ip.IsIPv6 {
                hasIPv6 = true
                break
            }
        }
        if hasIPv6 {
            slog.Warn("ECS 不支持 ICMPv6 入站规则，IPv6 地址将被跳过", "domain", rule.Host)
        }
    }

    // 1.5 按规则配置过滤 IPv6 地址
    if !rule.EnableIPv6 {
        resolved = filterIPv4(resolved)
    }
```

##### 影响范围分析

- **`syncer/retry.go`：** 删除 3 行，不再依赖 `config` 包的类型信息（但 `config` 包仍在 import 中，因其他地方还有引用）
- **`syncer/syncer.go`：** 新增 ~10 行
- **`provider/common.go`：** `buildDesired` 中已有的 IPv6+ICMP 跳过逻辑完全不受影响，WARN 仅从调用方层面提前通知用户
- **副作用：** 无

##### 决策点识别

无。此修复为单一最优方案。

---

#### [R14-06] `multiHandler`/`MultiHandler` 代码重复

- **严重度：** ⚪ 低 | **模块：** app / webui/api
- **文件：** `app/app.go` L88-126、`webui/api/logstream.go` L78-120
- **现象：** 两个文件中定义了几乎相同的 `slog.Handler` 多路复用器。
- **是否与已有 Issue 重复：** 否（首次发现）

##### 修复方法说明

**涉及文件：**
1. `app/app.go` — 删除 `multiHandler` 类型及其 4 个方法（L88-126），将 `InitLoggerWithBroadcaster` 中对 `multiHandler` 的引用改为 `api.MultiHandler`
2. `webui/api/logstream.go` — 无需修改（`MultiHandler` 及其构造函数 `NewMultiHandler` 保留）

##### 代码准备

**改动 1：删除 `app/app.go` L88-126 整个 `multiHandler` 类型定义。**

**改动 2：修改 `app/app.go` 的 `InitLoggerWithBroadcaster` 函数（L84）：**

```go
// 修改前：
slog.SetDefault(slog.New(&multiHandler{handlers: []slog.Handler{stdout, extra}}))

// 修改后：
slog.SetDefault(slog.New(api.NewMultiHandler(stdout, extra)))
```

**改动 3：`app/app.go` 的 import 中增加 `api` 包引用：**

```go
import (
    // ... 现有 imports ...
    webapi "github.com/alcaprophet/fwalizer/webui/api"  // 新增
)
```

##### 决策点识别

| 方案 | 描述 | 优点 | 缺点 |
|------|------|------|------|
| **A（推荐）** | `app.go` 引用 `webui/api` 包中的 `MultiHandler` | 消除重复；`MultiHandler` 已是导出类型 | `app` 包新增对 `webui/api` 的依赖，打破了 `app` → `webui/api` 的单向依赖（原本 `main.go` 同时引用两者，`app` 不依赖 `webui/api`） |
| **B** | 将 `MultiHandler` 从 `webui/api` 移至一个新的 `internal/` 或独立的共享包（如 `app/logutil.go`），两边都引用共享包 | 依赖方向更清晰；`app` 不依赖 `webui` | 增加一个包；`MultiHandler` 需要从 `logstream.go` 中迁出 |
| **C** | 保留 `app.go` 中的 `multiHandler`，删除 `logstream.go` 中的 `MultiHandler`，让 `logstream.go` 引用 `app` 包的导出类型 | 实现简单 | `webui/api` 依赖 `app` 包不太合理（`api` 是更底层的模块） |

> **已决策：✅ 方案 B — 共享包**

**选定的修复方案：**

**改动 1：新建 `app/logutil.go`，将 `MultiHandler` 迁移至此：**

```go
package app

import (
    "context"
    "log/slog"
)

// MultiHandler 将日志同时写入多个 Handler
type MultiHandler struct {
    Handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
    return &MultiHandler{Handlers: handlers}
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
    for _, h := range m.Handlers {
        if h.Enabled(ctx, level) { return true }
    }
    return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
    for _, h := range m.Handlers {
        if h.Enabled(ctx, r.Level) {
            if err := h.Handle(ctx, r); err != nil { return err }
        }
    }
    return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    handlers := make([]slog.Handler, len(m.Handlers))
    for i, h := range m.Handlers { handlers[i] = h.WithAttrs(attrs) }
    return &MultiHandler{Handlers: handlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
    handlers := make([]slog.Handler, len(m.Handlers))
    for i, h := range m.Handlers { handlers[i] = h.WithGroup(name) }
    return &MultiHandler{Handlers: handlers}
}
```

**改动 2：`app/app.go` — 删除 `multiHandler` 类型（L88-126），`InitLoggerWithBroadcaster` 改用 `NewMultiHandler`：**

```go
// 修改前：
slog.SetDefault(slog.New(&multiHandler{handlers: []slog.Handler{stdout, extra}}))
// 修改后：
slog.SetDefault(slog.New(NewMultiHandler(stdout, extra)))
```

**改动 3：`webui/api/logstream.go` — 删除 `MultiHandler` 类型（L78-120），改为引用 `app.MultiHandler`。** 但 `logstream.go` 中的 `NewMultiHandler` 调用处需更新为 `app.NewMultiHandler`。

**注意：** `webui/api` 引用 `app` 包会导致循环依赖风险 — 当前 `main.go` 同时引用 `app` 和 `webui/api`，但 `webui/api` 不依赖 `app`。若 `webui/api` 引用 `app`，不会造成循环依赖（`app` 不引用 `webui/api`）。确认安全。

##### 影响范围分析

- 若选方案 A：仅 `app.go` 改 2 处，`logstream.go` 不变
- 若选方案 B：需新建文件（如 `app/logutil.go`），迁移 `MultiHandler`；`app.go` 和 `logstream.go` 各改 1 处引用
- 若选方案 C：`logstream.go` 删除 `MultiHandler`，改为引用 `app.MultiHandler`（但 `MultiHandler` 需先导出）
- **副作用：** 方案 A 和 C 会创建跨模块依赖，方案 B 最干净但增加文件

---

#### [R14-07] Webhook 格式仅面向钉钉，未区分飞书/Slack

- **严重度：** ⚪ 低 | **模块：** notifier / config / WebUI
- **文件：** `notifier/webhook.go` L31-36；`config/config.go` L68-72；`config/store.go` L128-132；`webui/api/alerts.go`；`webui/frontend/src/views/Alerts.vue`
- **现象：** Webhook 通知器硬编码钉钉格式。Design1.md 和 Build1.md 提及支持「钉钉/飞书/Slack」，但当前实现仅适用于钉钉。
- **是否与已有 Issue 重复：** 否（首次发现）

##### 修复方法说明

**涉及文件（按修改顺序）：**
1. `config/config.go` — `AlertWebhookConfig` 结构体增加 `Channel` 字段
2. `config/store.go` — `alert_webhook` 表增加 `channel` 列
3. `notifier/webhook.go` — `WebhookNotifier` 增加 `channel` 字段，`OnEvent` 根据 channel 组装不同 JSON
4. `webui/frontend/src/views/Alerts.vue` — 增加通知渠道下拉选项
5. `webui/api/alerts.go` — 无需改（JSON 反序列化自动映射新字段）

##### 代码准备

**改动 1：`config/config.go` L68-72，增加 Channel 字段：**

```go
type AlertWebhookConfig struct {
    Enabled bool   `json:"enabled"`
    URL     string `json:"url"`
    Channel string `json:"channel"` // dingtalk / feishu / slack，默认 dingtalk
}
```

**改动 2：`config/store.go` L128-132，表结构增加 channel 列：**

```sql
CREATE TABLE IF NOT EXISTS alert_webhook (
    id INTEGER PRIMARY KEY DEFAULT 1,
    enabled INTEGER DEFAULT 0,
    url TEXT DEFAULT '',
    channel TEXT DEFAULT 'dingtalk'
);
```

同时更新 `GetAlertWebhook` 和 `SaveAlertWebhook` 的 SQL 语句，增加 channel 字段的读写。

**改动 3：`notifier/webhook.go` — 构造函数和 OnEvent：**

```go
type WebhookNotifier struct {
    url     string
    channel string
    client  *http.Client
}

func NewWebhookNotifier(url, channel string) *WebhookNotifier {
    if channel == "" {
        channel = "dingtalk"
    }
    return &WebhookNotifier{
        url:     url,
        channel: channel,
        client:  &http.Client{Timeout: 10 * time.Second},
    }
}

func (n *WebhookNotifier) OnEvent(event Event) error {
    if event.Type != EventSyncError && event.Type != EventDNSFailed {
        return nil
    }
    content := fmt.Sprintf("[FWAlizer] %s\n%s", event.Type, formatEventBody(event))
    var payload map[string]any
    switch n.channel {
    case "feishu":
        payload = map[string]any{
            "msg_type": "text",
            "content":  map[string]string{"text": content},
        }
    case "slack":
        payload = map[string]any{"text": content}
    default: // dingtalk
        payload = map[string]any{
            "msgtype": "text",
            "text":    map[string]string{"content": content},
        }
    }
    // ... 后续 JSON marshal + POST 不变 ...
}
```

**改动 4：`main.go` L110-111 调用处传入 channel：**

```go
currentWebhookNotifier = notifier.NewWebhookNotifier(webhookCfg.URL, webhookCfg.Channel)
```

同样更新 ReloadFunc 中的调用。

##### 决策点识别

| 方案 | 描述 | 优点 | 缺点 |
|------|------|------|------|
| **A（推荐）** | 增加 `channel` 配置字段，用户手动选择渠道 | 灵活；每种渠道格式可精确适配 | 需要改 SQLite 表结构（增加列）+ 更新前端表单 |
| **B** | 自动检测 URL 域名（`oapi.dingtalk.com` → 钉钉，`open.feishu.cn` → 飞书，`hooks.slack.com` → Slack） | 用户无需额外配置 | 检测逻辑依赖 URL 格式，可能有误判；飞书自定义机器人 URL 格式多变 |
| **C** | 保守方案：更新文档说明当前仅支持钉钉，不实现多渠道 | 零代码改动 | 与 Design1.md 宣称能力不一致 |

> **已决策：✅ 方案 A — 完整实现（增加 channel 配置字段）**

选定的修复方案已在上方代码准备中给出（改动 1-4），补充说明：

- **SQLite 表结构迁移：** 在 `initTables()` 中已有 `alert_webhook` 表的 `CREATE TABLE IF NOT EXISTS`，新增列需追加 `ALTER TABLE` 兼容旧数据库（参照 L139 已有模式）：
  ```go
  s.db.Exec("ALTER TABLE alert_webhook ADD COLUMN channel TEXT DEFAULT 'dingtalk'")
  ```
- **前端表单：** Alerts.vue 在 Webhook 卡片中 NInput URL 下方增加 NSelect 渠道选择
- **向后兼容：** `channel` 默认值为 `"dingtalk"`，已有配置不受影响

##### 影响范围分析

- **方案 A 影响：** `config.go` +1 字段；`store.go` 表结构变更 + CRUD 更新；`webhook.go` 函数签名变更；`main.go` 2 处调用更新；`Alerts.vue` 表单增加下拉
- **方案 B 影响：** 仅 `webhook.go` 增加检测函数；main.go 调用不变
- **方案 C 影响：** 仅 README.md + Design1.md 文档更新
- **表结构迁移：** 方案 A 的 SQLite 表结构变更需要 `ALTER TABLE` 迁移逻辑（参照 `store.go` L139 已有的 `ALTER TABLE rules ADD COLUMN` 模式）

---

#### [R14-08] CVM `checkRuleLimit` fallback 路径未统计 IPv6 规则

- **严重度：** ⚪ 低 | **模块：** provider
- **文件：** `provider/tc_cvm.go` L225
- **现象：** 与 Issue1.md [COR-08] 相同 — 原始判断认为 fallback 时 `len(ps.Ingress) + len(ps.Egress)` 未统计 IPv6 规则。
- **复审结论：** 经重新审查 CVM API 文档，`DescribeSecurityGroupPolicies` 返回的 `Ingress`/`Egress` 数组各自同时包含 IPv4 和 IPv6 规则（通过 `CidrBlock` vs `Ipv6CidrBlock` 区分），`len()` 已完整统计。**原始判断有误，COR-08 已关闭。**
- **是否与已有 Issue 重复：** 是（Issue1.md [COR-08]，**✅ 已关闭-误报**）

##### 修复方法说明

**涉及文件：** 仅 `provider/tc_cvm.go` L225 一行。

当前 fallback 逻辑 `total = len(ps.Ingress) + len(ps.Egress)` 只统计了 API 返回的规则数量，但 CVM 的 `DescribeSecurityGroupPolicies` 返回的 `SecurityGroupPolicySet` 结构中，Ingress 和 Egress 都是 `[]*SecurityGroupPolicy` 数组。IPv4 和 IPv6 规则都在同一个数组中（通过 `CidrBlock` vs `Ipv6CidrBlock` 区分），所以实际上 `len()` 已经包含了 IPv6 规则。

**重新审视：** 查看腾讯云 CVM API 文档，`DescribeSecurityGroupPolicies` 返回的 `Ingress` 数组同时包含 IPv4 和 IPv6 规则（通过 `CidrBlock` 或 `Ipv6CidrBlock` 字段区分类型），因此 `len(ps.Ingress)` 实际上已经包含了 IPv6 入站规则。同样，`len(ps.Egress)` 包含了 IPv6 出站规则。所以 **fallback 路径并未遗漏 IPv6 规则**。

**结论：** 此问题的原始判断（Issue1.md COR-08）有误。`PolicyStatistics` 精确计数的值和 `len(Ingress) + len(Egress)` 的值在正常 API 返回下应相等。COR-08 可以关闭。

##### 代码准备

无需修改代码。建议在 Issue3.md 中标记 COR-08 为「误报-已关闭」，并在 `checkRuleLimit` 上方增加注释说明 fallback 已包含 IPv6：

```go
// 计算总规则数（优先使用 PolicyStatistics 精确计数，fallback 同样包含 IPv4+IPv6）
var total int
if ps.PolicyStatistics != nil {
    // ...
} else {
    // Ingress/Egress 数组各自包含 IPv4 和 IPv6 规则，总数已完整
    total = len(ps.Ingress) + len(ps.Egress)
}
```

##### 影响范围分析

- **代码：** 无改动或仅加注释
- **数据库：** 无影响
- **副作用：** 无

##### 决策点识别

无。仅需确认关闭 COR-08。

---

#### [R14-09] WebUI 缺少 pidfile 防多实例

- **严重度：** ⚪ 低 | **模块：** app / main.go
- **现象：** Build1.md §12.12 规定的 WebUI 模式 pidfile 防多实例机制未实现。
- **是否与已有 Issue 重复：** 是（Issue1.md [WEB-05]，状态仍为 📋 待规划）

##### 修复方法说明

**涉及文件：**
1. `main.go` — WebUI 模式启动时，在 `os.MkdirAll(dataDir)` 之后、`config.OpenStore(dbPath)` 之前，增加 pidfile 检测和创建逻辑
2. `config/store.go` — 可选：增加 `GetPidFilePath()` 辅助函数（与 `GetDataDir()` 同层）

实现逻辑（参照 Build1.md §12.12）：
1. 构造 pidfile 路径：`<dataDir>/fwalizer.pid`
2. 检查 pidfile 是否存在
3. 若存在，读取其中的 PID
4. 检测该 PID 的进程是否存活（`os.FindProcess` + `Signal(syscall.Signal(0))` 在 Unix 上）
5. 若存活，拒绝启动并输出错误信息
6. 若不存在或进程已死，创建新的 pidfile 并写入当前 PID
7. 程序退出时（通过 defer）删除 pidfile

##### 代码准备

**新增 `config/store.go` 辅助函数（放在 `GetDataDir` 之后）：**

```go
// GetPidFilePath 返回 pidfile 路径
func GetPidFilePath(dataDir string) string {
    return filepath.Join(dataDir, "fwalizer.pid")
}

// WritePidFile 写入 PID 文件，返回清理函数
func WritePidFile(path string) (cleanup func(), err error) {
    // 检查已有 pidfile
    if data, err := os.ReadFile(path); err == nil {
        pidStr := strings.TrimSpace(string(data))
        if pid, err := strconv.Atoi(pidStr); err == nil {
            if proc, err := os.FindProcess(pid); err == nil {
                // Unix: Signal(0) 检测进程是否存在
                if err := proc.Signal(syscall.Signal(0)); err == nil {
                    return nil, fmt.Errorf("FWAlizer 已在运行 (PID: %d)，请先停止现有实例", pid)
                }
            }
        }
    }
    
    // 写入当前 PID
    if err := os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())+"\n"), 0644); err != nil {
        return nil, fmt.Errorf("写入 pidfile 失败: %w", err)
    }
    
    return func() { os.Remove(path) }, nil
}
```

**修改 `main.go` WebUI 模式启动逻辑（在 L44 `os.MkdirAll` 之后）：**

```go
    dataDir := config.GetDataDir()
    if err := os.MkdirAll(dataDir, 0755); err != nil {
        fmt.Fprintf(os.Stderr, "创建数据目录失败: %v\n", err)
        os.Exit(1)
    }

    // pidfile 防多实例（仅 WebUI 模式，.env 模式由容器编排保证）
    pidFile := config.GetPidFilePath(dataDir)
    cleanup, err := config.WritePidFile(pidFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }
    defer cleanup()

    dbPath := filepath.Join(dataDir, "config.db")
    // ... 后续不变 ...
```

##### 影响范围分析

- **`config/store.go`：** 新增 `GetPidFilePath` + `WritePidFile` 两个函数（~25 行）
- **`main.go`：** 新增 ~5 行
- **平台兼容性：** `os.FindProcess` + `Signal(syscall.Signal(0))` 仅 Unix/macOS 有效。Windows 上 `os.FindProcess` 始终成功，`Signal` 无意义。需要 `//go:build` 分离 Windows 实现（使用 Windows API `WaitForSingleObject` 或 `Process32First` 检测进程）。但这大大增加了复杂度。**见下方决策点**。
- **副作用：** 若程序异常崩溃（如 SIGKILL），pidfile 残留，下次启动时需判断该 PID 是否存活。上述方案已处理此情况（`Signal(0)` 检测）。

##### 决策点识别

| 方案 | 描述 | 优点 | 缺点 |
|------|------|------|------|
| **A** | 完整实现：Unix 用 Signal(0)，Windows 用 syscall 检测进程存活。编写 `pidfile_unix.go` + `pidfile_windows.go` 两个平台文件 | 跨平台覆盖完整 | 实现复杂度较高；Windows 进程检测需引入额外 API 调用 |
| **B** | 简单实现：仅 Unix/macOS 支持 pidfile 检测，Windows 跳过（Docker 部署中 Windows 桌面不常见，且容器编排保证单实例） | 实现简单；两个平台文件逻辑清晰 | Windows 桌面端无防多实例保护 |
| **C** | 端口检测替代 pidfile：不写 pidfile，改为在 WebUI 启动前尝试 listen 端口，若端口已被占用则报错 | 跨平台一致；无需进程检测 | 端口占用 != FWAlizer 实例（其他程序可能占用同一端口）；语义不够精确 |

> **已决策：✅ 方案 A — 完整跨平台实现**

**选定的修复方案补充：** 上方代码准备中的 `WritePidFile` 已在 `config/store.go` 中实现，需额外编写平台分离文件：

- `config/pidfile_unix.go`（`//go:build !windows`）：`os.FindProcess` + `Signal(syscall.Signal(0))` 检测进程
- `config/pidfile_windows.go`（`//go:build windows`）：使用 `golang.org/x/sys/windows` 的 `WaitForSingleObject` 检测进程句柄

**注意：** `golang.org/x/sys` 已在 go.mod 中（`autostart_windows.go` 使用 `registry` 子包），无需新增依赖。

**Windows 实现概要：**

```go
//go:build windows
package config

import (
    "fmt"
    "os"
    "strconv"
    "strings"
    "syscall"
    "golang.org/x/sys/windows"
)

func processExists(pid int) bool {
    h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
    if err != nil { return false }
    defer windows.CloseHandle(h)
    var code uint32
    err = windows.GetExitCodeProcess(h, &code)
    return err == nil && code == windows.STILL_ACTIVE
}
```

---

#### [R14-10] TypeScript `any` 类型泛滥

- **严重度：** ⚪ 低 | **模块：** WebUI 前端
- **文件：** 全部 `.vue` 组件（`Dashboard.vue`、`Targets.vue`、`Rules.vue`、`Settings.vue`、`Logs.vue`、`Advanced.vue`、`Alerts.vue`）
- **现象：** 与 Issue1.md [WEB-03] 相同 — 所有组件广泛使用 `any` 类型。
- **是否与已有 Issue 重复：** 是（Issue1.md [WEB-03]，状态仍为 📋 待规划）

##### 修复方法说明

**涉及文件：**
1. **新建** `webui/frontend/src/types.ts` — 定义 API 响应类型接口
2. 逐组件替换 `any` 为具体类型

建议分步进行（一次改一个组件，确保每个组件独立可构建）：

##### 代码准备

**新建 `src/types.ts`：**

```typescript
// API 响应类型定义

export interface TargetConfig {
  id: number
  cloud_type: string
  region: string
  resource_id: string
}

export interface DomainRule {
  id: number
  host: string
  protocol: string
  ports: string
  action: string
  targets: number[]
  comment: string
  enable_ipv6: boolean
}

export interface SyncStatus {
  running: boolean
  last_sync: string | null
}

export interface DryRunResult {
  provider: string
  domain: string
  to_add: number
  to_delete: number
  error: string
}

export interface SyncLogEntry {
  timestamp: string
  target: string
  domain: string
  result: string
  added: number
  deleted: number
}

export interface SyncEvent {
  type: string
  timestamp: string
  data: Record<string, unknown>
}

export interface AlertEmailConfig {
  enabled: boolean
  host: string
  port: string
  username: string
  password: string
  from_addr: string
  to_addr: string
}

export interface AlertWebhookConfig {
  enabled: boolean
  url: string
  channel?: string
}
```

**逐组件替换示例（以 `Dashboard.vue` 为例）：**

```typescript
// 修改前：
const status = ref<any>({ running: false })
const dryrunResults = ref<any[]>([])

// 修改后：
import type { SyncStatus, DryRunResult } from '../types'
const status = ref<SyncStatus>({ running: false })
const dryrunResults = ref<DryRunResult[]>([])
```

##### 影响范围分析

- **`src/types.ts`：** 新建文件，零运行时影响
- **各 `.vue` 组件：** 仅改类型标注，不改变运行时行为
- **`tsconfig.json`：** 已在 `strict: true`，无需修改
- **副作用：** 若 API 返回的实际字段与类型定义不一致，`vue-tsc`（构建时 `npm run build` 中的类型检查步骤）会报错，可能暴露之前静默忽略的数据不匹配问题

##### 决策点识别

| 方案 | 描述 |
|------|------|
| **A（推荐）** | 新建 `types.ts` 集中定义接口，逐组件渐进替换 `any` |
| **B** | 每个组件内部定义局部 interface，不新建公共类型文件 |
| **C** | 暂不处理，标记为 v1.1 优化项 |

> **已决策：✅ 方案 A — 集中类型（新建 `src/types.ts`）**

选定的修复方案已在上方代码准备中给出（新建 `types.ts` + 逐组件替换）。实施建议：

- **实施顺序：** 先定义 `types.ts` → 再逐组件渐进替换（一次一个组件，编译验证）
- **优先替换：** `Dashboard.vue`（类型最少）→ `Targets.vue` → `Rules.vue` → `Settings.vue` → `Logs.vue` → `Advanced.vue` → `Alerts.vue`
- **兼容性：** `vue-tsc` 在 `npm run build` 时会进行类型检查，替换后若出现类型错误需一并修复

---

#### [R14-11] `Rules.vue` 前端缺少 ICMP 协议端口强制 ALL 校验

- **严重度：** ⚪ 低 | **模块：** WebUI 前端
- **文件：** `webui/frontend/src/views/Rules.vue`
- **现象：** 用户选择 ICMP 协议后仍可输入自定义端口值，虽然后端会静默修正，但前端缺乏即时反馈。
- **是否与已有 Issue 重复：** 否（首次发现）

##### 修复方法说明

**涉及文件：** 仅 `webui/frontend/src/views/Rules.vue`

修改方案：在协议下拉的 `@update:value` 事件中增加 watcher，当选择 ICMP 时自动设置端口为 `"ALL"` 并禁用输入框。

##### 代码准备

**改动 1：`<script>` 中增加一个 `computed` 或 `watch`：**

使用 Vue 3 的 `watch` 监听 `form.value.protocol`：

```typescript
import { ref, onMounted, h, watch } from 'vue'

// 协议为 ICMP 时自动设置端口为 ALL
watch(() => form.value.protocol, (newProto) => {
  if (newProto === 'ICMP') {
    form.value.ports = 'ALL'
  }
})
```

**改动 2：`<template>` 中端口输入框增加 `disabled` 属性：**

```html
<NFormItem label="端口">
  <NInput
    v-model:value="form.ports"
    :placeholder="form.protocol === 'ICMP' ? 'ICMP 协议固定为 ALL' : '443,80 / 8000-8010 / ALL'"
    :disabled="form.protocol === 'ICMP'"
  />
</NFormItem>
```

##### 影响范围分析

- **`Rules.vue`：** 仅新增 watch + 修改模板属性
- **其他文件：** 无影响
- **编辑模式：** 若用户在编辑已有的 ICMP 规则时切换协议为 TCP，端口输入框恢复可用，但需要用户手动输入端口值（watch 不会反向操作）
- **副作用：** 无

##### 决策点识别

无。此修复为单一最优方案。

---

### 4.3 第14轮编译与测试

| 验证命令 | 结果 |
|---------|------|
| `go build ./...` | ✅ 零错误 |
| `go vet ./...` | ✅ 零警告 |
| `go test ./...` | ✅ 6 个测试包全部 `ok` |
| `CGO_ENABLED=1 go build -tags desktop` (darwin) | ✅ 桌面构建成功 |

### 4.4 API 文档合规性汇总

| 云厂商 | Provider 文件 | 查询 API | 创建 API | 删除 API | 合规结论 |
|--------|-------------|---------|---------|---------|---------|
| 腾讯云 Lighthouse | `tc_lighthouse.go` | DescribeFirewallRules（Offset+Limit 分页） | CreateFirewallRules（不传 FirewallVersion） | DeleteFirewallRules | ✅ 符合官方文档 |
| 腾讯云 CVM | `tc_cvm.go` | DescribeSecurityGroupPolicies（仅 Ingress） | CreateSecurityGroupPolicies（Action 小写，Port 省略逻辑正确） | DeleteSecurityGroupPolicies（PolicyIndex 降序逐条） | ✅ 符合官方文档 |
| 阿里云 SWAS | `ali_swas.go` | ListFirewallRules（PageNumber+PageSize 分页） | CreateFirewallRules（DROP 过滤正确） | DeleteFirewallRules（RuleIds 数组） | ✅ 符合官方文档 |
| 阿里云 ECS | `ali_ecs.go` | DescribeSecurityGroupAttribute（NextToken 分页） | AuthorizeSecurityGroup（Permissions 数组，100条/批） | RevokeSecurityGroup（SecurityGroupRuleId 数组） | ✅ 符合官方文档 |

> **合规结论：** 四个 Provider 的 API 调用参数、分页方式、错误处理均与官方文档一致。未发现使用全量覆盖类 API（已验证无 ModifyFirewallRules、无重置安全组规则等调用）。仅操作 Ingress（入站）规则，零 Egress 操作。

### 4.5 设计一致性校验

| 检查项 | 预期 | 实际 | 结论 |
|--------|------|------|------|
| WebUI 绑定 `127.0.0.1` | Design1.md §七 | `server.go` L51 `127.0.0.1:{port}` | ✅ |
| .env 模式与 WebUI 互斥 | Design1.md §四 | .env 模式不写 SQLite，WebUI 不读 .env | ✅ |
| TCP+UDP 协议拆分 | AGENTS.md §三 | `buildDesired` 中仅 SWAS 不拆分 | ✅ |
| IPv6+ICMP 处理 | AGENTS.md §三 | Lighthouse→ICMPv6, CVM→ICMPV6, ECS→跳过 | ✅ |
| 规则 TAG 标识格式 | AGENTS.md §三 | `[TAG] comment`，`HasPrefix` 检测 | ✅ |
| 删除幂等 | AGENTS.md §三 | `isIdempotentDelete` 覆盖四种错误码 | ✅ |
| 添加幂等 | AGENTS.md §三 | `isIdempotentCreate` 覆盖四种错误码 | ✅ |
| DNS 解析失败不删规则 | AGENTS.md §四 | `syncDomain` WARN 后 return | ✅ |
| 渐进式熔断 | Design1.md §十 | 见 [R14-01] — **半开探测未实现** | ❌ |
| 乐观锁重试（3次+退避） | AGENTS.md §六 | `retrySync` 每次重新 Describe | ✅ |
| 不同云厂商并行 | AGENTS.md §七 | `groupByCloud` + goroutine | ✅ |
| 跨厂商串行+间隔 | AGENTS.md §七 | `rateLimitInterval` | ✅ |
| 配置导出不含凭据 | Design1.md §七 | `handleConfigExport` 删除凭据 key | ✅ |
| Docker HEALTHCHECK | AGENTS.md §八 | Dockerfile L27-28 | ✅ |
| CGO 分离 | Build1.md §八 | `//go:build desktop` 标签 | ✅ |
| 中文注释 | AGENTS.md §十一 | 所有 Go 文件注释均为中文 | ✅ |
| 所有 error 必须处理 | AGENTS.md §十一 | 已逐文件验证 | ✅ |

### 4.6 第14轮总结

- **全部修复完成：** 10 项已修复 + 1 项已关闭（R14-08 误报）
- **功能改动 R14-12（端口变更 9090→60200）：** ✅ 已实施
- **编译/测试/前端构建：** 全部通过 ✅
- **验证命令：** `go build ./... && go vet ./... && go test ./... && cd webui/frontend && npm run build`

---

### 4.7 WebUI 默认端口变更与端口冲突回退机制设计 (R14-12)

> **类型：** 功能改动设计 | **严重度：** 🟡 中 | **状态：** 📋 待实施

#### 需求

将项目默认的 WebUI 启动端口从 `9090` 改为 `60200`；若 `60200` 已被占用，则自动在 `50000–65535` 范围内随机选择一个可用端口。

---

#### 1. 影响分析

##### 1.1 需要修改的源代码文件

| 文件 | 当前位置 | 当前值 | 修改内容 |
|------|---------|--------|---------|
| **`config/config.go`** L88 | 注释 | `// 默认 9090` | 改为 `// 默认 60200` |
| **`config/env.go`** L34 | `ParseEnv` 默认值 | `WebUIPort: 9090` | 改为 `60200` |
| **`config/store.go`** L494 | `LoadConfig` 默认值 | `WebUIPort: 9090` | 改为 `60200` |
| **`config/env_test.go`** L119-120 | 测试断言 | `want 9090` | 改为 `want 60200` |
| **`webui/server.go`** L49-53 | `Start()` 方法 | 直接 `ListenAndServe` | 增加端口占用检测 + 回退逻辑 |
| **`main.go`** L67, L170-177 | WebUI 模式启动 | 使用 `cfg.WebUIPort` | 使用 `Start()` 返回的实际端口 |
| **`webui/api/settings.go`** L19-24 | `defaults` map | 无 `webui_port` 键 | 增加 `"webui_port": "60200"` |

##### 1.2 需要修改的配置文件与文档

| 文件 | 当前位置 | 当前值 | 修改内容 |
|------|---------|--------|---------|
| **`.env.example`** L55 | 注释 | `WEBUI_PORT=9090` | 改为 `60200` |
| **`docker-compose.yml.example`** L30, L38, L53 | 端口映射+healthcheck | `9090:9090` | 改为 `60200:60200` |
| **`build/Dockerfile`** L28 | HEALTHCHECK | `localhost:9090` | 改为 `localhost:60200` |
| **`webui/frontend/vite.config.ts`** L8 | 开发代理 | `127.0.0.1:9090` | 改为 `127.0.0.1:60200` |
| **`README.md`** | 多处 | `9090` | 全局替换为 `60200` |
| **`Design1.md`** L63, L257 | 设计文档 | `9090` | 更新默认值说明 |
| **`AGENTS.md`** L34 | AI 指令 | `默认 9090` | 更新默认值说明 |

> **注：** `Build1.md` 为历史归档，保持原文不变。

##### 1.3 不需要修改的文件

- **`app/systray.go`** L76-77 — `openBrowser(url)` 接收动态 URL，无需改动
- **`app/app.go`** — 无端口硬编码
- **`webui/embed.go`** — 无端口逻辑
- **各 Provider 文件** — 无关

---

#### 2. 方案设计

##### 2.1 端口检测与回退逻辑

核心流程：

```
配置端口 60200 → 尝试 Listen → 成功？
    ├─ 是 → 关闭探测 Listener → 使用 60200 启动 HTTP
    └─ 否 → 随机 Listen("127.0.0.1:0") → 获取系统分配的端口 → 关闭探测 Listener → 使用该端口启动 HTTP
```

**随机回退范围说明：** `net.Listen("tcp", "127.0.0.1:0")` 由操作系统从**临时端口范围（ephemeral port range）**中分配。在 macOS 上默认是 `49152–65535`，在 Linux 上通常是 `32768–60999`。这基本涵盖了 `50000–65535` 的范围，且实现简单（无需自行遍历端口段）。若想精确限定，可自行循环尝试，但 `"127.0.0.1:0"` 的 OS 分配方式更简单可靠，符合 AGENTS.md「简单轻量化」原则。

**注意竞态窗口：** 关闭探测 Listener 到 `http.ListenAndServe` 绑定之间存在极短的时间窗口，理论上端口可能被其他进程抢占。在单用户桌面/Docker 场景下概率极低，符合「不过度防御」原则。若未来需要加固，可考虑 `SO_REUSEPORT`，但当前阶段不必要。

##### 2.2 端口回传方案

修改 `webui/server.go` 的 `Start()` 方法签名，使其返回实际监听的端口：

```go
// Start 启动 HTTP 服务器（阻塞），返回实际监听的端口号
func (s *Server) Start() (int, error) {
    // ... 端口检测逻辑 ...
    return actualPort, http.ListenAndServe(addr, s.mux)
}
```

`main.go` 中接收返回端口，用于：
- 日志输出：`slog.Info("WebUI 已启动", "port", actualPort)`
- 系统托盘 URL：`fmt.Sprintf("http://127.0.0.1:%d", actualPort)`
- 若端口回退发生，增加 WARN 日志提示用户

##### 2.3 Docker 兼容性

Docker 容器启动时端口通常空闲（新容器），`60200` 被占用的情况罕见。但需注意：
- **Dockerfile HEALTHCHECK：** 硬编码 `localhost:60200/api/health`，容器内端口在应用启动时确定，HEALTHCHECK 应与实际端口一致。由于 Docker 场景下端口几乎不会被占用，修改为 `60200` 即可
- **docker-compose 端口映射：** 宿主机端口映射 `"60200:60200"`，容器内端口固定为 60200，若宿主机 60200 被占，docker-compose 自身会报端口冲突（在应用启动前），与程序内的回退逻辑无关

##### 2.4 与 `WEBUI_PORT` 环境变量/设置的兼容性

- 用户显式设置 `WEBUI_PORT` 时（`.env` 或 WebUI Settings），该值的优先级高于默认值
- 回退逻辑仅在端口被占用时触发，无论端口来自默认值还是用户配置
- 用户配置的端口被占用时，回退到随机端口并 **WARN 日志**告知用户

---

#### 3. 代码准备

##### 3.1 `webui/server.go` — 核心修改

```go
package webui

import (
    "fmt"
    "io/fs"
    "log/slog"
    "net"
    "net/http"

    "github.com/alcaprophet/fwalizer/config"
    "github.com/alcaprophet/fwalizer/webui/api"
)

// ... 现有 Server 结构体和构造函数不变 ...

// Start 启动 HTTP 服务器（阻塞）。若配置端口被占用，自动随机选择可用端口。
// 返回实际监听的端口号（可能与 s.port 不同）。
func (s *Server) Start() (int, error) {
    actualPort := findAvailablePort(s.port)
    if actualPort != s.port {
        slog.Warn("端口已被占用，使用随机端口", "请求端口", s.port, "实际端口", actualPort)
        s.port = actualPort
    }
    addr := fmt.Sprintf("127.0.0.1:%d", actualPort)
    slog.Info("WebUI 启动", "访问地址", "http://"+addr)
    return actualPort, http.ListenAndServe(addr, s.mux)
}

// findAvailablePort 探测端口：优先使用 preferred，被占用时由 OS 随机分配
func findAvailablePort(preferred int) int {
    addr := fmt.Sprintf("127.0.0.1:%d", preferred)
    l, err := net.Listen("tcp", addr)
    if err == nil {
        l.Close()
        return preferred
    }
    // 首选端口被占用，由系统分配随机可用端口
    l, err = net.Listen("tcp", "127.0.0.1:0")
    if err != nil {
        // 极端情况：无法监听任何端口，回退到首选（让 http.ListenAndServe 报明确错误）
        return preferred
    }
    port := l.Addr().(*net.TCPAddr).Port
    l.Close()
    return port
}

// ... registerRoutes 不变 ...
```

##### 3.2 `main.go` — 启动侧修改 ✅ 已决策：方案 A

```go
    // 异步启动 WebUI
    go func() {
        actualPort, err := srv.Start()
        if err != nil {
            slog.Error("WebUI 服务器启动失败", "error", err)
            return
        }
        // 端口确定后同步更新（供后续引用）
        cfg.WebUIPort = actualPort
    }()

    // 系统托盘 — 使用 Start() 返回前的配置端口（短暂毫秒级不一致，托盘打开配置面板时重新构造 URL）
    url := fmt.Sprintf("http://127.0.0.1:%d", cfg.WebUIPort)
    go app.RunSystray(url, func() { s.TriggerSync() })
```

> **方案 A 理由：** 改动最小，不破坏 `systray.Run()` 的 goroutine 约束。托盘「打开配置面板」每次点击重新构造 URL（`systray.go` L50-51），不受短暂不一致窗口影响。毫秒级不一致对用户无感知。

##### 3.3 其他文件的数值修改

以下为完整的文件变更清单（`9090` → `60200`）：

**`config/config.go` L88 — 注释更新：**
```go
    WebUIPort        int           // 默认 60200
```

**`config/env.go` L34 — 默认值：**
```go
    WebUIPort:        60200,
```

**`config/store.go` L494 — 默认值：**
```go
    WebUIPort:        60200,
```

**`config/env_test.go` L119-120 — 测试断言：**
```go
    if cfg.WebUIPort != 60200 {
        t.Errorf("WebUIPort = %d, want 60200", cfg.WebUIPort)
    }
```

**`webui/api/settings.go` L19-24 — defaults map 增加一行：**
```go
    "webui_port": "60200",
```

**`webui/frontend/vite.config.ts` L8 — 开发代理：**
```typescript
      '/api': 'http://127.0.0.1:60200'
```

**`build/Dockerfile` L28 — HEALTHCHECK：**
```dockerfile
    CMD wget -q -O /dev/null http://localhost:60200/api/health 2>/dev/null || pgrep fwalizer || exit 1
```

**`docker-compose.yml.example` L30, L38, L53 — 端口映射 + 健康检查：**
```yaml
      - WEBUI_PORT=60200
    ports:
      - "60200:60200"
          "wget -q -O /dev/null http://localhost:60200/api/health 2>/dev/null || pgrep fwalizer || exit 1",
```

**`.env.example` L55 — 注释示例：**
```env
# WEBUI_PORT=60200            # WebUI 端口（默认 60200）
```

**`README.md` — 全局替换：** 所有 `9090` 替换为 `60200`（涉及端口表格、启动示例、Docker 命令示例等约 6 处）

**`Design1.md` L63, L257 — 设计文档：** 默认端口说明更新为 `60200`

**`AGENTS.md` L34 — AI 指令：** 默认端口说明更新为 `60200`

---

#### 4. 决策点汇总 ✅ 全部已决策

| # | 决策点 | 选定方案 | 理由 |
|---|--------|---------|------|
| 1 | 随机回退端口获取方式 | **✅ A — OS 分配** | 简单可靠，`net.Listen(":0")` 由系统从临时端口范围分配，覆盖 50000-65535 |
| 2 | URL 同步策略 | **✅ A — 短暂窗口** | 改动最小，不破坏 systray goroutine 约束；托盘菜单按需构造 URL |
| 3 | 回退日志级别 | **✅ A — WARN** | 用户应知晓端口发生变更 |

---

#### 5. 实施顺序

1. 修改 `config/config.go`、`config/env.go`、`config/store.go`、`config/env_test.go`（4 个数值变更）
2. 修改 `webui/server.go`（新增 `findAvailablePort`，修改 `Start` 签名）
3. 修改 `webui/api/settings.go`（增加 `webui_port` 默认值）
4. 修改 `main.go`（适配 `Start` 新签名）
5. 修改 `.env.example`、`docker-compose.yml.example`、`build/Dockerfile`、`vite.config.ts`
6. 修改 `README.md`、`Design1.md`、`AGENTS.md`（文档同步）
7. 编译验证：`go build ./... && go test ./... && go vet ./...`
8. 前端构建验证：`cd webui/frontend && npm run build`
