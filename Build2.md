# FWAlizer 问题修复与功能构建（Build2 · 历史归档）

> **⚠️ 本文档为历史归档。** 全部 11 个构建 Step 已于 2026-07-27 验收通过。
> 最终验收详情见 [Issue3.md 全量复核记录](./Issue3.md#全量复核记录)。
>
> 原始问题记录见 [Issue2.md](./Issue2.md)（已归档），设计大纲见 [Design1.md](./Design1.md)，历史构建计划见 [Build1.md](./Build1.md)，AI 编码指令见 [AGENTS.md](./AGENTS.md)。

---

## 构建进度追踪

| Step | 编号 | 内容 | 严重度 | 状态 |
|------|------|------|--------|------|
| 1 | R11-01 | `.dockerignore` 去重 | 🟡 中 | ✅ 验收通过 |
| 2 | R11-02 | 配置导入事务保护 | 🔴 高 | ✅ 验收通过 |
| 3 | R11-03 | 前端数组索引→DB ID | 🔴 高 | ✅ 验收通过 |
| 4 | R11-04 | CI/CD 前端构建步骤 | 🔴 高 | ✅ 验收通过 |
| 5 | R11-05 | .env 模式 0-based 规则过滤 | 🟡 中 | ✅ 验收通过 |
| 6 | R11-06 | 移除 `app.Run` mode 参数 | ⚪ 低 | ✅ 验收通过 |
| 7 | R11-07 | README DNS 默认值同步 | ⚪ 低 | ✅ 验收通过 |
| 8 | WEB-06 | `/advanced` + `/alerts` 页面 + 告警 API | 🟡 中 | ✅ 验收通过 |
| 9 | FEA-02 | 告警通知器接入 EventBus | 🟡 中 | ✅ 验收通过 |
| 10 | FEA-03 | CLI `backup` / `restore` | 🟡 中 | ✅ 验收通过 |
| 11 | FEA-06 | systray 同步触发 + 开机自启 | ⚪ 低 | ✅ 验收通过（**已搁置**，见 FutureDesktopDevelop.md） |

> 状态标记：☐ 未开始 / ◧ 进行中 / ✅ 验收通过
> 复核日期：2026-07-27，详见 [Issue3.md 全量复核记录](./Issue3.md#全量复核记录)

---

## 构建概要

> 原始分步构建计划（含详细代码差异、验收命令）已从本文档移除，仅保留步骤概要和关键文件清单。

| Step | 涉及文件 | 要点 |
|------|---------|------|
| 1 | `.dockerignore` | 删除 L8-19 重复条目及不存在的 `Ref/` 目录 |
| 2 | `config/store.go`、`webui/api/settings.go` | 新增 6 个 Tx 方法；事务回调使用 `tx.Exec()` |
| 3 | `Targets.vue`、`Rules.vue` | `openEdit`/`deleteTarget`/`deleteRule` 改用 `row.id` |
| 4 | `.github/workflows/docker-publish.yml`、`release.yml` | Go 编译前插入 Node.js 设置 + 前端构建 |
| 5 | `config/env.go`、`env_test.go`、`.env.example`、`README.md` | RULES targets 改为 0-based |
| 6 | `app/app.go`、`main.go` | 移除 `Run` 的 `mode Mode` 参数 |
| 7 | `README.md` | DNS 默认值 `8.8.8.8:53` → `223.5.5.5` |
| 8 | `config/config.go`、`store.go`、`alerts.go`、`Advanced.vue`、`Alerts.vue`、`main.ts`、`App.vue`、`deps.go` | 新增 alert 表+CRUD+API+前端页面+路由+菜单 |
| 9 | `main.go` | 启动时从 alert 表注册 Notifier；ReloadFunc 含热重载 Unsubscribe+重建 |
| 10 | `config/store.go`、`app/cli.go`、`main.go` | 提取 `GetDataDir()`；新增 `backup`/`restore` CLI 命令（含 `copyFile`/`cleanOldBackups`/`verifyBackup`） |
| 11 | `app/systray.go`、`systray_stub.go`、`autostart.go`、`autostart_darwin.go`、`autostart_windows.go`、`syncer/syncer.go`、`main.go` | 回调触发同步；macOS plist + Windows registry 开机自启；优雅退出 channel（**已搁置**，见 FutureDesktopDevelop.md） |

---

## 构建顺序依赖图

```
Step 1 (R11-01: .dockerignore)
  └─ Step 2 (R11-02: 事务修复, store.go + settings.go)
       └─ Step 3 (R11-03: 前端索引, .vue)
            └─ Step 4 (R11-04: CI/CD)
                 └─ Step 5 (R11-05: 0-based, env.go + 测试)
                      └─ Step 6 (R11-06: mode 参数, app.go + main.go)
                           └─ Step 7 (R11-07: README DNS)
                                └─ Step 8 (WEB-06: 页面 + alert 表/API)
                                     └─ Step 9 (FEA-02: EventBus 接入, 依赖 Step 8)
                                └─ Step 10 (FEA-03: CLI, 无强依赖 Step 8-9)
                                     └─ Step 11 (FEA-06: systray, 依赖 Step 10)
```

**关键约束：**
- Step 1–7 必须线性串行执行（修复阶段），每步完成后验证编译通过
- Step 8 → Step 9 强依赖链（alert 表/API → EventBus 注册）
- Step 10 可在 Step 7 之后任意位置执行（CLI 不依赖 WebUI）
- Step 11 依赖 Step 10（main.go 需先完成 `config.GetDataDir()` 提取）
