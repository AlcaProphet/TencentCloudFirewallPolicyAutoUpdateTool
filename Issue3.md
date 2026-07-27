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
