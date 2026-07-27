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
| [R13-03](#r13-03-windows-开机自启未实现) | Windows 开机自启未实现 | 🟡 中 | 待决策 |
| [R13-04](#r13-04-告警配置热重载未处理重复订阅) | 告警配置热重载未处理重复订阅 | 🟡 中 | 待决策 |

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

**需要决策：** 是否在此版本实现 Windows 注册表操作，以及选择哪种方案。

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

**需要决策：** 是否需要告警配置热重载能力，还是维持重启生效。

---
