# FWAlizer 问题追踪（历史归档）

> 第11-12轮审查（2026-07-26），以 Design1.md、Build1.md、AGENTS.md 为基准。
> **全部 11 项问题已于 Build2.md 实施完毕，最终验收结果见 [Issue3.md 全量复核记录](./Issue3.md#全量复核记录)。**
> 第1-10轮已修复项及独有待规划项见 [Issue1.md](./Issue1.md)。

---

## 一、状态速查（以 Issue3.md 全量复核确认）

### 全部已修复/已完成（11 项）

| 编号 | 问题摘要 | 严重度 | 修复结论 |
|------|---------|--------|---------|
| [R11-01](#r11-01) | `.dockerignore` 严重重复（19 行→7 行） | 🟡 中 | 删除 L8-19 重复条目及 `Ref/`，保留 7 行核心排除项 |
| [R11-02](#r11-02) | 配置导入事务回调未使用 `tx.Exec()`，回滚无效 | 🔴 高 | 新增 6 个 Tx 方法；`settings.go` 回调内替换为 Tx 版本 |
| [R11-03](#r11-03) | 前端 Targets.vue / Rules.vue 用 `index+1` 代替 DB ID | 🔴 高 | `openEdit`/`deleteTarget` 改为 `row.id`；函数签名移除 `index` 参数 |
| [R11-04](#r11-04) | CI/CD 两个 workflow 缺少 `npm ci && npm run build` | 🔴 高 | 两个 YAML 在 `go build` 前插入 `setup-node@v4` + 前端构建步骤 |
| [R11-05](#r11-05) | `.env` 模式 Rule targets 与 Provider index 语义不一致 | 🟡 中 | RULES targets 改为 0-based；同步更新 `env.go`/`env_test.go`/`.env.example`/`README.md` |
| [R11-06](#r11-06) | `app.Run` 未使用的 `mode Mode` 参数 | ⚪ 低 | 移除参数，签名改为 `Run(cfg *config.Config) error` |
| [R11-07](#r11-07) | README DNS 默认值 `8.8.8.8:53` 与代码 `223.5.5.5` 不一致 | ⚪ 低 | README L91 改为 `223.5.5.5`，注明端口自动补全 |
| [WEB-06] | 前端缺少 `/advanced`、`/alerts` 页面 | 🟡 中 | 新增 `Advanced.vue`/`Alerts.vue` + `alerts.go` API + 路由/菜单 |
| [FEA-02] | 告警通知器未接入 EventBus | 🟡 中 | `main.go` 启动时读取 alert 表注册 Notifier；ReloadFunc 含热重载 |
| [FEA-03] | CLI 缺少 `backup` / `restore` | 🟡 中 | `store.go` 提取 `GetDataDir()`；`cli.go` 新增两个 case + 辅助函数 |
| [FEA-06] | systray 缺少同步触发和开机自启 | ⚪ 低 | 回调触发同步；macOS plist + Windows registry 开机自启；优雅退出（**已搁置**，见 FutureDesktopDevelop.md） |

---

## 二、R11 修复项精简

> 以下仅保留问题发现与修复结论，详细实施计划已从本文档移除。

### [R11-01] `.dockerignore` 重复条目

- **问题：** 文件 19 行，L8-19 全部为 L1-7 的重复副本，含不存在的 `Ref/` 目录
- **关联：** Issue1.md [DOC-03] 曾误标记为已修复
- **修复：** 删除 L8-19，保留 7 行核心排除项

### [R11-02] 配置导入事务无效

- **问题：** `handleConfigImport` 用 `WithTransaction` 包裹，但回调内全部使用 `s.db.Exec()` 而非 `tx.Exec()`，回滚无法撤销清空操作，可造成不可逆数据丢失
- **关联：** Issue1.md [WEB-02]
- **修复：** `store.go` 新增 6 个 Tx 方法；`settings.go` 回调内 4 处替换

### [R11-03] 前端数组索引代替 DB ID

- **问题：** `Targets.vue` L33/L52 和 `Rules.vue` L37/L56 使用 `index + 1` 作为 API 参数
- **关联：** Issue1.md [WEB-01] 后端已完成，前端遗漏
- **修复：** `openEdit`/`deleteTarget`/`deleteRule` 使用 `row.id`；render 函数移除 `index` 参数

### [R11-04] CI/CD 缺少前端构建

- **问题：** 两个 workflow 有 `go build` 但无前端构建，`//go:embed frontend/dist` 找不到文件
- **关联：** Issue1.md [BLD-01]/[BLD-03]/[BLD-04]
- **修复：** `docker-publish.yml` 和 `release.yml` 在 Go 编译前插入 `setup-node@v4` + 前端构建

### [R11-05] `.env` 模式规则过滤不匹配

- **问题：** RULES targets 为 1-based，Provider 为 0-based，指定 target 的规则永久不匹配
- **用户决策：** 方案 B — 将 RULES targets 改为 0-based
- **修复：** `env.go` 校验 `[0, max)`；测试 + 文档同步

### [R11-06] `app.Run` mode 参数未使用

- **问题：** 函数签名含 `mode Mode` 参数但从未引用；WebUI 逻辑已迁至 `main.go`
- **关联：** Issue1.md [DSC-01]
- **修复：** `app.go` L16 移除参数；`main.go` 调用同步

### [R11-07] README DNS 默认值不一致

- **问题：** README L91 `8.8.8.8:53`，代码默认 `223.5.5.5`
- **修复：** README L91 改为 `223.5.5.5`，加注"端口 :53 自动补全"

---

## 三、第12轮合规验证结论

- **API 合规：** 零全量覆盖类 API；仅操作入站规则；零 Egress 操作
- **安全合规：** WebUI 绑定 `127.0.0.1`；凭据不导出/不导入；webui/ 零 panic
- **日志规范：** 全部使用 `log/slog`；`main.go` 启动前用 `fmt.Fprintf(os.Stderr)`
- **编译检查：** `go vet ./...` 零警告；`go build ./...` 零错误；`go test ./...` 全包 `ok`

---

> **审查历史**：第11轮发现 7 项修复问题 + 4 项待规划，第12轮合规验证确认安全/日志/编译通过。全部通过 Build2.md 11 步构建实施完毕，Issue3.md 全量复核确认验收通过。
