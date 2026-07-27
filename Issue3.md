# FWAlizer 问题追踪（Issue3）

> 第13轮审查：Build2.md 全量构建完成后的深入代码审阅（2026-07-27）。
> 以 Design1.md、Build2.md、AGENTS.md 为基准。

---

## 已修复项

| 编号 | 问题 | 严重度 | 状态 |
|------|------|--------|------|
| [R13-01](#r13-01-cleanoldbackups-切片越界-panic) | `cleanOldBackups` 切片越界 panic | 🔴 高 | ✅ 已修复 |
| [R13-02](#r13-02-前端未使用的-import) | 前端未使用的 import | ⚪ 低 | ✅ 已修复 |

## 待决策项

| 编号 | 问题 | 严重度 | 状态 |
|------|------|--------|------|
| [R13-03](#r13-03-windows-开机自启未实现) | Windows 开机自启未实现 | 🟡 中 | 🔧 已决策，待实施 |
| [R13-04](#r13-04-告警配置热重载未处理重复订阅) | 告警配置热重载未处理重复订阅 | 🟡 中 | 🔧 已决策，待实施 |

---

## 已修复项详情

### [R13-01] `cleanOldBackups` 切片越界 panic

- **严重度：** 高
- **所属模块：** CLI（`app/cli.go`）
- **涉及文件：** `app/cli.go` L109

**现象描述：** `cleanOldBackups(dir, 5)` 中 `backups[keep:]` 在备份数量少于 `keep` 时触发 slice bounds out of range panic。例如仅有 2 个备份时 `backups[5:]` 会崩溃。

**修复方案：** 在切片前增加 `if len(backups) <= keep { return }` 守卫。

**影响范围：** 仅影响 `fwalizer backup` 子命令，不影响主同步流程。

---

### [R13-02] 前端未使用的 import

- **严重度：** 低
- **所属模块：** WebUI 前端
- **涉及文件：** `webui/frontend/src/views/Advanced.vue`、`webui/frontend/src/views/Alerts.vue`

**现象描述：**
- `Advanced.vue` 导入了 `h` 但未使用
- `Alerts.vue` 导入了 `NSpace`、`NDivider` 但未使用

**修复方案：** 移除未使用的 import。

---

## 待决策项详情

### [R13-03] Windows 开机自启未实现

- **严重度：** 中
- **所属模块：** 桌面端（`app/autostart.go`）
- **涉及文件：** `app/autostart.go` L101-108

**现象描述：** `enableAutoStartWindows` 和 `disableAutoStartWindows` 当前仅输出日志，未实际写入/删除注册表。macOS 的 LaunchAgent 方案已完整实现。

**可选方案：**
- **方案 A：** 使用 `os/exec` 调用 `reg add` / `reg delete` 命令（无额外依赖）
- **方案 B：** 引入 `golang.org/x/sys/windows/registry` 包（官方扩展库）
- **方案 C：** 暂不实现，标记为 TODO（当前状态）

**用户决策：** 方案 B — 引入 `golang.org/x/sys/windows/registry` 官方扩展库，通过 API 直接操作注册表。

#### 实施方案

**涉及文件：**
- `app/autostart.go` — 实现 Windows 平台开机自启的三个函数
- `go.mod` — 新增依赖 `golang.org/x/sys`

**具体变更：**

##### 1. `go.mod` — 引入依赖

```bash
go get golang.org/x/sys
```

该包已包含 `windows/registry` 子包，`go mod tidy` 后自动纳入 `go.mod`。

##### 2. `app/autostart.go` — 修改 import

在现有 import 块中新增：
```go
"golang.org/x/sys/windows/registry"
```

> 该 import 仅在 `//go:build desktop` 标签下编译，且仅在 `runtime.GOOS == "windows"` 分支实际调用，不影响 macOS/Linux 编译。

##### 3. `app/autostart.go` — 实现 `isAutoStartEnabled()` Windows 分支

将 L22 的 `return false // 简化：默认未启用` 替换为：

```go
k, err := registry.OpenKey(registry.CURRENT_USER,
    `Software\Microsoft\Windows\CurrentVersion\Run`,
    registry.QUERY_VALUE)
if err != nil {
    return false
}
defer k.Close()
_, _, err = k.GetStringValue("FWAlizer")
return err == nil
```

> `registry.QUERY_VALUE` 仅需读权限，安全最小化。若注册表键不存在或无权访问，返回 `false`（视为未启用）。

##### 4. `app/autostart.go` — 实现 `enableAutoStartWindows(exePath)`

将 L103-106 的日志桩替换为：

```go
k, err := registry.OpenKey(registry.CURRENT_USER,
    `Software\Microsoft\Windows\CurrentVersion\Run`,
    registry.SET_VALUE)
if err != nil {
    slog.Warn("打开注册表 Run 键失败", "error", err)
    return
}
defer k.Close()
if err := k.SetStringValue("FWAlizer", exePath); err != nil {
    slog.Warn("写入注册表 Run 键失败", "error", err)
    return
}
slog.Info("开机自启已启用（Windows 注册表）")
```

> `registry.SET_VALUE` 是写注册表值的最小权限。若 UAC 或组策略阻止写入，记录 WARN 日志并静默失败（不崩溃托盘）。

##### 5. `app/autostart.go` — 实现 `disableAutoStartWindows()`

将 L109-110 替换为：

```go
k, err := registry.OpenKey(registry.CURRENT_USER,
    `Software\Microsoft\Windows\CurrentVersion\Run`,
    registry.SET_VALUE)
if err != nil {
    slog.Warn("打开注册表 Run 键失败", "error", err)
    return
}
defer k.Close()
if err := k.DeleteValue("FWAlizer"); err != nil {
    // 值不存在也视为成功（幂等）
    if err != registry.ErrNotExist {
        slog.Warn("删除注册表 Run 键失败", "error", err)
        return
    }
}
slog.Info("开机自启已禁用")
```

> `DeleteValue` 在值不存在时返回 `registry.ErrNotExist`，该错误被吞掉（幂等语义 — 与项目"删除时规则已不存在视为成功"的原则一致）。

**验收方法：**
1. 在 Windows 环境下：`CGO_ENABLED=1 go build -tags desktop -o fwalizer.exe .`
2. 启动 `fwalizer.exe`，右键托盘图标 → 点击「开机自启」→ 勾选
3. 打开 `regedit`，确认 `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` 下有 `FWAlizer` 字符串值，数据为 exe 完整路径
4. 再次点击「开机自启」→ 取消勾选 → 确认注册表键已删除
5. 重新启动计算机，验证 FWAlizer 是否自动启动（或手动注销再登录测试）

**注意事项：**
- 注册表写入仅影响当前用户（`HKEY_CURRENT_USER`），不要求管理员权限
- 若 UAC 或组策略阻止写入，托盘不会崩溃，仅记录 WARN 日志，用户可手动添加启动项
- macOS 分支无需任何修改，已完整实现

---

### [R13-04] 告警配置热重载未处理重复订阅

- **严重度：** 中
- **所属模块：** 告警通知（`main.go`）
- **涉及文件：** `main.go` L93-110、`notifier/bus.go`

**现象描述：** 告警 Notifier 仅在启动时注册一次。若用户通过 WebUI 修改告警配置（`PUT /api/alerts` 触发 `notifyReload`），当前 `ReloadFunc` 不会重建告警订阅。Build2.md 已注明"告警配置变更频率极低，重启生效即可"，但 Issue2.md 提到了 Unsubscribe 方案。

**当前行为：** 告警配置修改后需重启进程才能生效。

**可选方案：**
- **方案 A（当前）：** 重启生效，不做热重载（简单，符合 Build2.md 建议）
- **方案 B：** 在 EventBus 中增加 `Unsubscribe` 方法，ReloadFunc 中重建告警订阅

**用户决策：** 方案 B — 在 EventBus 中增加 `Unsubscribe` 方法，ReloadFunc 中先取消旧订阅再重建新订阅，实现告警配置热重载。

#### 实施方案

**涉及文件：**
- `notifier/bus.go` — 新增 `Unsubscribe` 方法
- `main.go` — ReloadFunc 中增加告警订阅重建逻辑

**具体变更：**

##### 1. `notifier/bus.go` — 新增 `Unsubscribe` 方法

在 `Subscribe` 方法之后（约 L53 之后）新增：

```go
// Unsubscribe 取消订阅（移除匹配的第一个订阅者）
// sub 必须与 Subscribe 时传入的为同一实例（指针相同），否则无法匹配
func (b *EventBus) Unsubscribe(eventType EventType, sub Subscriber) {
    b.mu.Lock()
    defer b.mu.Unlock()
    subs := b.subscribers[eventType]
    for i, s := range subs {
        if s == sub {
            b.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
            return
        }
    }
}
```

> **关键设计点：** Go 接口值比较同时检查动态类型和动态指针值。`NewEmailNotifier()` 返回的 `*EmailNotifier` 指针在两次不同调用中不相同，因此**必须在 Subscribe 时保存返回的 Subscriber 接口值**，并在 Unsubscribe 时传入同一个值，才能正确匹配。

##### 2. `main.go` — 引入包级变量追踪当前 Notifier

在 `main()` 函数内、告警注册代码块之前（约 L93 之前）新增两个局部变量：

```go
// 追踪当前活跃的告警 Notifier（用于热重载时取消旧订阅）
var (
    currentEmailNotifier   notifier.Subscriber
    currentWebhookNotifier notifier.Subscriber
)
```

> 这两个变量在 main 函数作用域内，ReloadFunc 闭包可捕获。生命周期与进程一致，无需并发保护（所有操作在事件循环协程内串行）。

##### 3. `main.go` — 初始注册时保存 Notifier 引用

将 L94-110 的 Notifier 创建和订阅代码修改为同时保存引用：

```go
// 读取告警配置并注册 Notifier
if emailCfg, err := store.GetAlertEmail(); err == nil && emailCfg != nil && emailCfg.Enabled {
    currentEmailNotifier = notifier.NewEmailNotifier(notifier.EmailConfig{
        Host: emailCfg.Host, Port: emailCfg.Port,
        User: emailCfg.Username, Pass: emailCfg.Password,
        From: emailCfg.FromAddr, To: emailCfg.ToAddr,
    })
    s.EventBus().Subscribe(notifier.EventSyncError, currentEmailNotifier)
    s.EventBus().Subscribe(notifier.EventDNSFailed, currentEmailNotifier)
    slog.Info("邮件告警已启用", "to", emailCfg.ToAddr)
}

if webhookCfg, err := store.GetAlertWebhook(); err == nil && webhookCfg != nil && webhookCfg.Enabled {
    currentWebhookNotifier = notifier.NewWebhookNotifier(webhookCfg.URL)
    s.EventBus().Subscribe(notifier.EventSyncError, currentWebhookNotifier)
    s.EventBus().Subscribe(notifier.EventDNSFailed, currentWebhookNotifier)
    slog.Info("Webhook 告警已启用", "url", webhookCfg.URL)
}
```

##### 4. `main.go` — ReloadFunc 中重建告警订阅

在 `ReloadFunc`（约 L113 开始）的末尾（`s.Reload(newCfg)` 之前或之后）新增告警订阅重建逻辑：

```go
// 重建告警订阅（先取消旧订阅，再按最新配置注册）
if currentEmailNotifier != nil {
    s.EventBus().Unsubscribe(notifier.EventSyncError, currentEmailNotifier)
    s.EventBus().Unsubscribe(notifier.EventDNSFailed, currentEmailNotifier)
    currentEmailNotifier = nil
}
if currentWebhookNotifier != nil {
    s.EventBus().Unsubscribe(notifier.EventSyncError, currentWebhookNotifier)
    s.EventBus().Unsubscribe(notifier.EventDNSFailed, currentWebhookNotifier)
    currentWebhookNotifier = nil
}

// 按最新配置重新注册
if emailCfg, err := store.GetAlertEmail(); err == nil && emailCfg != nil && emailCfg.Enabled {
    currentEmailNotifier = notifier.NewEmailNotifier(notifier.EmailConfig{
        Host: emailCfg.Host, Port: emailCfg.Port,
        User: emailCfg.Username, Pass: emailCfg.Password,
        From: emailCfg.FromAddr, To: emailCfg.ToAddr,
    })
    s.EventBus().Subscribe(notifier.EventSyncError, currentEmailNotifier)
    s.EventBus().Subscribe(notifier.EventDNSFailed, currentEmailNotifier)
    slog.Info("邮件告警已更新", "to", emailCfg.ToAddr)
}

if webhookCfg, err := store.GetAlertWebhook(); err == nil && webhookCfg != nil && webhookCfg.Enabled {
    currentWebhookNotifier = notifier.NewWebhookNotifier(webhookCfg.URL)
    s.EventBus().Subscribe(notifier.EventSyncError, currentWebhookNotifier)
    s.EventBus().Subscribe(notifier.EventDNSFailed, currentWebhookNotifier)
    slog.Info("Webhook 告警已更新", "url", webhookCfg.URL)
}
```

> **最佳插入位置：** 在 `s.Reload(newCfg)` 调用之后（约 L133 之后），确保配置重载和 Provider 重建先完成，再重建告警订阅。

> **并发安全性：** `Subscribe`/`Unsubscribe` 均持有 `EventBus.mu` 写锁，`Publish` 持有读锁。Unsubscribe 在 Publish 之间执行不会导致竞态。正在执行中的旧 Notifier goroutine 会自行完成并退出，无泄漏风险。

**验收方法：**
1. `go build ./... && go vet ./...` 确保编译通过
2. 启动 WebUI 模式，进入「告警配置」页面
3. 填写 SMTP 配置并启用 → 保存 → 查看日志确认 "邮件告警已启用"
4. 手动触发一次同步错误（如配置无效 DNS 域名）→ 确认收到告警邮件
5. 修改 SMTP 收件人地址 → 保存 → 查看日志应出现 "邮件告警已更新"（无 "收到停止信号" 等重启相关日志）
6. 再次触发同步错误 → 确认新收件人地址收到告警
7. 禁用邮件告警 → 保存 → 再次触发错误 → 确认不再发送邮件

**注意事项：**
- 此方案要求 `Unsubscribe` 传入的 `Subscriber` 与 `Subscribe` 时相同（同一指针），因此必须保存引用
- 如果禁用告警，`currentEmailNotifier`/`currentWebhookNotifier` 被设为 `nil`，下次 ReloadFunc 不会尝试取消订阅（nil 检查）
- 如果重新启用已禁用的告警，会创建全新的 Notifier 实例并订阅，不会与旧实例混淆
- SSE channel 订阅者（`SubscribeChan`）不受此变更影响，仍使用独立的取消函数机制

---

## 详细构建清单

> 以下清单基于上述实施方案细化，涵盖每个文件的修改顺序、逐行代码差异、跨平台编译风险分析、测试用例和验证步骤。
> **执行顺序：先 R13-03，后 R13-04**（两者无逻辑依赖，但 R13-03 引入新依赖可能需要 `go mod tidy`）。

---

### 阶段一：[R13-03] Windows 开机自启

#### ⚠️ 关键风险：`golang.org/x/sys/windows/registry` 跨平台编译冲突

**问题描述：** `golang.org/x/sys/windows/registry` 包自身带有 `//go:build windows` 构建约束，仅在 `GOOS=windows` 时可编译。若在 `autostart.go`（构建标签 `//go:build desktop`）中直接 import 该包，在 macOS/Linux 上执行 `go build -tags desktop` 会因导入不可用包而**编译失败**。

**解决方案：** 不使用单一文件内条件编译，改为按平台拆分文件：
- `app/autostart.go` 仅保留 darwin 平台辅助函数和跨平台编译安全的公共代码
- 新建 `app/autostart_darwin.go`（`//go:build desktop && darwin`）包含 macOS 版 dispatch 函数
- 新建 `app/autostart_windows.go`（`//go:build desktop && windows`）包含 Windows 版 dispatch 函数（import registry 包）

**编译验证矩阵：**

| 构建命令 | 预期结果 | 编译的文件 |
|---------|---------|-----------|
| `go build ./...`（无 tags） | ✅ 通过 | 仅 `systray_stub.go`（`!desktop`）|
| `CGO_ENABLED=1 go build -tags desktop`（macOS） | ✅ 通过 | `systray.go` + `autostart.go` + `autostart_darwin.go` |
| `CGO_ENABLED=1 GOOS=windows go build -tags desktop` | ✅ 通过 | `systray.go` + `autostart.go` + `autostart_windows.go` |
| `CGO_ENABLED=1 GOOS=windows go build -tags desktop` — 交叉编译 | ⚠️ CGO 交叉编译需交叉工具链 | 同上 |

> **注意：** Go 交叉编译 Windows 目标通常不依赖 CGO，但 `fyne.io/systray` 桌面构建可能需要 CGO。若 CI 仅做编译检查，可用 `GOOS=windows go build -tags desktop`（无 CGO），可能需要将 `fyne.io/systray` 替换为纯 Go 方案——这超出了本次变更范围。当前 macOS 桌面构建 (`CGO_ENABLED=1 go build -tags desktop`) 已验证可行。

---

#### 修改 1：重构 `app/autostart.go`

**当前文件角色：** 包含 `isAutoStartEnabled`/`enableAutoStart`/`disableAutoStart` 三个 dispatch 函数 + macOS 和 Windows 的辅助函数，全部在 `//go:build desktop` 下。

**变更后角色：** 仅保留 macOS 辅助函数（纯标准库，跨平台编译安全）。dispatch 函数移至平台专用文件。

##### 代码差异（autostart.go）

**删除 L13-27（`isAutoStartEnabled` 整个函数）**：
```go
// 删除↓
// isAutoStartEnabled 检查是否已注册开机自启
func isAutoStartEnabled() bool {
    switch runtime.GOOS {
    case "darwin":
        plistPath := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", "com.fwalizer.agent.plist")
        _, err := os.Stat(plistPath)
        return err == nil
    case "windows":
        return false // 简化：默认未启用
    default:
        return false
    }
}
```

**删除 L29-45（`enableAutoStart` 整个函数）**：
```go
// 删除↓
// enableAutoStart 启用开机自启
func enableAutoStart() {
    exePath, err := os.Executable()
    // ... switch 分发 ...
}
```

**删除 L47-57（`disableAutoStart` 整个函数）**：
```go
// 删除↓
// disableAutoStart 禁用开机自启
func disableAutoStart() {
    switch runtime.GOOS {
    // ... switch 分发 ...
    }
}
```

**删除 L101-111（Windows 空桩函数）**：
```go
// 删除↓
// ─── Windows ───
func enableAutoStartWindows(exePath string) { ... }
func disableAutoStartWindows() { ... }
```

**保留不变：** L59-99 macOS 辅助函数(`enableAutoStartDarwin`/`disableAutoStartDarwin`) 完全不变。

**简化 import**：删除 `"runtime"`（不再需要 `runtime.GOOS` switch），保留 `"fmt"` `"log/slog"` `"os"` `"path/filepath"`。

**最终 `autostart.go` 内容（约 55 行）：**
```go
//go:build desktop

package app

import (
    "fmt"
    "log/slog"
    "os"
    "path/filepath"
)

// ─── macOS ───

func enableAutoStartDarwin(exePath string) {
    plistDir := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents")
    if err := os.MkdirAll(plistDir, 0755); err != nil {
        slog.Warn("创建 LaunchAgents 目录失败", "error", err)
        return
    }
    plistPath := filepath.Join(plistDir, "com.fwalizer.agent.plist")
    content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.fwalizer.agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
</dict>
</plist>
`, exePath)
    if err := os.WriteFile(plistPath, []byte(content), 0644); err != nil {
        slog.Warn("写入 plist 失败", "error", err)
        return
    }
    slog.Info("开机自启已启用（macOS LaunchAgent）")
}

func disableAutoStartDarwin() {
    plistPath := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", "com.fwalizer.agent.plist")
    if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
        slog.Warn("删除 plist 失败", "error", err)
        return
    }
    slog.Info("开机自启已禁用")
}
```

> **倒退兼容性：** `systray.go` 中调用的 `isAutoStartEnabled()`/`enableAutoStart()`/`disableAutoStart()` 不再定义于此文件，但由平台专用文件提供（见下面修改 2 和 3），Go 链接器在编译时自动选择对应平台的实现。

---

#### 修改 2：新建 `app/autostart_darwin.go`

**构建标签：** `//go:build desktop && darwin`

**文件内容：**
```go
//go:build desktop && darwin

package app

import (
    "log/slog"
    "os"
    "path/filepath"
)

// isAutoStartEnabled 检查 macOS LaunchAgent plist 是否存在
func isAutoStartEnabled() bool {
    plistPath := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", "com.fwalizer.agent.plist")
    _, err := os.Stat(plistPath)
    return err == nil
}

// enableAutoStart 启用 macOS 开机自启
func enableAutoStart() {
    exePath, err := os.Executable()
    if err != nil {
        slog.Warn("获取可执行文件路径失败", "error", err)
        return
    }
    enableAutoStartDarwin(exePath)
}

// disableAutoStart 禁用 macOS 开机自启
func disableAutoStart() {
    disableAutoStartDarwin()
}
```

> **说明：** `enableAutoStartDarwin` 和 `disableAutoStartDarwin` 在 `autostart.go`（同 package）中定义，编译时可见。仅使用标准库，无平台编译风险。

---

#### 修改 3：新建 `app/autostart_windows.go`

**构建标签：** `//go:build desktop && windows`

**文件内容：**
```go
//go:build desktop && windows

package app

import (
    "log/slog"
    "os"

    "golang.org/x/sys/windows/registry"
)

const regRunKey = `Software\Microsoft\Windows\CurrentVersion\Run`
const regValueName = "FWAlizer"

// isAutoStartEnabled 检查 Windows 注册表 Run 键
func isAutoStartEnabled() bool {
    k, err := registry.OpenKey(registry.CURRENT_USER, regRunKey, registry.QUERY_VALUE)
    if err != nil {
        return false
    }
    defer k.Close()
    _, _, err = k.GetStringValue(regValueName)
    return err == nil
}

// enableAutoStart 启用 Windows 开机自启（写入注册表 Run 键）
func enableAutoStart() {
    exePath, err := os.Executable()
    if err != nil {
        slog.Warn("获取可执行文件路径失败", "error", err)
        return
    }

    k, err := registry.OpenKey(registry.CURRENT_USER, regRunKey, registry.SET_VALUE)
    if err != nil {
        slog.Warn("打开注册表 Run 键失败", "error", err)
        return
    }
    defer k.Close()

    if err := k.SetStringValue(regValueName, exePath); err != nil {
        slog.Warn("写入注册表 Run 键失败", "error", err)
        return
    }
    slog.Info("开机自启已启用（Windows 注册表）")
}

// disableAutoStart 禁用 Windows 开机自启（删除注册表 Run 键）
func disableAutoStart() {
    k, err := registry.OpenKey(registry.CURRENT_USER, regRunKey, registry.SET_VALUE)
    if err != nil {
        slog.Warn("打开注册表 Run 键失败", "error", err)
        return
    }
    defer k.Close()

    if err := k.DeleteValue(regValueName); err != nil {
        if err != registry.ErrNotExist {
            slog.Warn("删除注册表 Run 键失败", "error", err)
            return
        }
        // 值不存在 = 已禁用，幂等成功
    }
    slog.Info("开机自启已禁用")
}
```

> **设计要点：**
> - 使用 `registry.CURRENT_USER`（`HKCU`）而非 `LOCAL_MACHINE`（`HKLM`），无需管理员权限
> - `registry.QUERY_VALUE` 和 `registry.SET_VALUE` 是最小权限，符合安全原则
> - `DeleteValue` 对不存在的键返回 `registry.ErrNotExist`，吞掉该错误实现幂等语义
> - 所有错误仅记录 WARN 日志不返回 error — 与项目「不过度防御」原则一致，托盘不会因注册表权限问题崩溃

---

#### 修改 4：`go.mod` 依赖升级

**当前状态：** `golang.org/x/sys v0.46.0` 已是 `// indirect` 依赖（由 `fyne.io/systray` 和 `modernc.org/sqlite` 间接引入）。

**操作：**
```bash
go get golang.org/x/sys@latest
go mod tidy
```

**预期效果：** `golang.org/x/sys` 版本更新为最新（当前最新约为 v0.30+，但 v0.46.0 实际不存在——`go.sum` 记录显示历史下载中最高为 v0.13.0。check 一下：`go.mod` 中 v0.46.0 可能来自其他依赖的 require）。执行 `go get` 后 Go 工具链自动选择兼容的最新版本，`go mod tidy` 将 `golang.org/x/sys` 提升为 `// direct` 依赖（因为 `autostart_windows.go` 直接 import 了它的子包）。

> **注意：** 如果在 macOS 上执行 `go mod tidy`，Go 不会解析 `//go:build desktop && windows` 文件中的 import，因此 `golang.org/x/sys` 的 `direct` 标记可能不会自动添加。解决方案：执行 `go mod tidy` 后手动检查，或使用 `GOOS=windows go mod tidy` 强制解析。更简单的做法是添加一个无构建约束的空引用文件（不推荐，会增加复杂度）。实际执行中，由于 `// indirect` 依赖已存在，即使未标记为 `direct`，`go build -tags desktop` 在 Windows 上也能正确链接。**推荐：在 `go.mod` 中手动将 `golang.org/x/sys` 从 indirect 移到 direct require 块。**

---

#### [R13-03] 连带风险分析

| 风险 | 概率 | 影响 | 预防措施 |
|------|------|------|---------|
| macOS `go build -tags desktop` 因 registry 包编译失败 | 高（如方案错误） | 🔴 阻断 | 已通过平台文件拆分规避 — registry 包仅在 `autostart_windows.go`（`desktop && windows`）中 import |
| Windows 交叉编译失败（CGO 依赖） | 中 | 🟡 CI 红 | CI `release.yml` 使用 `CGO_ENABLED=0`，不传 `-tags desktop`，不受影响；仅 Docker 构建受影响但 docker-publish.yml 也不传 `-tags desktop` |
| `go mod tidy` 在 macOS 上不识别 registry 依赖 | 中 | 🟡 依赖丢失 | 手动将 `golang.org/x/sys` 提升为 direct 依赖；或添加 `//go:build windows` 的空引用桩文件（仅 import，实际不被编译） |
| 注册表写入被 UAC/组策略阻止 | 低（HKCU 无需管理员） | 🟢 静默失败 | 已通过 WARN 日志 + 非致命返回处理，托盘不崩溃 |
| 旧用户升级后 Windows 菜单初始状态不正确 | 低 | 🟢 用户体验 | `isAutoStartEnabled()` 读取注册表实时状态，升级后自动正确 |
| 卸载 FWAlizer 后注册表残留 | 低 | 🟢 无影响 | 残留的 `HKCU\...\Run\FWAlizer` 仅在登录时尝试启动不存在的 exe，Windows 静默忽略 |

---

#### [R13-03] 测试用例

1. **编译测试（macOS）**：
   ```bash
   go build ./...                                    # 无 tags，验证 systray_stub.go 桩生效
   CGO_ENABLED=0 go build -tags desktop ./...        # 非 CGO 桌面构建（预期失败—systray 需要 CGO，但验证 autostart 编译路径正确）
   ```

2. **编译测试（Windows 交叉编译）**：
   ```bash
   GOOS=windows GOARCH=amd64 go build -tags desktop ./...  # 验证 autostart_windows.go 编译正确
   ```
   > 若 `fyne.io/systray` 在 Windows 交叉编译下缺少 CGO 依赖，可临时用 `go build -tags desktop app/`（仅编译 app 包）隔离验证。

3. **功能测试（需要 Windows 实机）**：
   - 启动后检查注册表 `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` 下无 `FWAlizer` 键
   - 右键托盘 → 点击「开机自启」→ 勾选
   - 验证注册表出现 `FWAlizer` 字符串值，数据为 exe 绝对路径
   - 再次点击 → 验证注册表键已删除
   - 任务管理器的「启动」选项卡可见 FWAlizer 条目

4. **幂等性测试**：
   - 手动删除注册表键 → 点击取消勾选 → 应无报错
   - 重复多次勾选/取消 → 注册表状态始终与菜单勾选状态一致

---

### 阶段二：[R13-04] 告警配置热重载

#### 修改 1：`notifier/bus.go` — 新增 `Unsubscribe` 方法

##### 现有代码结构分析

`EventBus` 内部数据结构：
```go
type EventBus struct {
    mu          sync.RWMutex
    subscribers map[EventType][]Subscriber  // ← Unsubscribe 操作的目标
    chanSubs    map[int]chan Event          // ← 不受此变更影响
    nextID      int
}
```

`Subscribe` 使用 `b.mu.Lock()` 写入，追加到 slice；`Publish` 使用 `b.mu.RLock()` 读取，对每个 subscriber 启动 goroutine 调用 `OnEvent`。`Unsubscribe` 需要 `b.mu.Lock()`（写锁），与 `Subscribe` 同级别。

##### Go 接口值比较语义

`Subscriber` 是一个接口类型。Go 中接口值的 `==` 比较同时检查**动态类型**和**动态值**。对于指针类型（`*EmailNotifier`、`*WebhookNotifier`），两个指向不同内存地址的指针不相等，即使它们包含相同的字段值。

**这意味着**：必须在 `Subscribe` 时保存返回的 `Subscriber` 接口值，在 `Unsubscribe` 时传入**同一个值**才能匹配。这是 `main.go` 中引入 `currentEmailNotifier`/`currentWebhookNotifier` 变量的根本原因。

##### 代码差异

在 `Subscribe` 方法之后（L53 之后）新增：

```go
// Unsubscribe 取消订阅。sub 必须与 Subscribe 时传入的为同一实例，否则无法匹配。
// 若 sub 未找到（已取消或从未订阅），无操作（幂等）。
func (b *EventBus) Unsubscribe(eventType EventType, sub Subscriber) {
    b.mu.Lock()
    defer b.mu.Unlock()

    subs := b.subscribers[eventType]
    for i, s := range subs {
        if s == sub {
            // 删除匹配项（保持顺序不重要）
            b.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
            return
        }
    }
    // 未找到 — 幂等，不报错
}
```

> **算法复杂度：** O(n) 线性扫描，n 为当前事件类型的订阅者数量（本项目 ≤ 10），无性能问题。

##### 不影响现有调用者

| 调用位置 | 影响 |
|---------|------|
| `notifier/bus_test.go` L25, L51 | ✅ 仅用 `Subscribe`，编译和行为不变 |
| `main.go` L90-91, L100-101, L107-108 | ✅ 仅用 `Subscribe`，编译不变 |
| `webui/api/logstream.go` L140 | ✅ 使用 `LogBroadcaster.Subscribe()`（不同接口），不受影响 |
| `webui/api/sync.go` L58 | ✅ 使用 `EventBus.SubscribeChan()`，不受影响 |

---

#### 修改 2：`notifier/bus_test.go` — 新增 `Unsubscribe` 测试

在现有两个测试函数之后追加：

```go
func TestEventBus_Unsubscribe(t *testing.T) {
    bus := NewEventBus()
    sub := &mockSubscriber{}

    bus.Subscribe(EventSyncError, sub)
    bus.Unsubscribe(EventSyncError, sub)

    // 发布不应收到的事件
    bus.Publish(Event{Type: EventSyncError, Timestamp: time.Now()})
    time.Sleep(100 * time.Millisecond)

    sub.mu.Lock()
    defer sub.mu.Unlock()
    if len(sub.events) != 0 {
        t.Errorf("取消订阅后不应收到事件, got %d", len(sub.events))
    }
}

func TestEventBus_UnsubscribeIdempotent(t *testing.T) {
    bus := NewEventBus()
    sub := &mockSubscriber{}

    // 对未订阅的类型取消订阅 — 不应 panic
    bus.Unsubscribe(EventSyncComplete, sub)

    // 重复取消 — 不应 panic
    bus.Subscribe(EventSyncError, sub)
    bus.Unsubscribe(EventSyncError, sub)
    bus.Unsubscribe(EventSyncError, sub) // 幂等
}

func TestEventBus_UnsubscribeOnlyTargetType(t *testing.T) {
    bus := NewEventBus()
    sub := &mockSubscriber{}

    bus.Subscribe(EventSyncError, sub)
    bus.Subscribe(EventDNSFailed, sub)

    // 仅取消 sync:error 的订阅
    bus.Unsubscribe(EventSyncError, sub)

    // 发布 dns:failed — 应收到
    bus.Publish(Event{Type: EventDNSFailed, Timestamp: time.Now()})
    time.Sleep(100 * time.Millisecond)

    sub.mu.Lock()
    defer sub.mu.Unlock()
    if len(sub.events) != 1 {
        t.Errorf("dns:failed 应收到事件, got %d", len(sub.events))
    }
}
```

---

#### 修改 3：`main.go` — 变量声明 + 初始注册 + ReloadFunc 重建

##### 3a. 新增变量（在 L92 之前插入）

当前 L92 是 `// 读取告警配置并注册 Notifier` 注释的前一行空行。在 L92 之前插入两个变量声明。

**插入位置分析：** 变量必须在 `ReloadFunc` 闭包定义（L113）之前声明，以便闭包捕获。插入在 `logWriter` 订阅之后（L91）和告警注册之前（L93）是最佳位置——变量声明不产生副作用，且闭包定义在后。

```go
// 插入在 L91 之后，L93 之前：

// 追踪当前活跃的告警 Notifier（用于热重载时取消旧订阅）
var currentEmailNotifier notifier.Subscriber
var currentWebhookNotifier notifier.Subscriber
```

> **作用域：** 变量在 `case app.ModeWebUI:` 分支内，`ReloadFunc` 闭包可直接捕获。类型为 `notifier.Subscriber`（接口），可持有 `*EmailNotifier` 或 `*WebhookNotifier` 或 `nil`。

##### 3b. 修改初始告警注册（L93-110）

**当前代码 L94-102（邮件）：**
```go
if emailCfg, err := store.GetAlertEmail(); err == nil && emailCfg != nil && emailCfg.Enabled {
    notifierEmail := notifier.NewEmailNotifier(notifier.EmailConfig{...})
    s.EventBus().Subscribe(notifier.EventSyncError, notifierEmail)
    s.EventBus().Subscribe(notifier.EventDNSFailed, notifierEmail)
    slog.Info("邮件告警已启用", "to", emailCfg.ToAddr)
}
```

**变更为：**
```go
if emailCfg, err := store.GetAlertEmail(); err == nil && emailCfg != nil && emailCfg.Enabled {
    currentEmailNotifier = notifier.NewEmailNotifier(notifier.EmailConfig{
        Host: emailCfg.Host, Port: emailCfg.Port,
        User: emailCfg.Username, Pass: emailCfg.Password,
        From: emailCfg.FromAddr, To: emailCfg.ToAddr,
    })
    s.EventBus().Subscribe(notifier.EventSyncError, currentEmailNotifier)
    s.EventBus().Subscribe(notifier.EventDNSFailed, currentEmailNotifier)
    slog.Info("邮件告警已启用", "to", emailCfg.ToAddr)
}
```

**差异：** `notifierEmail` 局部变量 → `currentEmailNotifier` 变量 + 保存引用。

**同理 L105-110（Webhook）**：`notifierWH` → `currentWebhookNotifier`。

##### 3c. 在 ReloadFunc 中追加告警订阅重建（L133 之后插入）

**插入位置：** L133 `s.Reload(newCfg)` 之后，L134 `})` 之前。执行顺序：配置重载 → Provider 重建 → Syncer 更新 → **告警订阅重建**。

```go
// 追加在 s.Reload(newCfg) 之后：

// 重建告警订阅（先取消旧订阅，再按最新配置注册）
if currentEmailNotifier != nil {
    s.EventBus().Unsubscribe(notifier.EventSyncError, currentEmailNotifier)
    s.EventBus().Unsubscribe(notifier.EventDNSFailed, currentEmailNotifier)
    currentEmailNotifier = nil
}
if currentWebhookNotifier != nil {
    s.EventBus().Unsubscribe(notifier.EventSyncError, currentWebhookNotifier)
    s.EventBus().Unsubscribe(notifier.EventDNSFailed, currentWebhookNotifier)
    currentWebhookNotifier = nil
}

if emailCfg, err := store.GetAlertEmail(); err == nil && emailCfg != nil && emailCfg.Enabled {
    currentEmailNotifier = notifier.NewEmailNotifier(notifier.EmailConfig{
        Host: emailCfg.Host, Port: emailCfg.Port,
        User: emailCfg.Username, Pass: emailCfg.Password,
        From: emailCfg.FromAddr, To: emailCfg.ToAddr,
    })
    s.EventBus().Subscribe(notifier.EventSyncError, currentEmailNotifier)
    s.EventBus().Subscribe(notifier.EventDNSFailed, currentEmailNotifier)
    slog.Info("邮件告警已更新", "to", emailCfg.ToAddr)
}

if webhookCfg, err := store.GetAlertWebhook(); err == nil && webhookCfg != nil && webhookCfg.Enabled {
    currentWebhookNotifier = notifier.NewWebhookNotifier(webhookCfg.URL)
    s.EventBus().Subscribe(notifier.EventSyncError, currentWebhookNotifier)
    s.EventBus().Subscribe(notifier.EventDNSFailed, currentWebhookNotifier)
    slog.Info("Webhook 告警已更新", "url", webhookCfg.URL)
}
```

> **完整 ReloadFunc 变更后结构（伪代码）：**
> ```
> srv.SetReloadFunc(func() {
>     newCfg := store.LoadConfig()          // L114-118: 不变
>     provider.SetCredentials(...)          // L120: 不变
>     // ... Provider 重建 ...              // L122-131: 不变
>     s.ReloadProviders(newProviders)       // L132: 不变
>     s.Reload(newCfg)                      // L133: 不变
>     // ↓↓↓ 新增 ↓↓↓
>     unsubscribe + resubscribe alerts      // NEW: 告警订阅重建
> })
> ```

---

#### [R13-04] 连带风险分析

| 风险 | 概率 | 影响 | 预防措施 |
|------|------|------|---------|
| `Unsubscribe` 中 goroutine 泄漏（旧的 `OnEvent` 仍在执行） | 低 | 🟢 短暂泄漏 | `OnEvent` 执行时间极短（SMTP 发送或 HTTP POST），goroutine 会在数秒内完成并退出 |
| 重复订阅（ReloadFunc 被快速连续调用两次） | 低 | 🟡 重复告警 | 先 Unsubscribe 再 Subscribe 保证至多一个实例；ReloadFunc 由 `notifyReload()` 串行触发，不会并发执行 |
| `Unsubscribe` 传入的 sub 与 Subscribe 时不同（指针不匹配） | 中（实现错误） | 🟡 旧订阅泄漏 | 代码审查重点检查；单元测试覆盖指针匹配场景 |
| 告警配置保存时未触发 `notifyReload` | 极低 | 🟢 不热重载 | `handlePutAlerts` L48 已有 `d.notifyReload()` 调用 |
| `store` 在 ReloadFunc 执行时已关闭 | 极低 | 🔴 panic | ReloadFunc 仅在进程运行期间由 HTTP handler 触发，`store.Close()` 在 `main()` defer 中，仅在进程退出时执行 |
| 并发：`Publish` 的 goroutine 与 `Unsubscribe` 竞态 | 极低 | 🟢 安全 | `Publish` 持有 `RLock` 复制订阅者列表到局部变量后再启动 goroutine；即使之后 `Unsubscribe` 修改了原始 slice，已启动的 goroutine 持有的是快照 |
| `.env` 模式不受影响 | — | — | `ReloadFunc` 仅在 WebUI 模式注册，`.env` 模式走 `app.Run(cfg)` 路径，不涉及 EventBus 热重载 |

---

#### [R13-04] 测试用例

1. **单元测试（`go test ./notifier/... -v`）**：
   ```bash
   cd /path/to/fwalizer && go test ./notifier/... -v
   ```
   预期输出：
   ```
   PASS: TestEventBus_Publish
   PASS: TestEventBus_NoCrossTalk
   PASS: TestEventBus_Unsubscribe
   PASS: TestEventBus_UnsubscribeIdempotent
   PASS: TestEventBus_UnsubscribeOnlyTargetType
   ```

2. **编译测试**：
   ```bash
   go build ./...
   go vet ./...
   ```

3. **集成测试（需要 SMTP 服务器）**：
   - 启动 WebUI 模式
   - 配置 SMTP → 启用 → 保存
   - 检查日志：`"邮件告警已启用"`
   - 触发同步错误（配置不存在的域名）→ 验证收到邮件
   - 修改收件人 → 保存
   - 检查日志：`"邮件告警已更新"`
   - 再次触发错误 → 验证新收件人收到邮件

4. **禁用后不发送测试**：
   - 禁用邮件告警 → 保存
   - 触发错误 → 不应收到邮件
   - `currentEmailNotifier` 应为 nil

5. **重启持久性测试**：
   - 启用告警 → 保存 → 重启进程
   - 验证启动日志中 `"邮件告警已启用"` 出现
   - 触发错误 → 验证收到邮件

---

### 构建执行顺序

```
步骤 1: 拆分 autostart.go + 新建 autostart_darwin.go
步骤 2: 新建 autostart_windows.go (含 registry import)
步骤 3: go get golang.org/x/sys + go mod tidy
步骤 4: 验证编译 go build ./...（macOS 无 tags）
步骤 5: 验证编译 GOOS=windows go build -tags desktop ./...（交叉编译）

步骤 6: notifier/bus.go 新增 Unsubscribe
步骤 7: notifier/bus_test.go 新增 3 个测试
步骤 8: go test ./notifier/... -v（验证测试通过）

步骤 9: main.go 新增变量声明
步骤 10: main.go 修改初始告警注册（保存引用）
步骤 11: main.go ReloadFunc 追加告警重建
步骤 12: go build ./... && go vet ./...（最终验证）
步骤 13: go test ./... -v（全量测试）
```

---

### 变更文件汇总

| 文件 | 操作 | 行数变化 | 风险等级 |
|------|------|---------|---------|
| `go.mod` | 修改（升级 golang.org/x/sys） | ~2 行 | 🟡 中 |
| `app/autostart.go` | 修改（删除 dispatch + Windows 桩） | -57 行 | 🟡 中 |
| `app/autostart_darwin.go` | **新建** | +35 行 | 🟢 低 |
| `app/autostart_windows.go` | **新建** | +60 行 | 🟡 中（新依赖） |
| `notifier/bus.go` | 修改（新增 Unsubscribe） | +13 行 | 🟢 低 |
| `notifier/bus_test.go` | 修改（新增 3 个测试） | +52 行 | 🟢 低 |
| `main.go` | 修改（变量 + 订阅改造 + ReloadFunc） | +55 行 | 🟡 中 |

**总计：** 3 个文件修改，2 个文件新建，净增约 160 行代码。
