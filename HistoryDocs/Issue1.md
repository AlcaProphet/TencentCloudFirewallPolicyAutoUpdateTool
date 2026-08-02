# FWAlizer 构建与深度审查问题记录（历史归档）

> ⚠️ **已存档**：本文档已移入 `HistoryDocs/`，仅作历史记录与核查参考，**不再用于构建**。当前活跃文档：编码指令 [AGENTS.md](../AGENTS.md)、设计记录 [Design4.md](../Design4.md)、构建方案 [Build5.md](../Build5.md)、问题记录 [Issue4.md](../Issue4.md)。

> 本文档归档第1-10轮审查历史问题。**第11-12轮审查及当前修复状态追踪以 [Issue2.md](./Issue2.md) 为准。**
> 以下所有状态均以 Issue2.md 第11-12轮逐文件核查结果为最终确认。

---

## 一、状态速查（以 Issue2.md 第11-12轮确认）

### 已确认修复（20项）

Issue2.md Section 14 逐文件核查确认：

| 编号 | 问题摘要 |
|------|---------|
| [DOC-01] | `.env.example` 旧格式残留 → 已修复，59行干净格式 |
| [DOC-02] | `README.md` 旧版本残留 → 已修复，352行无旧内容 |
| [DOC-04] | `firewall/` 空目录残留 → 已修复，目录不存在 |
| [COR-01] | sync:start/complete 事件未发布 → 已修复 |
| [COR-02] | 熔断 IsOpen 未跳过同步 → 已修复，含 return |
| [COR-03] | truncateDesc 缺失 SWAS 限制 → 已修复，maxLen=50 |
| [COR-04] | strVal 位置不当 → 已修复，移至 common.go |
| [COR-05] | 同步日志未写入 SQLite → 已修复，StoreLogWriter 已订阅 |
| [COR-06] | LoadConfig 缺少配置项 → 已修复，含 webui_port/dns_fail_threshold |
| [COR-07] | 热重载不重建 Provider → 已修复，ReloadFunc 重建 providers+凭据 |
| [FEA-01] | getDataDir 未按平台区分 → 已修复，FWALIZER_DATA_DIR + 平台路径 |
| [FEA-07] | WebUI 缺少凭据配置 → 已修复，Settings.vue 含凭据输入框 |
| [DKR-01] | 根 Dockerfile 冗余 → 已修复，已删除 |
| [DKR-02] | Dockerfile 缺少前端构建 → 已修复，含 frontend-builder 阶段 |
| [DKR-03] | 缺少 WORKDIR → 已修复，WORKDIR /app |
| [BLD-02] | docker-publish.yml 重复 → 已修复，85行单文档 |
| [BLD-05] | Makefile 缺少 frontend → 已修复，build 依赖 frontend 目标 |
| [WEB-07] | JSON 字段命名不一致 → 已修复，前端统一 snake_case |
| [WEB-08] | 三个页面空白 → 已修复，字段名 + NMessageProvider + nil slice |
| [DSC-06] | Docker 数据目录耦合 → 已修复，FWALIZER_DATA_DIR + compose 示例 |

### 仍待修复（已迁移至 Issue2.md 跟踪）

