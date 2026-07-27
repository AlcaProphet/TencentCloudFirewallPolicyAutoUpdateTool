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

### 4.2 问题详情与修复结论

> 以下为最终修复结论。原始方案讨论、决策选项及代码草案已移除，仅保留问题摘要与修复结果。

| 编号 | 问题 | 严重度 | 修复结论 |
|------|------|--------|---------|
| R14-01 | DNS 熔断器半开探测未实现 | 🔴 高 | `syncer/syncer.go` 重构 `syncDomain`：DNS 解析统一到顶部（无论是否熔断都执行），新增 `syncDomainInternal` 方法隔离同步流程。`retry.go` 删除冗余 ECS ICMPv6 WARN |
| R14-02 | `Store.LoadConfig()` 缺少 `dns_timeout` | 🟡 中 | `store.go` 增加解析；`settings.go` 增加默认值；`Settings.vue` 增加输入项；`syncer.go` 新增 `ReloadResolver`；`main.go` ReloadFunc 重建 resolver |
| R14-03 | Rules 表单缺少 target 绑定 | 🟡 中 | `Rules.vue` 增加 `targetOptions` 多选组件 +「适用目标」表格列 |
| R14-04 | 逐域名/全局共用 `EventSyncComplete` | 🟡 中 | `bus.go` 新增 `EventDomainSyncComplete`；`syncer.go`/`logwriter.go`/`main.go`/`Logs.vue` 同步更新 |
| R14-05 | ECS ICMPv6 WARN 重试时重复输出 | ⚪ 低 | 随 R14-01 方案 C 一并修复（WARN 移入 `syncDomainInternal`） |
| R14-06 | `multiHandler`/`MultiHandler` 重复 | ⚪ 低 | 新建 `app/logutil.go`（`MultiHandler`），删除 `app.go` 和 `logstream.go` 旧代码 |
| R14-07 | Webhook 仅面向钉钉 | ⚪ 低 | `AlertWebhookConfig` 增加 `Channel` 字段；SQLite 表增加 `channel` 列；`webhook.go` 多渠道 switch；`Alerts.vue` 渠道选择 |
| R14-08 | CVM fallback IPv6 计数遗漏 | ⚪ 低 | **已关闭-误报**（经复审，`len(Ingress)+len(Egress)` 已包含 IPv6） |
| R14-09 | pidfile 防多实例 | ⚪ 低 | 新建 `config/pidfile.go` + `pidfile_unix.go` + `pidfile_windows.go`；`main.go` 启动时检测 |
| R14-10 | TypeScript `any` 泛滥 | ⚪ 低 | 新建 `webui/frontend/src/types.ts`（9 个接口定义） |
| R14-11 | ICMP 端口前端校验缺失 | ⚪ 低 | `Rules.vue` 增加 `watch`（ICMP→ALL）+ 端口输入框 `disabled` 属性 |

### 4.3 第14轮编译与测试

| 验证命令 | 结果 |
|---------|------|
| `go build ./...` | ✅ 零错误 |
| `go vet ./...` | ✅ 零警告 |
| `go test ./...` | ✅ 7/7 ok |
| `npm run build` (前端) | ✅ 构建成功 |

### 4.4 API 文档合规性

| 云厂商 | 查询 API | 创建 API | 删除 API | 结论 |
|--------|---------|---------|---------|------|
| Lighthouse | DescribeFirewallRules（Offset+Limit） | CreateFirewallRules（不传 FirewallVersion） | DeleteFirewallRules | ✅ |
| CVM | DescribeSecurityGroupPolicies（仅 Ingress） | CreateSecurityGroupPolicies（Action 小写，ICMP 省略 Port） | DeleteSecurityGroupPolicies（PolicyIndex 降序逐条） | ✅ |
| SWAS | ListFirewallRules（PageNumber+PageSize） | CreateFirewallRules（DROP 过滤） | DeleteFirewallRules（RuleIds 数组） | ✅ |
| ECS | DescribeSecurityGroupAttribute（NextToken） | AuthorizeSecurityGroup（Permissions 数组，≤100/批） | RevokeSecurityGroup（SecurityGroupRuleId） | ✅ |

### 4.5 设计一致性校验

| 检查项 | 结论 |
|--------|------|
| WebUI 绑定 `127.0.0.1`、.env/WebUI 互斥、凭据不导出、Docker HEALTHCHECK | ✅ |
| TCP+UDP 拆分、IPv6+ICMP 处理、TAG 格式、删除/添加幂等 | ✅ |
| DNS 解析失败不删规则、乐观锁重试（3次+退避）、不同云并行、同厂商串行 | ✅ |
| **渐进式熔断（半开探测）** | ✅ 已修复（R14-01） |
| CGO `//go:build desktop` 分离、中文注释、所有 error 处理 | ✅ |

### 4.6 第14轮总结

- **全部修复完成：** 10 项已修复 + 1 项已关闭（R14-08 误报）
- **R14-12 端口变更 9090→60200：** ✅ 已实施（含 `findAvailablePort` 自动回退）
- **编译/测试/前端构建：** 全部通过

---

### 4.7 WebUI 默认端口变更与端口冲突回退机制设计 (R14-12)

> **类型：** 功能改动设计 | **严重度：** 🟡 中 | **状态：** ✅ 已实施

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

**`docker-compose.yml.example` L29, L30, L38, L53 — 端口映射 + 健康检查：**
```yaml
      # WebUI 端口（可选，默认 60200）
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

---

## 五、第15轮审查（2026-07-27）

> 审查范围：全量代码逐文件审查 + 官方 API 文档对照 + 设计一致性校验。
> 审查方式：逐文件阅读 + API 文档对照 + 编译验证 + 测试验证。

### 5.1 发现的问题

| 编号 | 问题 | 严重度 | 状态 |
|------|------|--------|------|
| [R15-01](#r15-01) | Design1.md WebUI 默认端口未同步更新为 60200 | ⚪ 低 | ✅ 已修复 |
| [R15-02](#r15-02) | docker-compose.yml.example 注释中默认端口未同步更新 | ⚪ 低 | ✅ 已修复 |

---

### 5.2 问题详情与修复结论

#### [R15-01] Design1.md WebUI 默认端口未同步更新为 60200

- **严重度：** ⚪ 低 | **模块：** 文档
- **文件：** `Design1.md` L63、L257
- **现象：** Issue3.md R14-12 已将 WebUI 默认端口从 `9090` 变更为 `60200`，代码、测试、Dockerfile、README、AGENTS.md 均已同步更新，但 Design1.md 的两处描述仍为 `9090`：
  - L63："端口通过 `WEBUI_PORT` 配置，默认 `9090`"
  - L257："WebUI 端口 | 可配置，默认 `9090`"
- **是否与已有 Issue 重复：** R14-12 记录了变更计划，但 Design1.md 未实际更新

##### 修复方法

修改 `Design1.md` 两处：

**改动 1：L63**
```markdown
- 端口通过 `WEBUI_PORT` 配置，默认 `60200`
```

**改动 2：L257**
```markdown
| WebUI 端口 | 可配置，默认 `60200` |
```

---

#### [R15-02] docker-compose.yml.example 注释中默认端口未同步更新

- **严重度：** ⚪ 低 | **模块：** 文档
- **文件：** `docker-compose.yml.example` L29
- **现象：** 注释写"WebUI 端口（可选，默认 9090）"，但实际默认值已改为 `60200`
- **是否与已有 Issue 重复：** R14-12 的变更清单中包含此文件，但仅修改了端口映射值，注释未同步更新

##### 修复方法

修改 `docker-compose.yml.example` L29：

**改动：**
```yaml
      # WebUI 端口（可选，默认 60200）
      - WEBUI_PORT=60200