| Issue1 编号 | Issue2 对应 | 问题 | 严重度 |
|---|---|---|---|
| [DOC-03] | [R11-01](./Issue2.md#r11-01-dockerignore-存在严重重复条目doc-03-修复未生效) | .dockerignore 重复条目 | 中 |
| [WEB-02] | [R11-02](./Issue2.md#r11-02-handleconfigimport-事务未实际保护-store-操作) | 配置导入事务无效 | 高 |
| [WEB-01] | [R11-03](./Issue2.md#r11-03-前端-targetsvue--rulesvue-仍使用数组索引代替数据库-id) | 前端数组索引代替 DB ID | 高 |
| [BLD-01]/[BLD-03]/[BLD-04] | [R11-04](./Issue2.md#r11-04-cicd-流水线缺少前端构建步骤bld-03bld-04-确认仍存在) | CI/CD 缺少前端构建 | 高 |
| — | [R11-05](./Issue2.md#r11-05-env-模式-apprun-使用数组索引导致规则过滤不匹配) | .env 模式规则过滤不匹配 | 中 |
| [DSC-01] | [R11-06](./Issue2.md#r11-06-apprun-的-mode-参数未被使用) | app.Run mode 参数未使用 | 低 |
| — | [R11-07](./Issue2.md#r11-07-readme-dns-默认值与代码不一致) | README DNS 默认值不一致 | 低 |

### 仍待规划（Issue2.md Section 16 / R11-08 跟踪）

| 编号 | 问题 | 严重度 |
|------|------|--------|
| [WEB-06] | 前端缺少 /advanced、/alerts 页面 | 中 |
| [FEA-02] | 告警通知器未接入 EventBus | 中 |
| [FEA-03] | CLI 缺少 backup/restore | 中 |
| [FEA-06] | systray 缺少同步触发和开机自启（**已搁置**，见 FutureDesktopDevelop.md） | 低 |

### Issue1 独有待规划/已裁定项（Issue2 未覆盖）

| 编号 | 问题 | 严重度 | 状态 |
|------|------|--------|------|
| [WEB-03] | TypeScript `any` 类型泛滥 | 低 | ✅ 已修复（R14-10：新建 `webui/frontend/src/types.ts`，9 个接口定义） |
| [WEB-04] | Dashboard 轮询而非 SSE | 低 | 📋 待规划 |
| [WEB-05] | WebUI 缺少 pidfile 防多实例 | 低 | ✅ 已修复（R14-09：`config/pidfile.go` + Unix/Windows 平台文件） |
| [COR-08] | CVM checkRuleLimit IPv6 计数不完整 | 低 | ✅ 已关闭-误报（R14-08：经复审，fallback 已完整统计） |
| [FEA-04] | README 需清理旧内容并补充 | 中 | 待规划 |
| [FEA-05] | 测试覆盖缺口（Provider/Syncer/WebUI API） | 低 | 📋 待规划 |
| [DSC-02] | CVM checkRuleLimit 重复 API 调用 | 低 | ✅ 已裁定-保持现状 |
| [DSC-03] | testConnection 复用 ClientPool | 低 | ✅ 已裁定-注入复用 |
| [DSC-04] | HTTP server 优雅退出 | 低 | ✅ 已裁定-保持现状 |
| [DSC-05] | npm vs pnpm | 低 | ✅ 已裁定-选npm（代码已实施，Design1.md/Build1.md 文档待同步） |

---

## 二、Issue1 独有项精简说明

以下仅保留关键结论，详细分析和修复方案已从本文档移除。

### [WEB-03] TypeScript `any` 类型泛滥
- **严重度：** 低 | **状态：** ✅ 已修复（R14-10）
- 所有 `.vue` 组件广泛使用 `any` 类型。已在 `webui/frontend/src/types.ts` 中定义 9 个接口（TargetConfig、DomainRule、SyncStatus 等），类型文件已就绪供后续逐组件替换。

### [WEB-04] Dashboard 使用轮询而非 SSE 获取状态
- **严重度：** 低 | **状态：** 📋 待规划
- Dashboard 使用 `setInterval(fetchStatus, 5000)` 轮询，Logs 页面已用 SSE。建议 Dashboard 也接入 SSE，保留轮询作为 fallback（延长至 30s）。5秒轮询对内部工具影响极小，待 sync 事件体系完善后统一改造。

### [WEB-05] WebUI 模式缺少 pidfile 防多实例机制
- **严重度：** 低 | **状态：** ✅ 已修复（R14-09）
- Build1.md 12.12 节规定的 pidfile 逻辑已实现。新建 `config/pidfile.go`（核心逻辑）+ `pidfile_unix.go`（`Signal(0)`）+ `pidfile_windows.go`（`OpenProcess`），`main.go` WebUI 模式启动时自动检测 pidfile 并拒绝重复运行。

### [COR-08] CVM `checkRuleLimit` fallback 路径未统计 IPv6 规则
- **严重度：** 低 | **状态：** ✅ 已关闭-误报（R14-08）
- 经复审 CVM API 文档，`DescribeSecurityGroupPolicies` 返回的 Ingress/Egress 数组各自同时包含 IPv4 和 IPv6 规则，`len()` 已完整统计。原始判断有误。

### [FEA-04] 项目 README 需清理旧内容并补充
- **严重度：** 中 | **状态：** 待规划
- [DOC-02] 已清理旧版本残留，但仍需补充项目简介、功能特性、快速开始等面向新用户的内容。

### [FEA-05] 测试覆盖缺口
- **严重度：** 低 | **状态：** 📋 待规划
- 已有测试覆盖 config/dns/internal/notifier/provider/common 六个包。缺失：四个 Provider 实现、syncer/、webui/api/、config/store.go。Provider/Syncer 层测试依赖 Mock 基础设施，可在 v1.1 建立 Mock 体系后补充。

### [DSC-02] CVM `checkRuleLimit` 重复 API 调用
- **严重度：** 低 | **裁定：** ✅ 保持现状
- `CreateRules` 和 `retrySync` 各自调用 DescribeSecurityGroupPolicies，存在重复。裁定保持现状：CVM 查询 API 配额充足（100次/秒），额外调用作为安全冗余，保留 Provider 自治边界。

### [DSC-03] `testConnection` 是否复用应用级 `ClientPool`
- **严重度：** 低 | **裁定：** ✅ 注入复用 ClientPool
- 将应用级 `ClientPool` 注入 `Deps` 统一复用，消除"测试通过但同步失败"的排查盲区。

### [DSC-04] HTTP server 优雅退出必要性
- **严重度：** 低 | **裁定：** ✅ 保持现状
- 项目为内部工具，WebUI 仅有少量管理请求，不存在长连接。实现优雅退出增加复杂度但收益极小，不符合"简单轻量化"原则。

### [DSC-05] 前端包管理器：npm vs pnpm
- **严重度：** 低 | **裁定：** ✅ 选 npm
- 代码已使用 npm（`package-lock.json` 存在），Build1.md/Design1.md 中文档仍需从 pnpm 更新为 npm。

---

## 三、已关闭问题（仅保留编号与结论）

### 构建期修复（FIX-01 ~ FIX-03）

| 编号 | 问题 | 结论 |
|------|------|------|
| [FIX-01] | `go.mod` 重复内容 | 重新写入 go.mod，单份 module + go 声明 |
| [FIX-02] | RULES 端口逗号与分隔符冲突 | 实现 `splitRuleEntries` 智能分割 |
| [FIX-03] | 深度审查批量修复（6项） | 熔断集成、EventBus 集成、热重载接通、ClientPool key 修复、删除无用 `h` 导入、fmt.Println 改 slog |

### 首轮修复（FIX-04 ~ FIX-12）

| 编号 | 问题 | 结论 |
|------|------|------|
| [FIX-04] | truncateDesc 字节截断破坏中文 | 改为 rune 切片截断 |
| [FIX-05] | fmt.Sscanf 返回值未处理 | 检查返回值，非 nil 返回 400 |
| [FIX-06] | os.MkdirAll 返回值未处理 | 添加错误检查 |
| [FIX-07] | os.UserHomeDir 错误忽略 | 提取 getDataDir()，错误时退出 |
| [FIX-08] | CVM 规则计数未使用 PolicyStatistics | 优先使用 PolicyStatistics 精确计数 |
| [FIX-09] | 半开探测失败仍递增计数器 | 已熔断时 RecordFailure 不再递增 |
| [FIX-10] | webui/api/ 端点缺失 | 6个缺失端点已全部补齐（含 SSE） |
| [FIX-11] | truncateLighthouseDesc 代码重复 | 删除，CreateRules 直接用 Syncer 层截断值 |
| [FIX-12] | go mod tidy 变更未提交 | 执行 go mod tidy 并提交 |

### 合规检查通过项

- 错误处理合规、统一 slog、中文注释、未使用全量覆盖 API、仅操作入站规则
- WebUI 绑定 127.0.0.1、凭据独立传入、配置导出剔除凭据
- 频率限制/幂等/重试/TCP+UDP 拆分/IPv6+ICMP 处理均正确
- SPA hash 路由、四云分页、协议处理全部正确

---

> **审查历史**：第1-10轮审查共发现 52 项问题，经 Issue2.md 第11-12轮逐文件复核：20 项已确认修复，7 项仍待修复（已迁移至 Issue2.md），4 项待规划（Issue2 跟踪），6 项 Issue1 独有待规划，4 项已裁定。原始详细记录已精简，当前状态以 Issue2.md 为准。