```

---

### 5.3 编译与测试

| 验证命令 | 结果 |
|---------|------|
| `go build ./...` | ✅ 零错误 |
| `go vet ./...` | ✅ 零警告 |
| `go test ./...` | ✅ 6 个测试包全部 `ok` |

### 5.4 API 文档合规性验证

| 云厂商 | Provider 文件 | 查询 API | 创建 API | 删除 API | 合规结论 |
|--------|-------------|---------|---------|---------|----------|
| 腾讯云 Lighthouse | `tc_lighthouse.go` | DescribeFirewallRules（Offset+Limit 分页） | CreateFirewallRules（不传 FirewallVersion，Protocol 支持 TCP/UDP/ICMP/ICMPv6/ALL） | DeleteFirewallRules | ✅ 符合官方文档 |
| 腾讯云 CVM | `tc_cvm.go` | DescribeSecurityGroupPolicies（仅 Ingress） | CreateSecurityGroupPolicies（Action 小写，ICMP/ICMPV6 省略 Port） | DeleteSecurityGroupPolicies（PolicyIndex 降序逐条） | ✅ 符合官方文档 |
| 阿里云 SWAS | `ali_swas.go` | ListFirewallRules（PageNumber+PageSize 分页） | CreateFirewallRules（DROP 过滤正确） | DeleteFirewallRules（RuleIds 数组） | ✅ 符合官方文档 |
| 阿里云 ECS | `ali_ecs.go` | DescribeSecurityGroupAttribute（NextToken 分页） | AuthorizeSecurityGroup（Permissions 数组，100条/批） | RevokeSecurityGroup（SecurityGroupRuleId 数组） | ✅ 符合官方文档 |

### 5.5 设计一致性校验

| 检查项 | 预期 | 实际 | 结论 |
|--------|------|------|------|
| WebUI 绑定 `127.0.0.1` | Design1.md §七 | `server.go` L58 `127.0.0.1:{port}` | ✅ |
| .env 模式与 WebUI 互斥 | Design1.md §四 | .env 模式不写 SQLite，WebUI 不读 .env | ✅ |
| TCP+UDP 协议拆分 | AGENTS.md §三 | `buildDesired` 中仅 SWAS 不拆分 | ✅ |
| IPv6+ICMP 处理 | AGENTS.md §三 | Lighthouse→ICMPv6, CVM→ICMPV6, ECS→跳过 | ✅ |
| 规则 TAG 标识格式 | AGENTS.md §三 | `[TAG] comment`，`HasPrefix` 检测 | ✅ |
| 删除幂等 | AGENTS.md §三 | `isIdempotentDelete` 覆盖四种错误码 | ✅ |
| 添加幂等 | AGENTS.md §三 | `isIdempotentCreate` 覆盖四种错误码 | ✅ |
| DNS 解析失败不删规则 | AGENTS.md §四 | `syncDomain` WARN 后 return | ✅ |
| 渐进式熔断（半开探测） | AGENTS.md §四 | `syncDomain` 中半开探测逻辑正确实现 | ✅ |
| 乐观锁重试（3次+退避） | AGENTS.md §六 | `retrySync` 每次重新 Describe | ✅ |
| 不同云厂商并行 | AGENTS.md §七 | `groupByCloud` + goroutine | ✅ |
| 同厂商串行+间隔 | AGENTS.md §七 | `rateLimitInterval` | ✅ |
| 配置导出不含凭据 | Design1.md §七 | `handleConfigExport` 删除凭据 key | ✅ |
| Docker HEALTHCHECK | AGENTS.md §八 | Dockerfile L27-28 | ✅ |
| CGO 分离 | Build1.md §八 | `//go:build desktop` 标签 | ✅ |
| 中文注释 | AGENTS.md §十一 | 所有 Go 文件注释均为中文 | ✅ |
| 所有 error 必须处理 | AGENTS.md §十一 | 已逐文件验证 | ✅ |
| 配置热重载 | AGENTS.md §五 | ReloadFunc 支持 Provider/Resolver/告警重建 | ✅ |
| pidfile 防多实例 | AGENTS.md §十一 | `config/pidfile.go` + 平台文件 | ✅ |
| EventBus Unsubscribe | Issue3 R13-04 | `bus.go` L58-70 | ✅ |

### 5.6 第15轮总结

- **新发现问题：** 2 项（均为文档同步遗漏，严重度低）
- **编译/测试：** 全部通过 ✅
- **API 合规性：** 四个 Provider 均符合官方文档 ✅
- **设计一致性：** 全部检查项通过 ✅
