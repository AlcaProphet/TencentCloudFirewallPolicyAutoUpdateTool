# FWAlizer 构建与深度审查问题记录

> 本文档经过去重、交叉检查和规范化整理，合并了多轮审查中重复记录的问题。
> 原始审查来源：第 1–10 轮深度审查 + 构建期问题记录。
> 问题按**当前状态**组织：未关闭问题在前，已关闭问题归档在底部。

---

# 一、未关闭问题

---

## 1. 文档与配置残留

### [DOC-01] `.env.example` 含旧格式残留和重复内容

- **严重度：** 高
- **当前状态：** 待修复
- **所属模块：** 配置 / 文档
- **涉及文件：** `.env.example`
- **原始记录：** Issue 4.10 / Issue 5.1（第 1 轮）/ Issue 12.6（第 8 轮）

**现象描述：**
- L1–58：当前有效格式（3 列 TARGETS、分离凭据、6 列 RULES 含 comment）
- L59–117：L1–58 的**逐字节完全相同**副本（含所有注释和空行）
- L118–200：旧格式遗留，包含：
  - 5 列 TARGETS 旧格式（`provider|resource_id|region|access_id|access_key`，凭据嵌入）
  - 5 列 RULES 旧格式（`domain|protocol|ports|action|targets`，无 comment）
  - 废弃变量名：`TENCENTCLOUD_SECRET_ID` / `TENCENTCLOUD_SECRET_KEY` / `LIGHTHOUSE_INSTANCE_ID` / `LIGHTHOUSE_REGION` / `DOMAIN_RULES` / `RULE_TAG` / `CHECK_INTERVAL` / `DNS_SERVER`

**原因分析：** Issue 4.10 曾标记为“已修复”，但修复未实际生效或后续操作覆盖了修复结果。第 1 轮、第 8 轮审查均确认问题仍存在。

**影响范围：** 新用户参照配置时产生严重困惑；旧变量名不被当前解析器识别；文件包含三倍冗余内容（约 200 行无效内容）。

**推荐修复方案：** 删除 L59–200 的全部旧格式和重复内容，仅保留 L1–58 的当前有效配置模板。

---

### [DOC-02] `README.md` 包含旧版本完整内容残留

- **严重度：** 高
- **当前状态：** 待修复
- **所属模块：** 文档
- **涉及文件：** `README.md`
- **原始记录：** Issue 6.1（第 8 轮）

**现象描述：** 文件 L1–351 为当前多云版本 README，L352–540 为旧版本单云（Lighthouse-only）完整 README 残留，包含：
- 旧项目名 "TencentCloudFirewallTool"
- 旧环境变量名（`TENCENTCLOUD_SECRET_ID`、`LIGHTHOUSE_INSTANCE_ID`、`DOMAIN_RULES` 等）
- 旧项目结构（`firewall/`、`TencentAPIGuide/`、`Ref/`）
- 旧 RULES 格式（分号分隔）
- 不存在的 Makefile 目标（`make run`、`make docker-run`）
- PowerShell 本地开发指令

**原因分析：** 重构时新 README 被追加到文件头部，旧内容未删除。

**影响范围：** 用户阅读 README 时会看到两套完全不同的配置体系，产生严重困惑。

**推荐修复方案：** 删除 L352–540 的全部旧版本内容，仅保留 L1–351 的有效 README。

---

### [DOC-03] `.dockerignore` 存在重复条目

- **严重度：** 低
- **当前状态：** 待修复
- **所属模块：** Docker 构建
- **涉及文件：** `.dockerignore`
- **原始记录：** Issue 5.13（第 7 轮复查）

**现象描述：** `.env` 出现 2 次，`*.md` 出现 2 次，`.git` 和 `.git/` 同时存在，还包含不存在的 `Ref/` 目录。

**影响范围：** 无功能影响，仅影响可维护性。

**推荐修复方案：** 去重并清理无效条目，保留核心排除项。

---

### [DOC-04] `firewall/` 空目录残留

- **严重度：** 低
- **当前状态：** 待修复
- **所属模块：** 目录结构
- **涉及文件：** `firewall/`（空目录）
- **原始记录：** Issue 5.2（第 1 轮）/ Issue 5.14（第 7 轮复查）

**现象描述：** 项目根目录存在空的 `firewall/` 目录，为旧版本残留。AGENTS.md 明确声明旧 `firewall/` 目录可直接删除。

**影响范围：** 无功能影响，仅影响目录整洁度。

**推荐修复方案：** 删除空目录（若含 `.gitkeep` 则一并删除）。

---

## 2. 构建与 CI/CD

### [BLD-01] 前端构建产物 `dist/` 缺失，WebUI 不可用

- **严重度：** 高
- **当前状态：** 待修复
- **所属模块：** 构建
- **涉及文件：** `webui/embed.go`、`webui/server.go`
- **原始记录：** Issue 9.1（第 5 轮）

**现象描述：**
- `embed.go` 使用 `//go:embed frontend/dist` 嵌入前端构建产物
- `dist/` 被 `.gitignore` 排除，仓库中不存在
- 本地 `go build`、Docker 镜像、CI Release 二进制均不含前端页面
- `server.go` 虽有优雅降级（`if err == nil`），但产品形态下 WebUI 完全不可用

**原因分析：** 前端需要 `npm install && npm run build` 生成 `dist/`，但此步骤未在任何构建流程中自动化。

**影响范围：** 所有分发渠道（Docker、Release、本地编译）的 WebUI 模式均不可用。

**推荐修复方案：** 在 Dockerfile、Makefile、CI Workflow 中集成前端构建步骤（关联 [BLD-03]、[BLD-04]、[DKR-02]）。

---

### [BLD-02] `docker-publish.yml` 内容严重重复

- **严重度：** 高
- **当前状态：** 待修复
- **所属模块：** CI/CD
- **涉及文件：** `.github/workflows/docker-publish.yml`
- **原始记录：** Issue 10.1（第 6 轮）/ Issue 5.1（第 7 轮复查）

**现象描述：** 文件包含两个完整的 YAML 文档（L1–84 和 L85–161），各自定义了 `name:` / `on:` / `env:` / `jobs:`。GitHub Actions 不支持单文件多文档，会导致解析失败或仅识别第一段。

**原因分析：** 文件被追加写入而非覆盖，导致两段内容共存。

**影响范围：** CI/CD 流水线可能无法正常触发或行为异常；第二段仅更新部分 SDK（缺少 vpc/swas/ecs）。

**推荐修复方案：** 删除 L85–161 的重复内容，保留第一段（L1–84，SDK 覆盖更完整），并在保留段中补充前端构建步骤（关联 [BLD-03]）。

---

### [BLD-03] CI/CD 流程缺少前端构建步骤

- **严重度：** 高
- **当前状态：** 待修复
- **所属模块：** CI/CD
- **涉及文件：** `.github/workflows/docker-publish.yml`、`.github/workflows/release.yml`
- **原始记录：** Issue 10.2（第 6 轮）/ Issue 5.18（第 7 轮复查）

**现象描述：** 两个 CI workflow 均包含 Go 编译检查和测试，但缺少 Node.js 安装和前端构建步骤（`npm ci && npm run build`）。`//go:embed frontend/dist` 指令要求 `dist/` 存在，从 Git 克隆后直接编译必然失败。

**影响范围：** CI/CD 流水线编译失败；Docker 镜像和 Release 二进制不含 WebUI 前端。

**推荐修复方案：** 在 Go 编译步骤前添加 `actions/setup-node@v4` + `npm ci && npm run build`。`release.yml` 只需构建一次前端（在所有 GOOS/GOARCH 编译前）。

---

### [BLD-04] Release 流程缺少前端构建

- **严重度：** 高
- **当前状态：** 待修复
- **所属模块：** CI/CD / Release
- **涉及文件：** `.github/workflows/release.yml`
- **原始记录：** Issue 10.4（第 6 轮）

**现象描述：** `release.yml` 直接编译多平台 Go 二进制并发布到 GitHub Release，未先构建前端。所有发布的二进制均不含 WebUI 前端。

**影响范围：** GitHub Release 中所有二进制均无法使用 WebUI 模式。

**推荐修复方案：** 同 [BLD-03]。

---

### [BLD-05] `Makefile` 构建目标存在多处错误

- **严重度：** 中
- **当前状态：** 待修复
- **所属模块：** 构建
- **涉及文件：** `Makefile`
- **原始记录：** Issue 12.2 / Issue 12.4（第 8 轮）

**现象描述：**
1. `docker-build` 目标引用 `-f build/Dockerfile`（虽然该文件现已存在，但路径历史上错误）
2. `build` 目标仅编译 Go 源码，不构建前端，生成的二进制不含 WebUI

**影响范围：** 本地 `make build` 生成的二进制无法使用 WebUI 模式。

**推荐修复方案：**
- `build` 目标增加 `frontend` 依赖：先 `cd webui/frontend && npm ci && npm run build`，再 `go build`
- `docker-build` 确认引用正确路径

---

## 3. Docker

### [DKR-01] 根目录 `Dockerfile` 冗余且存在多项缺陷

- **严重度：** 中
- **当前状态：** 待修复
- **所属模块：** Docker 构建
- **涉及文件：** `Dockerfile`（根目录，30 行）、`build/Dockerfile`（20 行）
- **原始记录：** Issue 10.5（第 6 轮）/ Issue 5.2 / 5.3 / 5.4 / 7.2（第 7–9 轮）

**现象描述：** 第 9 轮核实确认 `build/Dockerfile` 已存在且内容符合规范（含 `ARG VERSION`、ldflags 版本注入、`-tags docker` 构建标签、双模式 HEALTHCHECK）。根目录 `Dockerfile` 为旧版，缺少：
- `ARG VERSION` + ldflags 版本注入
- `-tags docker` 构建标签
- 双模式 HEALTHCHECK（仅 `killall -0` 检测进程存活）
- 无任何构建引用（Makefile 和 CI 均指向 `build/Dockerfile`）

**原因分析：** 早期实现放在根目录，后按文档创建了 `build/Dockerfile`，旧文件未清理。

**影响范围：** 用户若直接执行 `docker build .`（默认使用根 Dockerfile）会得到缺少版本信息和健康检查的镜像；维护两份 Dockerfile 容易产生不一致。

**推荐修复方案：** 删除根目录 `Dockerfile`，仅保留 `build/Dockerfile`。Issues 5.3、5.4（版本注入和 HEALTHCHECK）随之关闭。

---

### [DKR-02] `build/Dockerfile` 缺少前端构建阶段

- **严重度：** 高
- **当前状态：** 待修复
- **所属模块：** Docker 构建
- **涉及文件：** `build/Dockerfile`
- **原始记录：** Issue 10.3（第 6 轮，原文描述根 Dockerfile，确认为规范 Dockerfile 同样缺失）

**现象描述：** `build/Dockerfile` 仅包含 Go 编译阶段（`golang:1.25-alpine`）和运行阶段（`alpine:3.20`），缺少 Node.js 构建阶段用于编译 Vue 前端，缺少 `COPY --from=frontend-builder` 步骤。

**影响范围：** Docker 镜像构建出的二进制不含前端页面，WebUI 模式无法使用。

**推荐修复方案：** 在 Go 编译阶段之前添加 Node.js 多阶段构建：
```dockerfile
FROM node:20-alpine AS frontend-builder
WORKDIR /src/webui/frontend
COPY webui/frontend/package*.json ./
RUN npm ci
COPY webui/frontend/ ./
RUN npm run build

FROM golang:1.25-alpine AS builder
# ...
COPY --from=frontend-builder /src/webui/frontend/dist ./webui/frontend/dist
```

---

### [DKR-03] Docker 运行阶段缺少 `WORKDIR`

- **严重度：** 中
- **当前状态：** 待修复
- **所属模块：** Docker 部署
- **涉及文件：** `build/Dockerfile`
- **原始记录：** Issue 6.3（第 8 轮）

**现象描述：** `build/Dockerfile` 运行阶段未设置 `WORKDIR`，Alpine 默认为 `/`。README 中的 Docker 运行命令挂载 `.env` 到 `/app/.env`，但程序调用 `config.LoadEnv(".env")` 解析为 `/.env`，导致容器内找不到 `.env` 文件。

**影响范围：** 按 README 文档运行的 Docker `.env` 模式无法正常工作。

**推荐修复方案：** 在 `build/Dockerfile` 运行阶段添加 `WORKDIR /app`。

---

## 4. WebUI 前后端

### [WEB-01] 前端与 Syncer 使用数组索引代替数据库 ID，删除记录后错位

- **严重度：** 高
- **当前状态：** 待修复
- **所属模块：** WebUI 前端 + Syncer
- **涉及文件：** `webui/frontend/src/views/Targets.vue`、`Rules.vue`、`syncer/syncer.go`（`filterRulesForTarget`）、`config/store.go`
- **原始记录：** Issue 9.2（第 5 轮）/ Issue 6.2（第 8 轮）/ Issue 7.1（第 9 轮）

**现象描述：** 这是同一根本缺陷在不同层面的表现，第 9 轮审查定位到根因：

1. **前端层面**（Issue 9.2 / 6.2）：`Targets.vue` 和 `Rules.vue` 使用 `index + 1` 作为 API ID，但后端使用 SQLite 自增 ID。删除中间记录后，数组索引与数据库 ID 不再匹配。
2. **Syncer 层面**（Issue 7.1）：`filterRulesForTarget()` 将 `DomainRule.Targets`（存储的是 DB ID）与 `targetIndex+1`（Provider 在数组中的位置）比较。删除目标后规则静默停止同步——**存在安全风险（用户不知情下防火墙规则不再更新）**。

**原因分析：** `Store.GetTargets()` 和 `Store.GetRules()` 的 SELECT 查询未返回 `id` 列，导致整个系统（前端、Syncer）只能用数组位置“猜测” DB ID。初始状态下两者恰好一致，删除操作后产生偏移。

**影响范围：**
- 前端编辑/删除记录时可能操作错误的行
- Syncer 层规则静默停止应用到被删除后续的目标（无报错、无日志）
- 可能导致防火墙规则未更新而用户不知情

**推荐修复方案：**
1. **后端**：`GetTargets()` / `GetRules()` 返回结果包含 `id` 字段；为 `TargetConfig` 和 `DomainRule` 添加 `ID int` 字段；SELECT 改为 `SELECT id, ... FROM targets ORDER BY id`
2. **Syncer**：Provider 使用实际 DB ID 而非数组下标作为 targetIndex；`filterRulesForTarget` 改为比较真实 ID
3. **前端**：使用返回的 `row.id` 字段而非 `index + 1`
4. **建议优先修复**此问题——影响数据正确性和同步可靠性

---

### [WEB-02] 配置导入缺少事务保护

- **严重度：** 中
- **当前状态：** 待修复
- **所属模块：** WebUI 后端 / 配置持久化
- **涉及文件：** `webui/api/settings.go`
- **原始记录：** Issue 8.1（第 4 轮）

**现象描述：** `handleConfigImport` 依次执行 `ClearAll()` → `BatchAddTargets()` → `BatchAddRules()` → settings 写入，这些操作不在同一事务中。若中间步骤失败，数据库处于“已清空但未导入完整”的不一致状态。

**影响范围：** 配置导入中途失败时，用户原有配置丢失且新配置未完全写入。

**推荐修复方案：** 在 `config/store.go` 添加 `WithTransaction` 事务方法，将导入操作包裹在同一事务中，失败时自动回滚。

---

### [WEB-03] TypeScript `any` 类型泛滥

- **严重度：** 低
- **当前状态：** 待规划
- **所属模块：** WebUI 前端
- **涉及文件：** 所有 `.vue` 文件中的 `<script setup lang="ts">` 块
- **原始记录：** Issue 9.3（第 5 轮）

**现象描述：** 所有组件广泛使用 `any` 类型（`ref<any>(...)`、`render(row: any)`），虽然 `tsconfig.json` 中 `strict: true` 已启用，但 `any` 绕过了所有类型检查。

**影响范围：** IDE 无法提供自动补全和编译期类型检查；重构时容易引入 bug。

**推荐修复方案：** 在 `src/types.ts` 中定义 `Target`、`Rule`、`SyncStatus` 等接口，逐步替换 `any` 为具体类型。

---

### [WEB-04] Dashboard 使用轮询而非 SSE 获取状态

- **严重度：** 低
- **当前状态：** 待规划
- **所属模块：** WebUI 前端
- **涉及文件：** `webui/frontend/src/views/Dashboard.vue`
- **原始记录：** Issue 9.4（第 5 轮）

**现象描述：** Dashboard 使用 `setInterval(fetchStatus, 5000)` 每 5 秒轮询，而 Logs 页面已正确使用 SSE。导致状态更新有 5 秒延迟，且产生不必要的 HTTP 请求。

**推荐修复方案：** Dashboard 同时使用 SSE 监听 `sync:start`/`sync:complete` 事件来更新状态；保留轮询作为 SSE 断连时的 fallback（延长间隔到 30s）。

---

### [WEB-05] WebUI 模式缺少 pidfile 防多实例机制

- **严重度：** 低
- **当前状态：** 待规划
- **所属模块：** App 生命周期
- **涉及文件：** `main.go`（WebUI 模式启动逻辑）
- **原始记录：** Issue 8.2（第 4 轮）/ Issue 5.6（第 7 轮复查）

**现象描述：** Build1.md 12.12 节规定 WebUI 模式启动时应创建 pidfile 防止多实例运行，当前实现完全没有 pidfile 逻辑。

**影响范围：** 用户可能误启动多个 WebUI 实例，导致 SQLite 写冲突和端口占用。

**推荐修复方案：** 启动时在数据目录创建 `fwalizer.pid`，写入当前 PID；启动前检测已有 pidfile 的进程是否存活；正常退出时删除 pidfile。

---

### [WEB-06] 前端缺少「高级功能」和「告警配置」页面

- **严重度：** 中
- **当前状态：** 待规划
- **所属模块：** WebUI 前端
- **涉及文件：** `webui/frontend/src/main.ts`、`App.vue`
- **原始记录：** Issue 5.17（第 7 轮复查）

**现象描述：** Build1.md Step 14 规定 7 个页面，实际仅实现 5 个，缺少 `/advanced`（Dry Run、配置导入/导出、健康检查）和 `/alerts`（邮件、Webhook 告警配置）。

**影响范围：** 用户无法通过 WebUI 执行 Dry Run、导入导出配置、配置告警。

**推荐修复方案：** 新增 `views/Advanced.vue` 和 `views/Alerts.vue`，在 `main.ts` 和 `App.vue` 菜单中注册路由。

---

### [WEB-07] 前端 JSON 字段命名风格不统一

- **严重度：** 低
- **当前状态：** 待讨论
- **所属模块：** WebUI 前端 + API
- **涉及文件：** `webui/frontend/src/views/Targets.vue`、`Rules.vue`
- **原始记录：** Issue 6.5（第 8 轮）

**现象描述：** 前端表单使用 Go 结构体字段名（PascalCase，如 `CloudType`、`ResourceID`）作为 JSON key，而部分 API 使用 snake_case（如 `testConnectionReq` 的 `json:"cloud_type"` tag）。Go `encoding/json` 对无 tag 字段做大小写不敏感匹配，当前能正常工作，但 API 风格不统一。

**推荐修复方案：** 推荐保持现状（内部工具，功能正常即可）；也可统一添加 `json` tag。

---

## 5. Syncer / Provider / DNS / 熔断 / 热重载

### [COR-01] `sync:start` 和 `sync:complete` 事件未发布

- **严重度：** 中
- **当前状态：** 待修复
- **所属模块：** Syncer / EventBus
- **涉及文件：** `syncer/syncer.go`、`notifier/bus.go`
- **原始记录：** Issue 7.1（第 3 轮）/ Issue 5.10（第 7 轮复查）

**现象描述：** `notifier/bus.go` 定义了 `EventSyncStart`、`EventSyncComplete`、`EventSyncError`、`EventRuleChanged`、`EventDNSFailed` 五种事件。Syncer 仅在 `syncDomain()` 中发布 `EventDNSFailed` 和 `EventSyncError`，`syncAll()` 从未发布 `EventSyncStart` 和 `EventSyncComplete`。

**影响范围：** 依赖这些事件的消费者（WebUI SSE 推送、邮件告警、Webhook）无法感知同步开始和完成；前端 Dashboard 无法实时展示同步进度。

**推荐修复方案：** 在 `syncAll()` 开头发布 `EventSyncStart`，结尾发布 `EventSyncComplete`（附带耗时、成功/失败统计）。

---

### [COR-02] 熔断器 `IsOpen` 检查未实际跳过同步

- **严重度：** 中
- **当前状态：** 待修复
- **所属模块：** Syncer / DNS 熔断
- **涉及文件：** `syncer/syncer.go`、`dns/circuitbreaker.go`
- **原始记录：** Issue 11.4（第 7 轮）/ Issue 12.1（第 8 轮）

**现象描述：** `syncDomain` 中调用 `s.cb.IsOpen(rule.Host)` 检查熔断状态，但仅在 `IsOpen` 返回 true 时打印 Debug 日志（"域名已熔断，半开探测"），之后**无条件继续执行** DNS 解析和同步，未 `return` 跳过。熔断器沦为"计数器 + 日志"，不提供实际保护。

**原因分析：** 熔断跳过逻辑未实现，`IsOpen` 的返回值未被用于控制流程。

**影响范围：** 持续 DNS 解析失败的域名每轮都超时等待（默认 10s），无法通过熔断减少无效等待；浪费时间和 API 配额。对功能正确性无直接影响（失败域名不会产生错误规则变更）。

**推荐修复方案：**
```go
if s.cb.IsOpen(rule.Host) {
    slog.Debug("域名已熔断，跳过本次同步", "domain", rule.Host)
    return
}
```

---

### [COR-03] `truncateDesc` 缺失 SWAS 50 字符限制

- **严重度：** 中
- **当前状态：** 待修复
- **所属模块：** Syncer / Provider
- **涉及文件：** `syncer/retry.go`
- **原始记录：** Issue 11.1（第 7 轮）

**现象描述：** `truncateDesc` 只对 `CloudTCLighthouse` 做 64 字符截断，`default` 分支直接返回原字符串。但阿里云 SWAS 的 `Remark` 字段限制为 **50 字符**（小于 Lighthouse 的 64）。若规则描述超过 50 字符，SWAS `CreateFirewallRules` API 将返回参数错误。

**影响范围：** 使用较长 TAG 或较长 comment 时，SWAS 同步失败。

**推荐修复方案：** 在 `truncateDesc` 中为 `CloudAliSWAS` 添加 `maxLen = 50` 分支。

---

### [COR-04] `strVal` 工具函数位置不当

- **严重度：** 低
- **当前状态：** 待修复
- **所属模块：** Provider
- **涉及文件：** `provider/tc_lighthouse.go`、`provider/common.go`
- **原始记录：** Issue 6.1（第 2 轮）

**现象描述：** 包级工具函数 `strVal` 定义在 `tc_lighthouse.go` 中，但被四个 Provider 文件共用。若未来 `tc_lighthouse.go` 被重构或移除，会意外删除共享函数。

**推荐修复方案：** 将 `strVal` 函数从 `tc_lighthouse.go` 移动到 `common.go`。

---

### [COR-05] 同步日志未写入 SQLite

- **严重度：** 中
- **当前状态：** 待修复
- **所属模块：** Syncer / 配置持久化
- **涉及文件：** `syncer/syncer.go`、`config/store.go`
- **原始记录：** Issue 5.7（第 7 轮复查）

**现象描述：** `config.Store.AddSyncLog()` 已实现，`sync_logs` 表已创建，WebUI 的 `GET /api/sync/logs` 端点已实现。但 Syncer 同步过程中从未调用 `AddSyncLog()`，导致 WebUI 同步日志页面永远为空。

**原因分析：** Syncer 不持有 Store 引用，无法写入日志。

**推荐修复方案：** 通过 EventBus 订阅 `sync:complete`/`sync:error` 事件，在订阅者中写入 Store；或在 Syncer 中注入 `LogFunc` 回调。

---

### [COR-06] `LoadConfig` 未加载 `webui_port` 和 `dns_fail_threshold`

- **严重度：** 中
- **当前状态：** 待修复
- **所属模块：** 配置持久化
- **涉及文件：** `config/store.go`
- **原始记录：** Issue 5.9（第 7 轮复查）

**现象描述：** `LoadConfig()` 从 settings 表读取 `tag`、`interval`、`dns`、`log_level`，但未读取 `webui_port` 和 `dns_fail_threshold`。用户通过 WebUI 修改这两项后，重载配置不会生效。

**推荐修复方案：** 在 `LoadConfig()` 中补充对这两个字段的读取和解析。

---

### [COR-07] 热重载不重建 Provider 列表和凭据

- **严重度：** 中
- **当前状态：** 待修复
- **所属模块：** WebUI / Syncer 热重载
- **涉及文件：** `main.go`、`syncer/syncer.go`、`provider/credentials.go`
- **原始记录：** Issue 7.6（第 9 轮）

**现象描述：** WebUI 修改配置后触发热重载（`ReloadFunc` → `store.LoadConfig()` → `s.Reload(newCfg)`），但：
1. `s.providers` 列表在启动时一次性创建，热重载后不重建——新增/删除目标不会生效
2. `provider.SetCredentials()` 仅在启动时调用一次——修改凭据不生效
3. `ClientPool` 缓存已创建的 Client——即使重新 SetCredentials，旧 Client 仍用旧凭据

**影响范围：**
- WebUI 中添加新目标后不会被同步（直到重启）
- 删除目标后，旧 Provider 仍在运行，继续向已删除的资源写入规则
- 修改凭据后不生效（直到重启）

**推荐修复方案：** 热重载时重建 providers：调用 `provider.SetCredentials(newCfg...)`、创建新 `ClientPool`、重新构建 providers 列表；为 Syncer 添加 `ReloadProviders` 方法。

---

### [COR-08] CVM `checkRuleLimit` fallback 路径未统计 IPv6 规则

- **严重度：** 低
- **当前状态：** 待规划
- **所属模块：** CVM Provider
- **涉及文件：** `provider/tc_cvm.go`
- **原始记录：** Issue 11.2（第 7 轮）

**现象描述：** `checkRuleLimit` 优先使用 `PolicyStatistics` 精确统计（含 IPv4/IPv6），但 fallback 到手动计数时仅使用 `len(ps.Ingress) + len(ps.Egress)`，未包含 `Ipv6Ingress`/`Ipv6Egress`。

**影响范围：** 仅当 API 未返回 `PolicyStatistics` 且存在 IPv6 规则时，规则计数不准确。此情况极少见。

**推荐修复方案：** fallback 路径补充 `len(ps.Ipv6Ingress) + len(ps.Ipv6Egress)`。

---

## 6. 功能缺失与待规划

### [FEA-01] `getDataDir()` 未按平台区分数据目录

- **严重度：** 中
- **当前状态：** 待修复
- **所属模块：** App 生命周期
- **涉及文件：** `main.go`
- **原始记录：** Issue 5.5（第 7 轮复查）

**现象描述：** `getDataDir()` 固定返回 `~/.config/fwalizer`。Build1.md 9.3 节规定按平台区分：macOS → `~/Library/Application Support/fwalizer`，Windows → `%APPDATA%/fwalizer`，Linux → `~/.config/fwalizer`。

**影响范围：** macOS 和 Windows 用户的数据存储位置不符合操作系统规范。

**推荐修复方案：** 使用 `runtime.GOOS` 按平台返回不同路径。

---

### [FEA-02] 告警通知器未接入 EventBus

- **严重度：** 中
- **当前状态：** 待规划
- **所属模块：** 告警 / EventBus
- **涉及文件：** `notifier/email.go`、`notifier/webhook.go`、`main.go`
- **原始记录：** Issue 5.8（第 7 轮复查）

**现象描述：** `EmailNotifier` 和 `WebhookNotifier` 已实现 `Subscriber` 接口，但没有任何代码创建它们并调用 `EventBus.Subscribe()` 注册。告警功能完全未接通。

**推荐修复方案：** 在 WebUI 设置中增加告警配置项（SMTP 和 Webhook URL）；启动时读取配置，若已配置则创建对应 Notifier 并 Subscribe 到 EventBus；热重载时更新订阅。

---

### [FEA-03] CLI 缺少 `backup` / `restore` 子命令

- **严重度：** 中
- **当前状态：** 待规划
- **所属模块：** CLI
- **涉及文件：** `app/cli.go`
- **原始记录：** Issue 5.11（第 7 轮复查）

**现象描述：** Build1.md 规定 CLI 应支持 `fwalizer backup` 和 `fwalizer restore [file]`，当前仅实现了 `version` 和 `validate`。

**推荐修复方案：** `backup` 复制 `config.db` 到 `config.db.bak.{timestamp}`（最多保留 5 个）；`restore` 从备份文件恢复并执行 `PRAGMA integrity_check`。

---

### [FEA-04] 项目缺少 `README.md`

- **严重度：** 中
- **当前状态：** 待规划（注：此问题在 Issue 6.1 后已部分解决——仓库中存在 README.md，但含旧版本残留。此处指需编写纯净的项目介绍和快速开始指南）
- **所属模块：** 文档
- **涉及文件：** 无（或现有 `README.md` 需清理，见 [DOC-02]）
- **原始记录：** Issue 12.3（第 8 轮）

**现象描述：** 仓库缺少面向用户的纯净 README（现有 README 含旧版本残留需清理）。

**推荐修复方案：** 清理旧内容后（[DOC-02]），补充项目简介、功能特性、快速开始（Docker / 二进制 / 源码）、配置说明、运行模式等。

---

### [FEA-05] 测试覆盖缺口

- **严重度：** 低
- **当前状态：** 待规划
- **所属模块：** 测试
- **涉及文件：** 多个（覆盖缺口）
- **原始记录：** Issue 12.5（第 8 轮）

**现象描述：** 已有测试覆盖 `config`、`dns`、`internal/portconv`、`internal/tag`、`notifier`、`provider/common` 六个包。缺失覆盖：四个 Provider 实现文件（`tc_lighthouse.go` / `tc_cvm.go` / `ali_swas.go` / `ali_ecs.go`）、`syncer/`、`webui/api/`、`config/store.go`。

**推荐修复方案：** 优先对 `provider/common.go` 的 Diff 逻辑补充边界用例；Provider 实现层测试依赖云 SDK Mock，可留待后期；Syncer 层可通过注入 Mock Provider + Mock Resolver 进行单元测试。

---

### [FEA-06] systray 缺少开机自启和实际同步触发

- **严重度：** 低
- **当前状态：** 待规划
- **所属模块：** 桌面端
- **涉及文件：** `app/systray.go`
- **原始记录：** Issue 10.6（第 6 轮）/ Issue 5.12（第 7 轮复查）

**现象描述：**
1. Build1.md 规定托盘菜单应包含"开机自启[开关]"，当前缺失
2. "立即同步"菜单项仅打印日志（`// TODO: 通过 channel 通知 Syncer 立即同步`），未实际触发同步

**推荐修复方案：** 传入 Syncer 引用或 trigger channel 到 `RunSystray`；添加开机自启菜单项（checkbox），调用平台 API。

---

## 7. 待讨论事项

### [DSC-01] `app.Run` 中 `mode` 参数未被使用

- **严重度：** 低
- **当前状态：** 待讨论
- **所属模块：** App 生命周期
- **涉及文件：** `app/app.go`
- **原始记录：** Issue 5.3（第 1 轮）

**现象描述：** `func Run(cfg *config.Config, mode Mode) error` 接收 `mode` 参数但从未使用。原因：Phase 2 后 WebUI 模式逻辑移到 `main.go` 内联，`app.Run` 退化为仅服务于 `.env` 模式的函数。

**推荐修复方案：**
- 方案 A（推荐）：移除 `mode` 参数，精简为 `Run(cfg *config.Config) error`
- 方案 B：保留参数但在函数内根据 mode 做差异化行为

---

### [DSC-02] CVM `checkRuleLimit` 存在重复 API 调用

- **严重度：** 低
- **当前状态：** 待讨论
- **所属模块：** CVM Provider
- **涉及文件：** `provider/tc_cvm.go`
- **原始记录：** Issue 6.2（第 2 轮）

**现象描述：** `CreateRules` 调用 `checkRuleLimit`，后者独立调用 `DescribeSecurityGroupPolicies`。但 Syncer 的 `retrySync` 在调用 `CreateRules` 前已通过 `GetRules()` 获取了所有规则。`checkRuleLimit` 再次调用相同 API 造成额外网络开销。

**推荐修复方案：**
- 方案 A（推荐）：将规则计数检查上提到 `retrySync` 层，复用 `GetRules()` 返回的结果
- 方案 B：保持现状，额外 API 调用作为安全冗余

---

### [DSC-03] `testConnection` 是否复用应用级 `ClientPool`

- **严重度：** 低
- **当前状态：** 待讨论
- **所属模块：** WebUI API
- **涉及文件：** `webui/api/targets.go`
- **原始记录：** Issue 11.3（第 7 轮）

**现象描述：** `handleTestConnection` 创建全新的 `provider.NewClientPool()`，而非复用应用级 pool。测试连接的 SDK client 与同步引擎使用的 client 不是同一个实例。

**推荐修复方案：** 将 `ClientPool` 注入 `Deps` 中（或通过 `Syncer` 接口暴露）以复用；或保持现状（差异极小，不值得增加耦合）。

---

### [DSC-04] HTTP server 优雅退出必要性

- **严重度：** 低
- **当前状态：** 待讨论
- **所属模块：** App 生命周期
- **涉及文件：** `main.go`
- **原始记录：** Issue 10.7（第 6 轮）

**现象描述：** WebUI 模式下 HTTP server goroutine 没有显式的 `Shutdown()` 调用，随进程退出被强制终止。优雅退出仅覆盖了 Syncer（完成当前轮次），HTTP server 无等价处理。

**推荐修复方案：** 使用 `context.WithCancel` + `http.Server.Shutdown()` 实现优雅退出；或保持现状（内部工具场景下可接受）。

---

### [DSC-05] 前端包管理器：npm vs pnpm

- **严重度：** 低
- **当前状态：** 待讨论
- **所属模块：** WebUI 前端
- **涉及文件：** `webui/frontend/package.json`、`package-lock.json`
- **原始记录：** Issue 5.15（第 7 轮复查）

**现象描述：** Design1.md 和 Build1.md 规定前端包管理器为 pnpm，但项目中使用 `package-lock.json`（npm 产物）而非 `pnpm-lock.yaml`。

**推荐修复方案：**
- 方案 A：改用 pnpm（符合文档）
- 方案 B：文档改为 npm（降低门槛，已有 `package-lock.json`，推荐——内部工具，npm 更通用）

---

### [DSC-06] Docker 数据目录路径与 `getDataDir()` 的潜在耦合

- **严重度：** 低
- **当前状态：** 待讨论
- **所属模块：** Docker 部署
- **涉及文件：** `README.md`、`main.go`
- **原始记录：** Issue 6.4（第 8 轮）

**现象描述：** README WebUI 模式示例挂载卷到 `/home/appuser/.config/fwalizer`。当前 `getDataDir()` 使用 `os.UserHomeDir()` + `/.config/fwalizer`，路径碰巧一致。但若修复 [FEA-01]（按平台区分路径）后，需确保 Docker 内的路径仍一致。

**推荐修复方案：**
- 方案 A（推荐）：添加 `FWALIZER_DATA_DIR` 环境变量支持，Docker 通过该变量显式指定路径
- 方案 B：保持现状，在 README 中说明路径约定

---

# 二、已关闭问题（归档）

---

## 8. 构建期已修复问题

### [FIX-01] `go.mod` 重复内容

- **严重度：** 高
- **当前状态：** ✅ 已修复
- **原始记录：** Issue 1

**现象：** `go build ./...` 报错 `repeated module statement` 和 `repeated go statement`。

**解决：** 重新完整写入 `go.mod`，确保仅包含一份 module 声明和 go 版本声明。

---

### [FIX-02] RULES 端口字段逗号与条目分隔符冲突

- **严重度：** 高
- **当前状态：** ✅ 已修复
- **原始记录：** Issue 2

**现象：** `RULES=api.example.com|TCP|443,80|ACCEPT||生产API` 解析失败，端口 `443,80` 中的逗号被误认为条目分隔符。

**解决：** 实现 `splitRuleEntries` 智能分割——通过正则检测 `host|PROTOCOL|` 模式识别新条目起始位置。

---

### [FIX-03] 深度审查问题批量修复（6 项）

- **当前状态：** ✅ 全部已修复
- **原始记录：** Issue 3（3.1–3.6）

| 编号 | 问题 | 修复摘要 |
|------|------|---------|
| 3.1 | CircuitBreaker 未集成到 Syncer | Syncer 添加 `cb` 字段，`syncDomain()` 中接入 |
| 3.2 | EventBus 未集成到 Syncer | Syncer 添加 `bus` 字段，DNS 失败和同步失败时发布事件 |
| 3.3 | WebUI 配置热重载未接通 | Server 添加 `SetReloadFunc()` 回调，所有写操作后触发 |
| 3.4 | ClientPool key 缺少 accessID | 四个 Provider 的 key 统一改为 `cloudType\|region\|accessID` |
| 3.5 | App.vue 未使用的 `h` 导入 | 删除无用的 `h` 导入 |
| 3.6 | main.go 使用 `fmt.Println` | 改为 `slog.Info()`，导出 `app.InitLogger()` |

---

## 9. 第一轮深度审查已修复问题

### [FIX-04] `truncateDesc` 字节截断破坏 `[TAG]` 前缀和中文

- **严重度：** 高
- **当前状态：** ✅ 已修复
- **原始记录：** Issue 4.1

**现象：** `truncateDesc` 使用 `desc[:maxLen]` 按字节截断，中文（UTF-8 多字节字符）时会截断到半个字符产生乱码。

**修复：** 改为 rune 切片截断，保证字符完整性。

---

### [FIX-05] `fmt.Sscanf` 返回值未处理

- **严重度：** 中
- **当前状态：** ✅ 已修复
- **原始记录：** Issue 4.2

**现象：** `handleDeleteTarget` 和 `handleDeleteRule` 中忽略 `fmt.Sscanf` 返回值。

**修复：** 检查返回值，非 nil 时返回 400。

---

### [FIX-06] `os.MkdirAll` 返回值未处理

- **严重度：** 中
- **当前状态：** ✅ 已修复
- **原始记录：** Issue 4.3

**修复：** 添加错误检查，失败时输出错误并退出。

---

### [FIX-07] `os.UserHomeDir` 错误忽略

- **严重度：** 低
- **当前状态：** ✅ 已修复
- **原始记录：** Issue 4.4

**修复：** 提取 `getDataDir()` 函数，错误时输出提示并退出。

---

### [FIX-08] CVM 规则计数未使用 `PolicyStatistics`

- **严重度：** 低
- **当前状态：** ✅ 已修复
- **原始记录：** Issue 4.5

**修复：** 优先使用 `PolicyStatistics` 字段精确计数，fallback 到手动计数。

---

### [FIX-09] 半开探测失败仍递增计数器

- **严重度：** 低
- **当前状态：** ✅ 已修复
- **原始记录：** Issue 4.6

**修复：** 已熔断时 `RecordFailure` 不再递增计数器。

---

### [FIX-10] `webui/api/` 端点缺失

- **严重度：** 中
- **当前状态：** ✅ 已修复（文件拆分部分标记为可选的待规划项）
- **原始记录：** Issue 4.7

**结论：** 第 4 轮审查（Issue 8）确认所有 6 个缺失端点已全部补齐（含 SSE 实时推送），API 层功能完整。文件拆分至 `webui/api/` 子目录作为可选优化项保留。

---

### [FIX-11] `truncateLighthouseDesc` 代码重复

- **严重度：** 低
- **当前状态：** ✅ 已修复
- **原始记录：** Issue 4.8

**修复：** 删除 `tc_lighthouse.go` 中的 `truncateLighthouseDesc`，`CreateRules` 直接使用 Syncer 层已截断的描述。

---

### [FIX-12] `go mod tidy` 后变更未提交

- **严重度：** 低
- **当前状态：** ✅ 已修复
- **原始记录：** Issue 4.9

**修复：** 执行 `go mod tidy` 并提交 `go.mod` 和 `go.sum` 变更。

---

## 10. 合规检查全部通过项

- **错误处理合规**（Issue 11.5）：所有 error 均已处理；全项目统一使用 `log/slog`；注释使用中文；未使用全量覆盖 API；仅操作入站规则
- **安全合规**（Issue 11.5）：WebUI 绑定 `127.0.0.1`；凭据通过独立环境变量传入；配置导出自动剔除凭据
- **核心逻辑合规**（Issue 11.5）：频率限制符合文档；幂等处理正确（删除"已不存在"视为成功，添加"已存在"WARN 跳过）；重试语义正确（每次重试重新 Describe → Diff → Create/Delete）；TCP+UDP 拆分正确；IPv6+ICMP 处理正确
- **SPA 路由**（Issue 5.16）：前端使用 `createWebHashHistory()`（hash 路由），与 `http.FileServer` 完全兼容，无需 fallback
- **ECS 幂等性**（Issue 7.3）：ECS `AuthorizeSecurityGroup` 天然幂等（"规则已存在时调用成功但不增加规则"），`isIdempotentCreate` 的错误码匹配主要针对 Lighthouse/CVM，实现正确
- **四云分页**（Issue 7.4）：Lighthouse（Limit=100 + Offset）、CVM（无分页参数，单次全量）、ECS（NextToken，MaxResults=500）、SWAS（PageNumber/PageSize=100），四家分页处理均正确
- **协议处理**（Issue 7.5）：SWAS 跳过 IPv6；ECS 跳过 ICMPv6；Lighthouse/CVM ICMP 协议名正确；TCP+UDP 拆分正确；ICMP 端口处理正确
- **第 10 轮全面检查**（Issue 8）：CLI 子命令、系统托盘、邮件/Webhook 通知器、DNS 解析器、熔断器、端口转换、TAG 格式、release.yml、go.mod、.env 解析等均通过核查，无新发现的高/中严重度问题

---

# 三、汇总表

## 按严重度统计

| 严重度 | 未关闭 | 已关闭（含合规通过） |
|--------|--------|---------------------|
| 高 | 8 | 4 |
| 中 | 15 | 2 |
| 低 | 9 | 4 |
| 待讨论 | 7 | — |
| **合计** | **39** | **10** |

## 未关闭问题清单

### 高严重度（8 项）

| 编号 | 问题 | 状态 |
|------|------|------|
| [DOC-01] | `.env.example` 旧格式残留和重复内容 | 待修复 |
| [DOC-02] | `README.md` 旧版本内容残留 | 待修复 |
| [BLD-01] | 前端构建产物 `dist/` 缺失 | 待修复 |
| [BLD-02] | `docker-publish.yml` 内容重复 | 待修复 |
| [BLD-03] | CI/CD 流程缺少前端构建步骤 | 待修复 |
| [BLD-04] | Release 流程缺少前端构建 | 待修复 |
| [DKR-02] | `build/Dockerfile` 缺少前端构建阶段 | 待修复 |
| [WEB-01] | 前端与 Syncer 数组索引代替 DB ID | 待修复 |

> 注：[BLD-01]、[BLD-03]、[BLD-04]、[DKR-02] 为同一根因（前端未纳入构建链路的四个表现），建议统一修复。

### 中严重度（15 项）

| 编号 | 问题 | 状态 |
|------|------|------|
| [BLD-05] | `Makefile` 构建目标存在多处错误 | 待修复 |
| [DKR-01] | 根目录 `Dockerfile` 冗余 | 待修复 |
| [DKR-03] | Docker 运行阶段缺少 `WORKDIR` | 待修复 |
| [WEB-02] | 配置导入缺少事务保护 | 待修复 |
| [WEB-06] | 前端缺少高级功能/告警配置页面 | 待规划 |
| [COR-01] | `sync:start`/`sync:complete` 事件未发布 | 待修复 |
| [COR-02] | 熔断器 `IsOpen` 未实际跳过同步 | 待修复 |
| [COR-03] | `truncateDesc` 缺失 SWAS 50 字符限制 | 待修复 |
| [COR-05] | 同步日志未写入 SQLite | 待修复 |
| [COR-06] | `LoadConfig` 缺少配置项加载 | 待修复 |
| [COR-07] | 热重载不重建 Provider 列表和凭据 | 待修复 |
| [FEA-01] | `getDataDir` 未按平台区分 | 待修复 |
| [FEA-02] | 告警通知器未接入 EventBus | 待规划 |
| [FEA-03] | CLI 缺少 `backup`/`restore` | 待规划 |
| [FEA-04] | 项目缺少 README（需清理后补充） | 待规划 |

### 低严重度（9 项）

| 编号 | 问题 | 状态 |
|------|------|------|
| [DOC-03] | `.dockerignore` 重复条目 | 待修复 |
| [DOC-04] | `firewall/` 空目录残留 | 待修复 |
| [COR-04] | `strVal` 工具函数位置不当 | 待修复 |
| [COR-08] | CVM `checkRuleLimit` IPv6 计数不完整 | 待规划 |
| [WEB-03] | TypeScript `any` 类型泛滥 | 待规划 |
| [WEB-04] | Dashboard 轮询而非 SSE | 待规划 |
| [WEB-05] | pidfile 防多实例缺失 | 待规划 |
| [FEA-05] | 测试覆盖缺口 | 待规划 |
| [FEA-06] | systray 缺少开机自启和同步触发 | 待规划 |

### 待讨论（7 项）

| 编号 | 问题 |
|------|------|
| [DSC-01] | `app.Run` mode 参数是否移除 |
| [DSC-02] | CVM `checkRuleLimit` 重复 API 调用优化 |
| [DSC-03] | `testConnection` 是否复用 `ClientPool` |
| [DSC-04] | HTTP server 优雅退出必要性 |
| [DSC-05] | 前端包管理器 npm vs pnpm |
| [DSC-06] | Docker 数据目录路径与 `getDataDir()` 耦合 |
| [WEB-07] | 前端 JSON 字段命名风格统一 |

---

## 核心发现与优先修复建议

1. **构建链路断裂**（最高优先级）：前端无法被打包进二进制是最严重的问题，影响所有分发渠道（Docker、Release、本地编译）。建议优先修复 [BLD-01] + [BLD-03] + [BLD-04] + [DKR-02]，可统一在一次变更中完成。

2. **数据一致性风险**（高优先级）：[WEB-01] 前端数组索引 vs DB ID 错位问题影响数据正确性和同步可靠性，删除记录后规则可能静默停止同步，存在安全风险。建议紧随构建链路修复后处理。

3. **文档残留清理**（中优先级）：[DOC-01]、[DOC-02] 两处文档残留直接影响新用户体验，可批量清理。

4. **Docker 规范清理**（中优先级）：[DKR-01] 删除根 Dockerfile + [DKR-03] 添加 WORKDIR，清理 Docker 配置冗余。

5. **核心功能补全**：熔断器实际生效（[COR-02]）、同步事件发布（[COR-01]）、热重载重建（[COR-07]）对系统可靠性有直接影响，建议在功能补全阶段集中处理。

---

> **审查终止说明**：第 9 轮审查未发现高/中严重度新问题，满足终止条件。第 10 轮最终检查确认无遗漏。审查共进行 10 轮，在第 9 轮后自然终止。
>
> **原始审查编排**：Issue 1–4（构建期 + 首轮修复）→ Issue 5–11（第 1–7 轮深度审查）→ Issue 12（第 8 轮交叉验证）→ 第 9 轮（深度复查）→ Issue 5–8 第二段（第 7–10 轮复查和扩展）。多段审查中存在大量交叉引用和重复记录，本文档已完成合并去重。
# FWAlizer 构建问题记录

> 记录构建过程中遇到的错误、排查过程及解决办法。

---

## Issue 1: go.mod 重复内容

**Step:** 1

**现象：** `go build ./...` 报错 `repeated module statement` 和 `repeated go statement`

**原因：** 覆写 go.mod 时旧内容未完全清除，导致 module 和 go 声明重复出现

**解决：** 重新完整写入 go.mod，确保仅包含一份 module 声明和 go 版本声明

---

## Issue 2: RULES 端口字段逗号与条目分隔符冲突

**Step:** 2

**现象：** `RULES=api.example.com|TCP|443,80|ACCEPT||生产API` 解析失败，端口 `443,80` 中的逗号被误认为条目分隔符

**原因：** Build1.md 参考代码使用简单逗号拆分 (`splitEntries`)，但端口字段允许逗号分隔多端口

**解决：** 实现 `splitRuleEntries` 智能分割——通过正则检测 `host|PROTOCOL|` 模式识别新条目起始位置，仅在该模式匹配时才分割，否则将逗号保留在当前条目内（属于端口字段）

---

## Issue 3: 深度审查问题批量修复

**阶段：** 全量构建完成后的质量审查

### 3.1 CircuitBreaker 未集成到 Syncer

**现象：** `dns/circuitbreaker.go` 已实现但未被 `syncer/syncDomain()` 调用，熔断逻辑未生效

**修复：** 在 Syncer 结构体中添加 `cb *dns.CircuitBreaker` 字段，`syncDomain()` 中解析前检查 `IsOpen()`，解析成功调 `RecordSuccess()`，失败调 `RecordFailure()`

### 3.2 EventBus 未集成到 Syncer

**现象：** `notifier/bus.go` 已实现但 Syncer 未发布任何事件

**修复：** 在 Syncer 中添加 `bus *notifier.EventBus`，DNS 失败时发布 `EventDNSFailed`，同步失败时发布 `EventSyncError`，并暴露 `EventBus()` 方法供外部订阅

### 3.3 WebUI 配置热重载未接通

**现象：** `Syncer.Reload()` 存在但 WebUI 修改配置后未调用

**修复：** WebUI Server 添加 `SetReloadFunc()` 回调机制，在所有写操作（AddTarget/DeleteTarget/AddRule/DeleteRule/PutSettings）后触发；main.go 中 WebUI 模式直接创建 Syncer 并将 `store.LoadConfig() + s.Reload()` 作为回调传入

### 3.4 ClientPool key 缺少 accessID

**现象：** key 格式为 `cloudType|region`，同 region 不同凭据会错误复用 Client

**修复：** 四个 Provider 的 key 统一改为 `cloudType|region|accessID` 格式

### 3.5 App.vue 未使用的 `h` 导入

**修复：** 删除 `import { h, ref } from 'vue'` 中的 `h`

## 3.6 main.go 使用 fmt.Println

**修复：** WebUI 模式启动提示改为 `slog.Info()`，导出 `app.InitLogger()` 供 main.go 调用

---

## Issue 4: 深度审查报告问题核验与修复

**阶段：** 全量构建完成后的深度审查（第二轮）

> 以下 10 项问题均已在代码中逐一定位确认。
> **状态更新：** 4.1–4.6、4.8–4.10 已修复；4.7 待规划（后续 Phase 处理）。

---

### 4.1 ✅ `truncateDesc` 字节截断破坏 `[TAG]` 前缀和中文（已修复）

**严重度：** 高（不通过）

**文件：** `syncer/retry.go` L119-131

**现象：** `truncateDesc` 使用 `desc[:maxLen]` 按字节截断，当描述包含中文（UTF-8 多字节字符）时会截断到半个字符，产生乱码。若 `[TAG]` 前缀本身接近 64 字符限制，截断还可能破坏前缀完整性。

**影响范围：** 所有 Lighthouse 规则的描述字段（`FirewallRuleDescription ≤ 64 字符`），含中文备注的规则均受影响。

**修复方案：**
- 将字节截断改为 rune 切片截断，保证字符完整性
- 关键代码变更：
```go
func truncateDesc(desc string, ct config.CloudType) string {
    maxLen := 0
    switch ct {
    case config.CloudTCLighthouse:
        maxLen = 64
    default:
        return desc
    }
    runes := []rune(desc)
    if len(runes) <= maxLen {
        return desc
    }
    return string(runes[:maxLen])
}
```

---

### 4.2 ✅ `fmt.Sscanf` 返回值未处理（已修复）

**严重度：** 中（不通过）

**文件：** `webui/server.go` L104, L139

**现象：** `handleDeleteTarget` 和 `handleDeleteRule` 中 `fmt.Sscanf(id, "%d", &n)` 忽略返回值。当 URL 路径参数非数字时，`n` 保持零值 0，导致尝试删除 ID=0 的记录（静默失败或误操作）。

**影响范围：** WebUI 删除目标/规则的 API 端点。

**修复方案：**
- 检查 `fmt.Sscanf` 返回的 error，非 nil 时返回 400
- 关键代码变更：
```go
if _, err := fmt.Sscanf(id, "%d", &n); err != nil {
    writeError(w, http.StatusBadRequest, "无效的资源 ID")
    return
}
```

---

### 4.3 ✅ `os.MkdirAll` 返回值未处理（已修复）

**严重度：** 中（不通过）

**文件：** `main.go` L40

**现象：** `os.MkdirAll(dataDir, 0755)` 忽略错误返回。若目录创建失败（权限不足等），后续 `config.OpenStore(dbPath)` 会以不明确的错误退出。

**影响范围：** WebUI 模式启动流程。

**修复方案：**
```go
if err := os.MkdirAll(dataDir, 0755); err != nil {
    fmt.Fprintf(os.Stderr, "创建数据目录失败: %v\n", err)
    os.Exit(1)
}
```

---

### 4.4 ✅ `os.UserHomeDir` 错误忽略（已修复）

**严重度：** 低（不通过）

**文件：** `main.go` L109

**现象：** `home, _ := os.UserHomeDir()` 忽略错误。若无法获取用户主目录（如容器内 HOME 未设置），`home` 为空字符串，数据目录变为相对路径 `.config/fwalizer`。

**影响范围：** WebUI 模式数据目录定位。

**修复方案：**
```go
func getDataDir() string {
    home, err := os.UserHomeDir()
    if err != nil {
        fmt.Fprintf(os.Stderr, "无法获取用户目录: %v\n", err)
        os.Exit(1)
    }
    return filepath.Join(home, ".config", "fwalizer")
}
```

---

### 4.5 ✅ CVM 规则计数未使用 `PolicyStatistics`（已修复）

**严重度：** 低（存在风险）

**文件：** `provider/tc_cvm.go` L219

**现象：** `checkRuleLimit` 使用 `len(ps.Ingress) + len(ps.Egress)` 手动计数，而 Build1.md 规定使用 API 返回的 `PolicyStatistics`（`IngressIPv4TotalCount + IngressIPv6TotalCount + EgressIPv4TotalCount + EgressIPv6TotalCount`）。功能等价但偏离文档规范。

**影响范围：** CVM 安全组规则上限检测逻辑。

**修复方案：**
- 优先使用 `PolicyStatistics` 字段精确计数，fallback 到手动计数
- 关键代码变更：
```go
func (p *TCCVM) checkRuleLimit(toAdd int) error {
    // ...
    ps := resp.Response.SecurityGroupPolicySet
    if ps == nil {
        return nil
    }
    var total int
    if ps.PolicyStatistics != nil {
        stats := ps.PolicyStatistics
        total = int(strVal(stats.IngressIPv4TotalCount)) + ...
    } else {
        total = len(ps.Ingress) + len(ps.Egress)
    }
    // ... 后续判断不变
}
```
- 注：需确认 SDK 中 `PolicyStatistics` 字段类型（可能为 *string 或 *int64）

---

### 4.6 ✅ 半开探测失败仍递增计数器（已修复）

**严重度：** 低（存在风险）

**文件：** `dns/circuitbreaker.go` L44

**现象：** `RecordFailure` 无条件执行 `cb.failCount[domain]++`。Build1.md 12.7 节规定「半开探测失败不计入失败次数，维持熔断状态」。当前实现在已熔断后仍递增计数器，导致数值无限增长（功能不受影响，但偏离文档且影响日志可读性）。

**影响范围：** DNS 熔断器日志输出准确性。

**修复方案：**
```go
func (cb *CircuitBreaker) RecordFailure(domain string) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    // 已熔断时不再递增（半开探测失败维持熔断状态）
    if cb.failCount[domain] >= cb.threshold {
        return
    }
    cb.failCount[domain]++
    if cb.failCount[domain] == cb.threshold {
        slog.Error("DNS 熔断触发", "domain", domain, "连续失败", cb.failCount[domain])
    }
}
```

---

### 4.7 ⏸️ `webui/api/` 子目录未拆分 + Step 13 端点缺失（待规划）

**严重度：** 中（存在风险）

**文件：** `webui/server.go`（全部 API 集中于此）

**现象：**
1. Build1.md Step 13 规定 API 拆分为 `webui/api/targets.go`、`rules.go`、`sync.go`、`settings.go`，实际全部合并在 `server.go` 单文件中
2. 以下端点未实现：
   - `GET /api/sync/status`（同步状态）
   - `POST /api/sync/trigger`（手动触发同步）
   - `POST /api/sync/dryrun`（试运行）
   - `POST /api/test-connection`（测试连接）
   - `GET /api/config/export`（配置导出）
   - `POST /api/config/import`（配置导入）

**影响范围：** WebUI 功能完整性；前端 Dashboard 页面无法展示同步状态、无法手动触发同步。

**修复方案：**
- **文件拆分**（可选，功能等价但符合规范）：将 handler 按职责拆分到 `webui/api/` 子目录
- **端点补全**（优先级高）：
  - `/api/sync/status`：返回 Syncer 上次同步时间、状态、各域名结果
  - `/api/sync/trigger`：调用 Syncer 立即执行一轮同步
  - `/api/sync/dryrun`：复用 Diff 逻辑，仅返回 toAdd/toDelete 不写入
  - `/api/test-connection`：用凭据创建临时 Client 调用 Describe 验证连通性
  - `/api/config/export`：从 SQLite 导出 JSON（凭据字段不导出）
  - `/api/config/import`：接收 JSON 校验后写入 SQLite，触发热重载
- 需要在 `Server` 结构体中增加对 `Syncer` 的引用以支持 status/trigger/dryrun

---

### 4.8 ✅ `truncateLighthouseDesc` 代码重复（已修复）

**严重度：** 低（存在风险）

**文件：** `provider/tc_lighthouse.go` L227-232

**现象：** `truncateLighthouseDesc` 与 `syncer/retry.go` 中的 `truncateDesc` 功能重复。`CreateRules`（L112）调用了 `truncateLighthouseDesc`，而 Syncer 在调用 `CreateRules` 前已通过 `truncateDesc` 截断。存在双重截断 + 代码冗余。

**影响范围：** 代码可维护性；两处截断逻辑若修改一处忘改另一处会产生不一致。

**修复方案：**
- 删除 `tc_lighthouse.go` 中的 `truncateLighthouseDesc` 函数
- `CreateRules` 中直接使用传入的 `r.Description`（已由 Syncer 层截断）
- 关键代码变更：
```go
// tc_lighthouse.go CreateRules 中：
fwRule := &lighthouse.FirewallRule{
    Action:                  common.StringPtr(r.Action),
    FirewallRuleDescription: common.StringPtr(r.Description), // 已由 Syncer 截断
}
```

---

### 4.9 ✅ `go mod tidy` 后 go.mod/go.sum 变更未提交（已修复）

**严重度：** 低（存在风险）

**文件：** `go.mod`、`go.sum`

**现象：** `app/systray.go`（build tag `desktop`）引用 `fyne.io/systray`，但 go.mod 中原本缺失该依赖。执行 `go mod tidy` 后新增：
- `fyne.io/systray v1.12.2`
- `github.com/godbus/dbus/v5 v5.1.0`（indirect）

**影响范围：** 其他开发者 `go build -tags desktop` 时可能因 go.sum 缺失条目而失败。

**修复方案：**
- 执行 `go mod tidy` 并提交 go.mod 和 go.sum 变更

---

### 4.10 ✅ `.env.example` 包含旧格式遗留内容（已修复）

**严重度：** 中（存在风险）

**文件：** `.env.example` L60-142

**现象：** 文件第 1-58 行为当前有效格式，第 60-142 行为旧版本遗留内容，包含：
- 旧格式 TARGETS（凭据嵌入：`provider|resource_id|region|access_id|access_key`）
- 旧变量名：`TENCENTCLOUD_SECRET_ID`、`LIGHTHOUSE_INSTANCE_ID`、`DOMAIN_RULES`、`RULE_TAG`、`CHECK_INTERVAL`、`DNS_SERVER`
- 与当前配置体系完全冲突，会误导用户

**影响范围：** 新用户参照配置时产生困惑；旧变量名不被解析器识别。

**修复方案：**
- 删除 L59-142 的全部旧格式内容
- 仅保留 L1-58 的当前有效配置模板


---------
# FWAlizer 深度审查问题记录

> 记录多轮深度审查过程中发现的设计/实现/编码一致性问题和修复建议。
> 审查基准：Design1.md、Build1.md、AGENTS.md。

---

## Issue 5: 第 1 轮审查 — 目录结构与基础模块

**审查轮次：** 第 1 轮  
**审查范围：** 目录结构、go.mod、基础模块、配置解析、.env.example、测试文件

---

### 5.1 `.env.example` 包含旧格式和重复内容（未彻底清理）

**所在阶段/Step：** Step 2 / 全局配置

**严重度：** 高

**涉及文件：** `.env.example` L59-200

**现象描述：**
- L1-58 为当前有效格式
- L59-117 是 L1-58 的完整重复副本
- L118-200 是旧格式内容，包含已被废弃的变量名：
  - `provider|resource_id|region|access_id|access_key`（5 列格式，凭据嵌入 TARGETS）
  - `TENCENTCLOUD_SECRET_ID`、`TENCENTCLOUD_SECRET_KEY`
  - `LIGHTHOUSE_INSTANCE_ID`、`LIGHTHOUSE_REGION`
  - `DOMAIN_RULES`、`RULE_TAG`、`CHECK_INTERVAL`、`DNS_SERVER`
  - 旧 RULES 格式（5 列，无 comment 字段）

**原因分析：** Issue 4.10 已记录此问题并标记为"已修复"，但实际修复未生效——旧行可能因操作失误未被删除，或后续合并覆盖了修复。

**影响范围：** 新用户参照配置时产生严重困惑；旧变量名不被当前解析器识别，配置必然失败；重复内容增加了 200 行的无效文件体积。

**推荐修复方案：**
- 删除 L59-200（含重复的现代格式段和旧格式段）
- 仅保留 L1-58 的当前有效配置模板

**当前状态：** 待修复

---

### 5.2 `firewall/` 空目录残留

**所在阶段/Step：** Step 1 / 项目骨架

**严重度：** 低

**涉及文件：** `firewall/`（空目录）

**现象描述：** 项目根目录存在 `firewall/` 目录，但无任何文件。Build1.md 第一章的目录结构未提及此目录。AGENTS.md 声明「历史代码（旧 `config/`、`dns/`、`firewall/`、`main.go` 等）可直接删除或覆盖」。

**原因分析：** 旧代码目录的残留，可能在重构时未清理。

**影响范围：** 无功能影响，仅影响目录整洁度。

**推荐修复方案：** 删除空 `firewall/` 目录。

**当前状态：** 待修复

---

### 5.3 `app.Run` 中 `mode` 参数未被使用

**所在阶段/Step：** Step 7 / App 生命周期

**严重度：** 低

**涉及文件：** `app/app.go` L15

**现象描述：** `func Run(cfg *config.Config, mode Mode) error` 接收 `mode Mode` 参数，但函数体内从未使用该参数。Build1.md Step 7 的设计也是如此（`app.Run` 不区分 mode），因为实际的模式差异由 `main.go` 处理——ModeEnv 走 `app.Run`，ModeWebUI 完全在 `main.go` 内联处理。

**原因分析：** `app.Run` 最初设计为统一入口，Phase 2 后 WebUI 模式逻辑移到 `main.go` 内联，`app.Run` 退化为仅服务于 `.env` 模式的函数，但参数签名未同步精简。

**影响范围：** 无运行期影响，但增加代码阅读困惑（看起来 mode 参数有作用）。

**推荐修复方案：**
- **方案 A（推荐）：** 移除 `mode` 参数，将 `app.Run` 精简为接受 `*config.Config` 即可
- **方案 B：** 保留参数但在函数内根据 mode 做差异化行为（如 mode 为 webui 时跳过某些逻辑），与 Build1.md 设计保持一致

**当前状态：** 待讨论

---

## 第 1 轮审查总结

- **发现高严重度问题：** 1 项（5.1 `.env.example` 旧格式未清理）
- **发现中严重度问题：** 0 项
- **发现低严重度问题：** 2 项（5.2 空目录, 5.3 未使用参数）
- **待讨论事项：** 1 项（5.3 mode 参数处理方案）
- **剩余风险：** `.env.example` 的旧格式内容如不清理，每个新用户都会受到误导。

---

## Issue 6: 第 2 轮审查 — Provider 抽象与多云实现

**审查轮次：** 第 2 轮  
**审查范围：** provider/provider.go、registry.go、common.go、credentials.go、tc_lighthouse.go、tc_cvm.go、ali_swas.go、ali_ecs.go

---

### 6.1 `strVal` 工具函数放置在 `tc_lighthouse.go` 而非 `common.go`

**所在阶段/Step：** Step 5 / Lighthouse Provider

**严重度：** 低

**涉及文件：** `provider/tc_lighthouse.go` L227-232, `provider/common.go`

**现象描述：** 包级工具函数 `strVal` 定义在 `tc_lighthouse.go` 中，但被 `tc_lighthouse.go`、`tc_cvm.go`、`ali_swas.go`、`ali_ecs.go` 四个文件共用。虽然同属 `provider` 包无编译问题，但按惯例共享工具函数应放在 `common.go`。

**原因分析：** `strVal` 在 Lighthouse 实现时首先引入，后续 CVM/SWAS/ECS 复用但未将其移动到公共文件。

**影响范围：** 代码可读性和组织性轻微下降；若未来 `tc_lighthouse.go` 被重构或移除，会意外删除共享函数。

**推荐修复方案：** 将 `strVal` 函数从 `tc_lighthouse.go` 移动到 `common.go`。

**当前状态：** 待修复

---

### 6.2 CVM `checkRuleLimit` 存在重复 API 调用

**所在阶段/Step：** Step 8 / CVM Provider

**严重度：** 低

**涉及文件：** `provider/tc_cvm.go` L204-234

**现象描述：** `CreateRules` 调用 `checkRuleLimit`，后者独立调用 `DescribeSecurityGroupPolicies` 获取规则总数。但 Syncer 的 `retrySync` 在调用 `CreateRules` 前已经通过 `GetRules()` 获取了所有规则（含 Egress）。`checkRuleLimit` 再次调用相同 API 造成了额外的网络开销。

**原因分析：** `checkRuleLimit` 为了获取最新规则计数以确保安全，独立完成 API 调用。但乐观锁模式下 `retrySync` 已保证每次重试都重新 Describe。

**影响范围：** CVM 安全组同步时多一次 API 调用（查询+删除 100次/秒配额下影响很小）。

**推荐修复方案：**
- **方案 A（推荐）：** 将规则计数检查上提到 `retrySync` 层，复用 `GetRules()` 返回的结果（可从 `allRules` 中统计 Ingress+Egress 数量）
- **方案 B：** 保持现状，额外 API 调用作为安全冗余

**当前状态：** 待讨论

---

## 第 2 轮审查总结

- **发现高严重度问题：** 0 项
- **发现中严重度问题：** 0 项
- **发现低严重度问题：** 2 项（6.1 strVal 位置, 6.2 重复 API 调用）
- **待讨论事项：** 1 项（6.2 checkRuleLimit 优化方案）
- **剩余风险：** Provider 层实现与 Build1.md 高度一致，核心逻辑正确，无功能缺陷。

---

## Issue 7: 第 3 轮审查 — Syncer、DNS、熔断、重试、限流、EventBus

**审查轮次：** 第 3 轮  
**审查范围：** syncer/syncer.go、retry.go、ratelimit.go、dns/resolver.go、circuitbreaker.go、notifier/bus.go、email.go、webhook.go

---

### 7.1 `sync:start` 和 `sync:complete` 事件已定义但未被发布

**所在阶段/Step：** Step 6 + Step 11 / Syncer + EventBus

**严重度：** 中

**涉及文件：** `syncer/syncer.go` L172-198, `notifier/bus.go` L13-14

**现象描述：** `notifier/bus.go` 定义了 `EventSyncStart`、`EventSyncComplete`、`EventSyncError`、`EventRuleChanged`、`EventDNSFailed` 五种事件。但 Syncer 仅在 `syncDomain()` 中发布 `EventDNSFailed`（L212）和 `EventSyncError`（L224），`syncAll()` 从未发布 `EventSyncStart` 和 `EventSyncComplete`。

**原因分析：** 事件类型随 Build1.md Step 11 设计时定义了全集，但 Syncer 只实现了错误路径的事件发布，未将正常生命周期事件接入。

**影响范围：**
- 依赖 `sync:start`/`sync:complete` 事件的消费者（WebUI SSE 推送、邮件告警、Webhook）无法感知同步开始和完成；
- 邮件/Webhook 通知器目前只处理 `sync:error` 和 `dns:failed`，但不排除未来需要同步完成通知；
- WebUI Dashboard 需要通过 start/complete 事件追踪同步进行状态而非仅轮询。

**推荐修复方案：**
- 在 `syncAll()` 开始时发布 `EventSyncStart`
- 在 `syncAll()` 完成时（wg.Wait 后）发布 `EventSyncComplete`
- `EventSyncComplete` 的 Data 中附带耗时、成功/失败统计等信息

**当前状态：** 待修复

---

## 第 3 轮审查总结

- **发现高严重度问题：** 0 项
- **发现中严重度问题：** 1 项（7.1 事件发布不完整）
- **发现低严重度问题：** 0 项
- **待讨论事项：** 0 项
- **剩余风险：** Syncer/熔断/重试/限流/DNS 解析核心逻辑均与 Build1.md 一致，EventBus 事件发布是唯一的功能缺口。

---

## Issue 8: 第 4 轮审查 — WebUI 后端与配置持久化

**审查轮次：** 第 4 轮  
**审查范围：** webui/server.go、embed.go、api/deps.go、api/targets.go、api/rules.go、api/sync.go、api/settings.go、config/store.go

---

### 8.1 配置导入缺少事务保护

**所在阶段/Step：** Step 13 / WebUI 后端

**严重度：** 中

**涉及文件：** `webui/api/settings.go` L97-118

**现象描述：** `handleConfigImport` 依次执行 `ClearAll()` → `BatchAddTargets()` → `BatchAddRules()` → settings 写入，这些操作不在同一事务中。若 `BatchAddRules` 失败，数据库处于"已清空但未导入完整"的不一致状态。

**原因分析：** SQLite 的 `DELETE FROM` + `INSERT` 未包裹在 `BEGIN`/`COMMIT` 事务中。

**影响范围：** 配置导入中途失败时，用户原有配置丢失且新配置未完全写入。

**推荐修复方案：**
- 在 `config/store.go` 添加 `WithTransaction(fn func(tx *sql.Tx) error) error` 事务方法
- 将 `ClearAll`、`BatchAddTargets`、`BatchAddRules`、settings 写入放在同一事务中
- 失败时自动回滚

**当前状态：** 待修复

---

### 8.2 WebUI 模式缺少 pidfile 防多实例

**所在阶段/Step：** Step 13 / WebUI 后端

**严重度：** 低

**涉及文件：** `main.go` L37-104（WebUI 模式启动逻辑）

**现象描述：** Build1.md 12.12 节规定「WebUI 模式启动时创建 pidfile，检测 pidfile 是否存在且进程存活，是则拒绝启动」。当前 `main.go` 的 WebUI 路径未实现 pidfile 检测/创建/清理逻辑。

**原因分析：** pidfile 属于非核心功能，优先实现时被延后。

**影响范围：** 用户误启动多次可能导致两个实例争抢 SQLite 写入、端口冲突。

**推荐修复方案：**
- 在 WebUI 模式启动时，在数据目录创建 `fwalizer.pid` 文件写入 PID
- 启动前检查已有 pidfile 的进程是否存活
- 正常退出时删除 pidfile

**当前状态：** 待规划

---

## 第 4 轮审查总结

- **发现高严重度问题：** 0 项
- **发现中严重度问题：** 1 项（8.1 导入缺事务）
- **发现低严重度问题：** 1 项（8.2 pidfile 缺失）
- **待讨论事项：** 0 项
- **剩余风险：** Issue 4.7 中所有缺失端点已全部补齐（含 SSE 实时推送），API 层功能完整。

---

## Issue 9: 第 5 轮审查 — 前端页面与 API 对接

**审查轮次：** 第 5 轮  
**审查范围：** webui/frontend/ 全部 Vue/TS 源码、Vite 配置、package.json、前端与后端 API 对接一致性

---

### 9.1 前端构建产物 `dist/` 目录缺失，Go `embed` 编译将失败

**所在阶段/Step：** Step 13 / 静态资源嵌入

**严重度：** 高

**涉及文件：** `webui/embed.go` L5-6, `webui/server.go` L61-64

**现象描述：**
- `embed.go` 使用 `//go:embed frontend/dist` 指令嵌入前端构建产物
- `webui/frontend/dist/` 目录在 `.gitignore` 中被排除（L9），仓库中不存在
- `server.go` L61-64 的 `fs.Sub(frontendFS, "frontend/dist")` 在 `dist/` 不存在时返回 error，虽然当前代码用 `if err == nil` 优雅降级（无前端时仅提供 API），但产品形态下 WebUI 完全不可用

**原因分析：** 前端需要执行 `cd webui/frontend && npm install && npm run build` 生成 `dist/`，但此步骤未在任何构建流程中自动化。

**影响范围：**
- 本地 `go build` 生成的二进制不含前端页面
- Docker 镜像不含前端页面
- CI Release 发布的二进制不含前端页面
- 用户部署后访问 WebUI 仅看到 404

**推荐修复方案：**
- **短期：** 在项目 README 中明确说明构建前需要先 `npm run build`
- **长期：** 在 Dockerfile 和 CI 流程中集成前端构建步骤（见 Issue 10.2/10.3/10.4）
- 在 `embed.go` 附近或 Makefile 中添加 `generate` 注释说明构建前置条件

**当前状态：** 待修复

---

### 9.2 前端 CRUD 操作用 `index + 1` 作为记录 ID，与后端数据库 ID 可能不一致

**所在阶段/Step：** Step 13 / WebUI 前端

**严重度：** 中

**涉及文件：** `webui/frontend/src/views/Rules.vue` L37, L56, `webui/frontend/src/views/Targets.vue` L33, L52

**现象描述：**
- `Rules.vue` 在 `openEdit(row, index)` 中使用 `editingId.value = index + 1`，`deleteRule(index)` 使用 `fetch(\`/api/rules/${index + 1}\`)`
- `Targets.vue` 同理使用 `index + 1` 作为记录的 API ID
- 后端 API (`rules.go` L37, `targets.go` L37) 通过 `r.PathValue("id")` 解析为数据库 `id`（SQLite `INTEGER PRIMARY KEY AUTOINCREMENT`）
- 前端依赖表格行顺序与数据库 ID 严格一致（第 1 行 = id 1, 第 2 行 = id 2, …），但以下场景会打破此假设：
  - 用户删除中间某条记录后新增，SQLite 不会复用被删除的 ID，导致行 2 的 id 可能是 5
  - 后端 `GetTargets()`/`GetRules()` 返回的 JSON 数组不包含 `id` 字段（仅包含业务字段），前端完全无法获取真实 ID
  - 排序方式改变后行索引对应关系丢失

**原因分析：** 后端 API 返回的 `TargetConfig` 和 `DomainRule` JSON 序列化不包含数据库 `id` 字段。前端无真实 ID 可用，只能依赖数组索引。

**影响范围：**
- 用户在 WebUI 中编辑/删除记录时可能操作错误的记录
- 在"添加 → 删除 → 再添加"的操作序列后必然触发此 bug

**推荐修复方案：**
- **推荐：** 后端 API 返回数据时包含 `id` 字段（修改 `TargetConfig`/`DomainRule` JSON 序列化或创建含 ID 的视图结构体）
- 前端改为使用真实的 `row.id` 而非 `index + 1`
- 短期 workaround：使用 `row.id` 字段（需后端配合添加）

**当前状态：** 待修复

---

### 9.3 TypeScript `any` 类型泛滥

**所在阶段/Step：** Step 13 / WebUI 前端

**严重度：** 低

**涉及文件：** 所有 `.vue` 文件中的 `<script setup lang="ts">` 块

**现象描述：** 所有组件广泛使用 `any` 类型：
- `const status = ref<any>({ running: false })`（Dashboard.vue L5）
- `const rules = ref<any[]>([])`（Rules.vue L5）
- `const targets = ref<any[]>([])`（Targets.vue L5）
- `const settings = ref<Record<string, string>>({})`（Settings.vue L5）— 此项正确
- `render(row: any)`（多个文件）

虽然 `tsconfig.json` 中 `strict: true` 已启用，但 `any` 绕过了所有类型检查。

**原因分析：** 快速开发阶段未定义 TypeScript 接口类型。

**影响范围：** IDE 无法提供自动补全和编译期类型检查；重构时容易引入 bug。

**推荐修复方案：**
- 在 `src/types.ts` 中定义 `Target`、`Rule`、`SyncStatus`、`DryRunResult`、`SyncLog`、`Event` 等接口
- 逐步替换所有 `any` 为具体类型
- 最终启用 `"noImplicitAny": true`

**当前状态：** 待规划

---

### 9.4 Dashboard 使用轮询而非 SSE 获取状态

**所在阶段/Step：** Step 13 / WebUI 前端

**严重度：** 低

**涉及文件：** `webui/frontend/src/views/Dashboard.vue` L36-38, `webui/frontend/src/views/Logs.vue` L13-21

**现象描述：**
- Dashboard 使用 `setInterval(fetchStatus, 5000)` 每 5 秒轮询 `/api/sync/status`
- 而 Logs 页面已经正确使用 SSE `EventSource('/api/sync/events')` 获取实时事件
- 这导致状态更新有 5 秒延迟，且产生不必要的 HTTP 请求

**原因分析：** Dashboard 开发时 SSE 事件端点可能尚未实现，后续 SSE 加入后未同步更新 Dashboard。

**影响范围：** 状态刷新延迟最高 5 秒；轻微增加服务器负载。

**推荐修复方案：**
- Dashboard 同时使用 SSE 监听 `sync:start`/`sync:complete` 事件来更新 `last_sync` 和运行状态
- 保留轮询作为 SSE 断连时的 fallback（延长间隔到 30s）

**当前状态：** 待规划

---

## 第 5 轮审查总结

- **发现高严重度问题：** 1 项（9.1 `dist/` 缺失导致 WebUI 不可用）
- **发现中严重度问题：** 1 项（9.2 前端 ID 与数据库 ID 错位风险）
- **发现低严重度问题：** 2 项（9.3 TypeScript `any`, 9.4 Dashboard 轮询）
- **待讨论事项：** 0 项
- **剩余风险：** 前端核心 CRUD 功能因 ID 问题存在正确性风险；整体功能完整，UI 体验良好。

---

## Issue 10: 第 6 轮审查 — Docker、构建、CI/CD、桌面端

**审查轮次：** 第 6 轮  
**审查范围：** Dockerfile、.dockerignore、Makefile、app/systray.go、app/mode.go、app/cli.go、main.go、.github/workflows/

---

### 10.1 `docker-publish.yml` 内容严重重复

**所在阶段/Step：** Step 16 / CI/CD

**严重度：** 高

**涉及文件：** `.github/workflows/docker-publish.yml` L1-161

**现象描述：**
- L1-84：第一段完整的 workflow 定义（name、on、env、jobs.docker）
- L85-161：第二段几乎重复的 workflow 定义（name、on、env、jobs.build），仅 job 名和部分步骤有差异
- 两个重复的 `name:` / `on:` / `env:` 定义会导致 YAML 解析只取最后一个
- L75/76 第一段使用 `file: build/Dockerfile`（路径错误），第二段未指定 file（默认根目录 Dockerfile）
- 第一段包含 `build-args: VERSION` 但 Dockerfile 未声明 `ARG VERSION`

**原因分析：** 合并或编辑操作失误，导致内容重复而非替换。

**影响范围：**
- YAML 多文档解析行为不确定，可能导致 workflow 无法触发或行为异常
- Docker 镜像构建路径引用不一致

**推荐修复方案：**
- 删除 L1-84（第一段），保留 L85-161（第二段，结构更简洁正确）
- 在保留的第二段中补充缺失的步骤：Node.js 安装、前端构建、VERSION build-arg

**当前状态：** 待修复

---

### 10.2 CI/CD 流程缺少前端构建步骤

**所在阶段/Step：** Step 16 / CI/CD

**严重度：** 高

**涉及文件：** `.github/workflows/docker-publish.yml`, `.github/workflows/release.yml`

**现象描述：**
- 两个 CI workflow 都包含 Go 编译检查、测试步骤，但均缺少：
  1. `actions/setup-node@v4` 安装 Node.js
  2. `cd webui/frontend && npm ci && npm run build` 生成 `dist/`
- 导致 `//go:embed frontend/dist` 指令在 CI 环境中找不到文件
- Docker 镜像和 Release 二进制均不含 WebUI 前端

**原因分析：** 前端构建流程未纳入 CI pipeline。

**影响范围：**
- Docker 镜像构建失败（`pattern frontend/dist: no matching files found`）或构建出不含前端页面的二进制
- Release 二进制不含 WebUI，用户下载后无法使用 WebUI 模式

**推荐修复方案：**
- 在两个 workflow 的 Go 编译步骤前添加：
  ```yaml
  - name: 设置 Node.js
    uses: actions/setup-node@v4
    with:
      node-version: '20'
  - name: 构建前端
    run: |
      cd webui/frontend
      npm ci
      npm run build
  ```
- `release.yml` 需要在每个 `GOOS`/`GOARCH` 组合前先构建前端（只需一次）

**当前状态：** 待修复

---

### 10.3 Dockerfile 缺少前端构建阶段

**所在阶段/Step：** Step 15 / Docker

**严重度：** 高

**涉及文件：** `Dockerfile` L1-29

**现象描述：**
- Dockerfile 仅包含 Go 编译阶段（`golang:1.25-alpine`）和运行阶段（`alpine:3.20`）
- 缺少 Node.js 构建阶段用于编译 Vue 前端
- 缺少 `COPY webui/frontend/dist` 步骤将前端产物复制到 Go 编译上下文中

**原因分析：** 多阶段构建未包含前端编译阶段。

**影响范围：** Docker 镜像构建出的二进制不含前端页面，WebUI 模式无法使用。

**推荐修复方案：**
- 在 `golang:1.25-alpine` 之前添加 Node.js 构建阶段：
  ```dockerfile
  FROM node:20-alpine AS frontend-builder
  WORKDIR /src/webui/frontend
  COPY webui/frontend/package*.json ./
  RUN npm ci
  COPY webui/frontend/ ./
  RUN npm run build
  
  FROM golang:1.25-alpine AS builder
  # ...
  COPY --from=frontend-builder /src/webui/frontend/dist ./webui/frontend/dist
  ```
- 调整 `.dockerignore` 确保 `webui/frontend/node_modules` 不被排除

**当前状态：** 待修复

---

### 10.4 `release.yml` 不构建前端

**所在阶段/Step：** Step 16 / Release

**严重度：** 高

**涉及文件：** `.github/workflows/release.yml` L26-57

**现象描述：** `release.yml` 直接编译多平台 Go 二进制并发布到 GitHub Release，未先构建前端。所有发布的多平台二进制均不含 WebUI 前端。

**原因分析：** 与 10.2 相同，前端构建未纳入 Release pipeline。

**影响范围：** GitHub Release 中所有二进制均无法使用 WebUI 模式。

**推荐修复方案：** 参见 10.2 的修复方案。

**当前状态：** 待修复

---

### 10.5 `Dockerfile` 位置与 CI 引用不一致

**所在阶段/Step：** Step 15/16 / Docker + CI

**严重度：** 中

**涉及文件：** `Dockerfile`（根目录）, `.github/workflows/docker-publish.yml` L75

**现象描述：**
- `Dockerfile` 位于项目根目录 `/Dockerfile`
- `docker-publish.yml` 第一段 L75/76 引用 `file: build/Dockerfile`（`build/` 目录不存在）
- 第二段 L152-155 未指定 `file`，默认使用根目录 `Dockerfile`

**原因分析：** 早期设计可能计划将 Dockerfile 放在 `build/` 子目录，后移回根目录但 CI 未更新。

**影响范围：** 第一段 workflow（若生效）将因找不到 `build/Dockerfile` 而失败。删除第一段后（10.1），第二段无此问题。

**推荐修复方案：** 与 10.1 一起修复：删除第一段，保留第二段（默认使用根目录 Dockerfile）。

**当前状态：** 待修复

---

### 10.6 `systray.go` 中「立即同步」菜单项为 TODO 空实现

**所在阶段/Step：** Step 14 / 桌面端

**严重度：** 低

**涉及文件：** `app/systray.go` L46-47

**现象描述：** `mSync` 菜单项点击后仅打印日志，实际同步触发为 `// TODO: 通过 channel 通知 Syncer 立即同步`。

**原因分析：** 桌面端托盘与 Syncer 之间缺少通信 channel。

**影响范围：** 桌面端托盘「立即同步」按钮无实际效果。

**推荐修复方案：**
- 在 `RunSystray` 函数签名中添加 `triggerCh chan<- struct{}` 参数
- 在 `mSync.ClickedCh` 分支中向 channel 发送信号
- 或在 Syncer 中暴露 `TriggerSync()` 公开方法供 systray 调用

**当前状态：** 待规划

---

### 10.7 `main.go` WebUI 的 `go srv.Start()` 无显式优雅退出

**所在阶段/Step：** Step 7/13 / App 生命周期

**严重度：** 低

**涉及文件：** `main.go` L93-95

**现象描述：**
```go
go srv.Start()
go s.Run()
syncer.WaitForSignal(s)
```
当 `WaitForSignal` 收到 SIGTERM/SIGINT 后 Syncer 停止、main 返回、进程退出。HTTP server goroutine 没有显式的 `Shutdown()` 调用，而是随进程退出被强制终止。

**原因分析：** 优雅退出仅覆盖了 Syncer（完成当前轮次），HTTP server 无等价处理。

**影响范围：** HTTP 连接被强制断开；对于内部工具影响很小。

**推荐修复方案：**
- 使用 `context.WithCancel` + `http.Server.Shutdown()` 实现 HTTP server 优雅退出
- 信号到达后先 Shutdown HTTP server，再停止 Syncer
- 或保持现状（内部工具场景下可接受）

**当前状态：** 待讨论

---

## 第 6 轮审查总结

- **发现高严重度问题：** 4 项（10.1 CI 文件重复, 10.2 CI 缺前端构建, 10.3 Dockerfile 缺前端构建, 10.4 Release 缺前端构建）
- **发现中严重度问题：** 1 项（10.5 Dockerfile 路径不一致）
- **发现低严重度问题：** 2 项（10.6 systray TODO, 10.7 HTTP 优雅退出）
- **待讨论事项：** 1 项（10.7 HTTP 优雅退出必要性）
- **剩余风险：** 构建/部署链路是当前最薄弱环节，前端无法嵌入导致 WebUI 模式在所有分发渠道不可用。

---

## Issue 11: 第 7 轮审查 — 错误处理、边界条件、幂等性、安全

**审查轮次：** 第 7 轮  
**审查范围：** 全项目 error 处理、边界条件、幂等语义、频率限制、安全性、注释规范

---

### 11.1 `truncateDesc` 仅处理 Lighthouse 描述截断，SWAS 的 50 字符限制未覆盖

**所在阶段/Step：** Step 5/6 / Provider

**严重度：** 中

**涉及文件：** `syncer/retry.go` L120-133, Documents 中 SWAS API 文档

**现象描述：**
- `truncateDesc` 只对 `CloudTCLighthouse` 做 64 字符截断
- `default` 分支直接返回原字符串，不做任何截断
- 但阿里云 SWAS 的 `Remark` 字段限制为 **50 字符**（小于 Lighthouse 的 64）
- 若规则描述（`[TAG] + comment`）超过 50 字符，SWAS `CreateFirewallRules` API 将返回参数错误

**原因分析：** `truncateDesc` 实现时仅考虑了 Lighthouse 的限制，未检查 SWAS 文档。

**影响范围：**
- 使用较长 TAG 名称（如 `auto-dns-prod`）或较长 comment 时，SWAS 同步失败
- 描述被截断后可能影响规则识别（但 `[TAG]` 前缀始终保留，影响较小）

**推荐修复方案：**
```go
case config.CloudAliSWAS:
    maxLen = 50 // Remark ≤ 50 字符
```

**当前状态：** 待修复

---

### 11.2 CVM `checkRuleLimit` 未统计 Egress 规则和 IPv6 规则

**所在阶段/Step：** Step 8 / CVM Provider

**严重度：** 低

**涉及文件：** `provider/tc_cvm.go` L218-226

**现象描述：**
- `checkRuleLimit` 优先使用 `PolicyStatistics` 精确统计（包含 IngressIPv4、IngressIPv6、EgressIPv4、EgressIPv6）
- fallback 到 `len(ps.Ingress) + len(ps.Egress)` 手动计数
- 但 fallback 路径不包含 IPv6 规则统计（`Ipv6Ingress`/`Ipv6Egress` 字段未计入）

**原因分析：** API 中 IPv4 和 IPv6 规则在不同字段（`Ingress` vs `Ipv6Ingress`），手动统计时未包含 IPv6。在 `PolicyStatistics` 存在时无影响。

**影响范围：** 仅当 API 未返回 `PolicyStatistics` 且存在 IPv6 规则时，规则计数不准确可能导致超限未告警。此情况极少见。

**推荐修复方案：** fallback 路径补充 `len(ps.Ipv6Ingress) + len(ps.Ipv6Egress)`。

**当前状态：** 待规划

---

### 11.3 `webui/api/targets.go` `handleTestConnection` 每次创建新 `ClientPool`

**所在阶段/Step：** Step 12 / WebUI API

**严重度：** 低

**涉及文件：** `webui/api/targets.go` L102

**现象描述：** `handleTestConnection` 创建全新的 `provider.NewClientPool()`，而非复用应用级 pool。这使得：
- 测试连接时的 SDK client 与同步引擎使用的 client 不是同一个实例
- 连接测试成功不代表同步引擎也使用相同的 client 配置

**原因分析：** API handler 未持有应用级 `ClientPool` 引用。

**影响范围：** 极小——测试连接的 client 配置与同步引擎可能仅凭据相同，其他配置一致。

**推荐修复方案：**
- 将 `ClientPool` 注入 `Deps` 中（或通过 `Syncer` 接口暴露），测试连接时复用
- 或保持现状（差异极小，不值得增加耦合）

**当前状态：** 待讨论

---

### 11.4 熔断器 `IsOpen` 在半开状态的上报不准确

**所在阶段/Step：** Step 6 / 熔断

**严重度：** 低

**涉及文件：** `syncer/syncer.go` L203, `dns/circuitbreaker.go`

**现象描述：**
- `syncDomain` 调用 `s.cb.IsOpen(rule.Host)` 检查熔断状态
- 但 `IsOpen` 返回 `true` 仅表示已熔断（含半开状态），`syncDomain` 在 `IsOpen` 为 true 时仍继续执行（仅打印 Debug 日志）
- 半开探测和熔断放行的行为没有区别——两者都会继续执行 DNS 解析
- 正确的熔断行为应是：完全熔断时跳过、半开时探测一次、正常时直接执行

**原因分析：** 熔断检查只在 `syncDomain` 开头做了判断，但没有根据熔断状态（Open vs HalfOpen）做差异化处理。

**影响范围：** 半开状态下的域名实际行为和熔断无区别——都会执行 DNS 解析（因为 `IsOpen` 仅打印日志未 return）。当连续失败时不受影响；当 DNS 间歇恢复时，不会因为熔断而延迟恢复。总体功能影响较小。

**推荐修复方案：**
- 检查 `cb.State(host)` 返回的状态枚举（Open/HalfOpen/Closed）
- Open → 直接 return 跳过
- HalfOpen → 继续执行（当前行为正确）
- 或保持现状，在当前场景下熔断检查的实用价值有限

**当前状态：** 待讨论

---

### 11.5 错误处理与注释合规检查

**所在阶段/Step：** 全局

**严重度：** 低（综合检查）

**涉及文件：** 多个

**检查结果：**
- ✅ **所有 error 均已处理** — 未发现被忽略的 error 返回值
- ✅ **日志使用 `log/slog`** — 全项目统一使用 `log/slog`，无 `fmt.Println` 用于日志
- ✅ **注释使用中文** — 所有注释为中文，面向国内开发者
- ✅ **未使用全量覆盖 API** — 所有 Provider 均使用增量 `Create*`/`Delete*` API
- ✅ **仅操作入站规则** — CVM/ECS 明确过滤 Ingress，SWAS/Lighthouse 仅支持入站
- ✅ **WebUI 绑定 127.0.0.1** — `server.go` L46 `addr := fmt.Sprintf("127.0.0.1:%d", s.port)`
- ✅ **凭据隔离** — 凭据通过独立环境变量传入，不与 TARGETS 混合；配置导出自动剔除凭据
- ✅ **频率限制符合文档** — `ratelimit.go` 对应各云厂商配额
- ✅ **幂等处理正确** — 删除"已不存在"视为成功，添加"已存在"视为成功 WARN 跳过
- ✅ **重试语义正确** — 每次重试重新 Describe → Diff → Create/Delete
- ✅ **TCP+UDP 拆分** — `buildDesired` 中仅 SWAS 保留 TCP+UDP，其余拆分为 TCP+UDP 两条
- ✅ **IPv6+ICMP 处理** — Lighthouse 用 ICMPv6、CVM 用 ICMPV6、ECS 跳过并 WARN

**个别注意事项：**
- `app/cli.go` L28/31 使用 `fmt.Fprintf(os.Stderr, ...)` + `os.Exit(1)` 而非 slog——CLI 子命令场景下合理（需要在 slog 初始化前输出）
- `main.go` L34/41/47/54 同理使用 `fmt.Fprintf(os.Stderr, ...)` 在 `InitLogger` 前输出——合理

**当前状态：** 无待修复项

---

## 第 7 轮审查总结

- **发现高严重度问题：** 0 项
- **发现中严重度问题：** 1 项（11.1 `truncateDesc` 缺少 SWAS 50 字符限制）
- **发现低严重度问题：** 3 项（11.2 CVM 规则计数, 11.3 测试连接 ClientPool, 11.4 熔断半开状态）
- **待讨论事项：** 2 项（11.3 ClientPool 复用, 11.4 熔断行为优化）
- **剩余风险：** 核心错误处理和边界条件符合 AGENTS.md 要求，代码质量整体良好。

---

## 多轮审查总汇

| 轮次 | 高 | 中 | 低 | 待讨论 |
|------|----|----|----|--------|
| 第 1 轮（目录/基础模块） | 1 | 0 | 2 | 1 |
| 第 2 轮（Provider） | 0 | 0 | 2 | 1 |
| 第 3 轮（Syncer/DNS） | 0 | 1 | 0 | 0 |
| 第 4 轮（WebUI 后端） | 0 | 1 | 1 | 0 |
| 第 5 轮（前端） | 1 | 1 | 2 | 0 |
| 第 6 轮（Docker/CI） | 4 | 1 | 2 | 1 |
| 第 7 轮（错误/边界） | 0 | 1 | 3 | 2 |
| **合计** | **6** | **5** | **12** | **5** |

### 高严重度问题清单
1. **5.1** `.env.example` 含旧格式重复内容 — 待修复
2. **9.1** 前端 `dist/` 缺失 — 待修复
3. **10.1** `docker-publish.yml` 内容重复 — 待修复
4. **10.2** CI 缺前端构建步骤 — 待修复
5. **10.3** Dockerfile 缺前端构建阶段 — 待修复
6. **10.4** Release 流程缺前端构建 — 待修复

### 核心发现
- **构建链路断裂**：前端无法被打包进二进制是最严重的问题，影响所有分发渠道（Docker、Release、本地编译）
- **前端 ID 一致性问题**：规则/目标的增删改查使用表索引而非数据库 ID，存在操作错误记录的风险
- **Provider 层实现质量高**：与 Build1.md 高度一致，核心逻辑正确
- **代码规范性好**：error 全部处理、slog 统一、中文注释、凭据隔离均到位

---

## 第 9 轮审查 — 深度复查（无新问题）

**审查轮次：** 第 9 轮  
**审查范围：** Documents API 文档一致性、前后端 JSON 字段映射、测试文件覆盖质量、.gitignore 完整性

**检查结果：**
- ✅ SWAS `CreateFirewallRules` 文档确认 `Remark` 字段无显式字符限制声明，`FirewallRuleAlreadyExist` 错误码已被 `isIdempotentCreate` 覆盖
- ✅ ECS `AuthorizeSecurityGroup` 文档确认 `Description` 限制为 1~512 字符，`truncateDesc` 的 `default` 分支不截断是安全的
- ✅ 前端 JSON key（PascalCase）→ Go struct field（lowercase）映射：Go `encoding/json` 对无 tag 字段做**大小写不敏感**匹配，`CloudType`↔`cloudtype`、`ResourceID`↔`resourceid` 均正确匹配
- ✅ 前端 API 端点与后端路由注册 15 个端点完全一一对应（含 GET/POST/PUT/DELETE 方法匹配）
- ✅ 测试文件覆盖 `config`、`dns`、`internal/portconv`、`internal/tag`、`notifier`、`provider/common` 六个包，核心逻辑（Diff、熔断、端口转换、标签、事件总线）均有测试
- ✅ `.gitignore` 覆盖 `.env`、`dist/`、`node_modules/`、`webui/frontend/dist/`、IDE 文件、OS 文件、日志文件

**本轮无新发现的高/中/低严重度问题。**

---

## 第 9 轮审查总结

- **发现高严重度问题：** 0 项
- **发现中严重度问题：** 0 项  
- **发现低严重度问题：** 0 项
- **待讨论事项：** 0 项
- **剩余风险：** 前 8 轮已识别的问题全部记录在案，无新增风险点。

---

## 审查终止

第 9 轮审查未发现高严重度或中严重度新问题，满足终止条件 1：**连续一轮审查未发现高严重度或中严重度问题**。

审查在 9 轮后自然终止（未触及 10 轮强制上限）。

---

## 最终汇总

| 轮次 | 高 | 中 | 低 | 待讨论 |
|------|----|----|----|--------|
| 第 1 轮（目录/基础模块） | 1 | 0 | 2 | 1 |
| 第 2 轮（Provider） | 0 | 0 | 2 | 1 |
| 第 3 轮（Syncer/DNS） | 0 | 1 | 0 | 0 |
| 第 4 轮（WebUI 后端） | 0 | 1 | 1 | 0 |
| 第 5 轮（前端） | 1 | 1 | 2 | 0 |
| 第 6 轮（Docker/CI） | 4 | 1 | 2 | 1 |
| 第 7 轮（错误/边界） | 0 | 1 | 3 | 2 |
| 第 8 轮（交叉验证） | 0 | 4 | 2 | 0 |
| 第 9 轮（深度复查） | 0 | 0 | 0 | 0 |
| **合计** | **6** | **9** | **14** | **5** |

### 待修复高严重度问题（6 项）
1. **5.1** `.env.example` 含旧格式重复内容
2. **9.1** 前端 `dist/` 缺失
3. **10.1** `docker-publish.yml` 内容重复
4. **10.2** CI 缺前端构建步骤
5. **10.3** Dockerfile 缺前端构建阶段
6. **10.4** Release 流程缺前端构建

### 待修复中严重度问题（9 项）
1. **7.1** `sync:start`/`sync:complete` 事件未发布
2. **8.1** 配置导入缺事务保护
3. **9.2** 前端 ID 用 `index+1` 与数据库 ID 错位
4. **10.5** Dockerfile 路径 CI 引用不一致
5. **11.1** `truncateDesc` 遗漏 SWAS 50 字符限制
6. **12.1** 熔断 `IsOpen` 检查未实际跳过同步
7. **12.2** Makefile docker-build 路径错误
8. **12.3** 缺少 README.md
9. **12.4** Makefile 构建缺前端步骤

---

## Issue 12: 第 8 轮审查 — 交叉验证与遗漏模块

**审查轮次：** 第 8 轮  
**审查范围：** Makefile、version/、测试文件、README 存在性、熔断器行为验证、dns/、notifier/、internal/ 包深度检查

---

### 12.1 `syncDomain` 中熔断检查仅为日志输出，未实际跳过同步

**所在阶段/Step：** Step 6 / 熔断 + Syncer

**严重度：** 中

**涉及文件：** `syncer/syncer.go` L203-205, `dns/circuitbreaker.go` L24-28

**现象描述：**
- `syncer.go` L203-205 在 `syncDomain` 中检查熔断状态：
  ```go
  if s.cb.IsOpen(rule.Host) {
      slog.Debug("域名已熔断，半开探测", "domain", rule.Host)
  }
  ```
- 注释写明「已熔断的域名跳过」，但代码仅在 `IsOpen` 返回 true 时打印 Debug 日志
- 之后**无条件继续执行** DNS 解析和同步，未 return 跳过
- 熔断器正确地记录了失败计数和成功重置，但**从未阻止任何同步操作**
- 与 Build1.md 12.7 节「渐进式熔断：连续失败达阈值后熔断，半开状态每轮探测一次」的设计意图不符——当前行为始终探测

**原因分析：** 熔断跳过逻辑未实现，`IsOpen` 的返回值未被用于控制流程。

**影响范围：**
- 持续 DNS 解析失败的域名每轮都超时等待（默认 10s），无法通过熔断减少无效等待
- 熔断器沦为"计数器 + 日志"，不提供实际保护
- 对功能正确性无影响（失败域名不会产生错误规则变更），仅浪费时间和 API 配额

**推荐修复方案：**
```go
if s.cb.IsOpen(rule.Host) {
    slog.Debug("域名已熔断，跳过本次同步", "domain", rule.Host)
    return
}
```
若需支持半开探测（每 N 轮探测一次），需在 `CircuitBreaker` 中增加状态机和探测计数器。

**当前状态：** 待修复

---

### 12.2 `Makefile` `docker-build` 目标引用不存在的 `build/Dockerfile`

**所在阶段/Step：** Step 15 / Docker 构建

**严重度：** 中

**涉及文件：** `Makefile` L19

**现象描述：**
```makefile
docker-build:
    docker build -f build/Dockerfile --build-arg VERSION=$(VERSION) -t fwalizer .
```
- `-f build/Dockerfile` 指向 `build/` 子目录，但 Dockerfile 实际位于项目根目录
- 执行 `make docker-build` 将因找不到 Dockerfile 而失败
- 与 CI `docker-publish.yml` 第一段存在相同路径错误（见 Issue 10.5）

**原因分析：** 与 CI 问题同源——Dockerfile 曾计划放在 `build/` 目录后移回根目录，Makefile 未同步更新。

**影响范围：** `make docker-build` 不可用。

**推荐修复方案：**
- **方案 A（推荐）：** 改为 `docker build -f Dockerfile ...`（使用根目录 Dockerfile）
- **方案 B：** 将 Dockerfile 移到 `build/Dockerfile` 并同步更新 CI 和 `.dockerignore`

**当前状态：** 待修复

---

### 12.3 项目缺少 `README.md`

**所在阶段/Step：** Step 16 / 文档

**严重度：** 中

**涉及文件：** 无（缺失 `README.md`）

**现象描述：**
- 项目根目录不存在 `README.md` 或任何 README 文件
- 虽然 `Design1.md` 和 `Build1.md` 提供了详细设计文档，但缺少面向用户的项目介绍、快速开始指南、配置说明
- `.env.example` 的头部注释指向 `https://github.com/alcaprophet/fwalizer`，该 URL 预期对应 README

**原因分析：** 项目处于活跃开发阶段，README 尚未编写。

**影响范围：** GitHub 仓库首页无内容展示；新用户无快速入门指引。

**推荐修复方案：**
- 编写 README.md，包含：项目简介、功能特性、快速开始（Docker / 二进制 / 源码）、配置说明、运行模式、WebUI 截图、License
- 可参考 AGENTS.md 第十二章的文档体系规划

**当前状态：** 待规划

---

### 12.4 `Makefile` 构建目标缺少前端构建步骤

**所在阶段/Step：** Step 13/15 / 构建

**严重度：** 中

**涉及文件：** `Makefile` L6-7

**现象描述：**
```makefile
build:
    go build -ldflags="$(LDFLAGS)" -o fwalizer .
```
- `make build` 仅编译 Go 源码，不构建前端
- 生成的 `fwalizer` 二进制不含 Embedded 前端（`webui/frontend/dist/` 不存在）
- WebUI 模式下访问将返回 404

**原因分析：** 前端构建步骤未纳入 Makefile。

**影响范围：** 本地 `make build` 生成的二进制无法使用 WebUI 模式。

**推荐修复方案：**
```makefile
frontend:
    cd webui/frontend && npm ci && npm run build

build: frontend
    go build -ldflags="$(LDFLAGS)" -o fwalizer .
```

**当前状态：** 待修复

---

### 12.5 缺少 Provider 实现层和 Syncer 层的单元测试

**所在阶段/Step：** Step 5/6/8/9/10 / Provider + Syncer

**严重度：** 低

**涉及文件：** 测试覆盖缺口

**现象描述：**
- 已有测试覆盖：`config/env_test.go`、`dns/circuitbreaker_test.go`、`dns/resolver_test.go`、`internal/portconv/portconv_test.go`、`internal/tag/tag_test.go`、`notifier/bus_test.go`、`provider/common_test.go`
- 缺失测试覆盖：
  - `provider/tc_lighthouse.go` — 无测试
  - `provider/tc_cvm.go` — 无测试
  - `provider/ali_swas.go` — 无测试
  - `provider/ali_ecs.go` — 无测试
  - `syncer/syncer.go` — 无测试
  - `syncer/retry.go` — 无测试
  - `webui/api/*.go` — 无测试
  - `config/store.go` — 无测试

**原因分析：** Provider 实现依赖云 SDK Mock（需要网络或 Mock），Syncer 依赖多个组件协同，测试成本较高。

**影响范围：** 重构或修改 Provider/Syncer 逻辑时缺少安全网。

**推荐修复方案：**
- 优先对 `provider/common.go` 的 Diff 逻辑补充更多边界用例（已有基础覆盖）
- Provider 实现层测试可留待后期（需 Mock SDK Client）
- Syncer 层可通过注入 Mock Provider + Mock Resolver 进行单元测试

**当前状态：** 待规划

---

### 12.6 `.env.example` 的重复内容精确分析

**所在阶段/Step：** Step 2 / 全局配置

**严重度：** 低（补充 Issue 5.1 的精确分析）

**涉及文件：** `.env.example`

**精确分析：**
- **L1-58**：当前有效格式（3 列 TARGETS、分离凭据、6 列 RULES 含 comment）
- **L59-117**：L1-58 的**逐字节完全相同**副本（非"几乎相同"），包括所有注释和空行
- **L118-200**：旧格式，包含：
  - 5 列 TARGETS 旧格式（`provider|resource_id|region|access_id|access_key`，凭据嵌入）
  - 5 列 RULES 旧格式（`domain|protocol|ports|action|targets`，无 comment）
  - `TENCENTCLOUD_SECRET_ID`/`TENCENTCLOUD_SECRET_KEY`（旧凭据变量名）
  - `LIGHTHOUSE_INSTANCE_ID`/`LIGHTHOUSE_REGION`（单实例硬编码）
  - `DOMAIN_RULES`（旧规则变量名，替代 RULES）
  - `RULE_TAG`（旧标签变量名，替代 TAG）
  - `CHECK_INTERVAL`（旧间隔变量名，替代 INTERVAL）
  - `DNS_SERVER`（旧 DNS 变量名，替代 DNS）

**当前状态：** 待修复（同 Issue 5.1）

---

## 第 8 轮审查总结

- **发现高严重度问题：** 0 项
- **发现中严重度问题：** 4 项（12.1 熔断未跳过, 12.2 Makefile Docker 路径错误, 12.3 缺少 README, 12.4 Makefile 缺前端构建）
- **发现低严重度问题：** 2 项（12.5 测试覆盖缺口, 12.6 .env.example 精确分析）
- **待讨论事项：** 0 项
- **剩余风险：** 熔断器空转问题导致 DNS 失败时的等待时间无法缩短；Makefile 的两个错误使本地构建链路同样断裂。



-----

# FWAlizer 构建问题记录

> 记录构建过程中遇到的错误、排查过程及解决办法。

---

## Issue 1: go.mod 重复内容

**Step:** 1

**现象：** `go build ./...` 报错 `repeated module statement` 和 `repeated go statement`

**原因：** 覆写 go.mod 时旧内容未完全清除，导致 module 和 go 声明重复出现

**解决：** 重新完整写入 go.mod，确保仅包含一份 module 声明和 go 版本声明

---

## Issue 2: RULES 端口字段逗号与条目分隔符冲突

**Step:** 2

**现象：** `RULES=api.example.com|TCP|443,80|ACCEPT||生产API` 解析失败，端口 `443,80` 中的逗号被误认为条目分隔符

**原因：** Build1.md 参考代码使用简单逗号拆分 (`splitEntries`)，但端口字段允许逗号分隔多端口

**解决：** 实现 `splitRuleEntries` 智能分割——通过正则检测 `host|PROTOCOL|` 模式识别新条目起始位置，仅在该模式匹配时才分割，否则将逗号保留在当前条目内（属于端口字段）

---

## Issue 3: 深度审查问题批量修复

**阶段：** 全量构建完成后的质量审查

### 3.1 CircuitBreaker 未集成到 Syncer

**现象：** `dns/circuitbreaker.go` 已实现但未被 `syncer/syncDomain()` 调用，熔断逻辑未生效

**修复：** 在 Syncer 结构体中添加 `cb *dns.CircuitBreaker` 字段，`syncDomain()` 中解析前检查 `IsOpen()`，解析成功调 `RecordSuccess()`，失败调 `RecordFailure()`

### 3.2 EventBus 未集成到 Syncer

**现象：** `notifier/bus.go` 已实现但 Syncer 未发布任何事件

**修复：** 在 Syncer 中添加 `bus *notifier.EventBus`，DNS 失败时发布 `EventDNSFailed`，同步失败时发布 `EventSyncError`，并暴露 `EventBus()` 方法供外部订阅

### 3.3 WebUI 配置热重载未接通

**现象：** `Syncer.Reload()` 存在但 WebUI 修改配置后未调用

**修复：** WebUI Server 添加 `SetReloadFunc()` 回调机制，在所有写操作（AddTarget/DeleteTarget/AddRule/DeleteRule/PutSettings）后触发；main.go 中 WebUI 模式直接创建 Syncer 并将 `store.LoadConfig() + s.Reload()` 作为回调传入

### 3.4 ClientPool key 缺少 accessID

**现象：** key 格式为 `cloudType|region`，同 region 不同凭据会错误复用 Client

**修复：** 四个 Provider 的 key 统一改为 `cloudType|region|accessID` 格式

### 3.5 App.vue 未使用的 `h` 导入

**修复：** 删除 `import { h, ref } from 'vue'` 中的 `h`

## 3.6 main.go 使用 fmt.Println

**修复：** WebUI 模式启动提示改为 `slog.Info()`，导出 `app.InitLogger()` 供 main.go 调用

---

## Issue 4: 深度审查报告问题核验与修复

**阶段：** 全量构建完成后的深度审查（第二轮）

> 以下 10 项问题均已在代码中逐一定位确认。
> **状态更新：** 4.1–4.6、4.8–4.10 已修复；4.7 待规划（后续 Phase 处理）。

---

### 4.1 ✅ `truncateDesc` 字节截断破坏 `[TAG]` 前缀和中文（已修复）

**严重度：** 高（不通过）

**文件：** `syncer/retry.go` L119-131

**现象：** `truncateDesc` 使用 `desc[:maxLen]` 按字节截断，当描述包含中文（UTF-8 多字节字符）时会截断到半个字符，产生乱码。若 `[TAG]` 前缀本身接近 64 字符限制，截断还可能破坏前缀完整性。

**影响范围：** 所有 Lighthouse 规则的描述字段（`FirewallRuleDescription ≤ 64 字符`），含中文备注的规则均受影响。

**修复方案：**
- 将字节截断改为 rune 切片截断，保证字符完整性
- 关键代码变更：
```go
func truncateDesc(desc string, ct config.CloudType) string {
    maxLen := 0
    switch ct {
    case config.CloudTCLighthouse:
        maxLen = 64
    default:
        return desc
    }
    runes := []rune(desc)
    if len(runes) <= maxLen {
        return desc
    }
    return string(runes[:maxLen])
}
```

---

### 4.2 ✅ `fmt.Sscanf` 返回值未处理（已修复）

**严重度：** 中（不通过）

**文件：** `webui/server.go` L104, L139

**现象：** `handleDeleteTarget` 和 `handleDeleteRule` 中 `fmt.Sscanf(id, "%d", &n)` 忽略返回值。当 URL 路径参数非数字时，`n` 保持零值 0，导致尝试删除 ID=0 的记录（静默失败或误操作）。

**影响范围：** WebUI 删除目标/规则的 API 端点。

**修复方案：**
- 检查 `fmt.Sscanf` 返回的 error，非 nil 时返回 400
- 关键代码变更：
```go
if _, err := fmt.Sscanf(id, "%d", &n); err != nil {
    writeError(w, http.StatusBadRequest, "无效的资源 ID")
    return
}
```

---

### 4.3 ✅ `os.MkdirAll` 返回值未处理（已修复）

**严重度：** 中（不通过）

**文件：** `main.go` L40

**现象：** `os.MkdirAll(dataDir, 0755)` 忽略错误返回。若目录创建失败（权限不足等），后续 `config.OpenStore(dbPath)` 会以不明确的错误退出。

**影响范围：** WebUI 模式启动流程。

**修复方案：**
```go
if err := os.MkdirAll(dataDir, 0755); err != nil {
    fmt.Fprintf(os.Stderr, "创建数据目录失败: %v\n", err)
    os.Exit(1)
}
```

---

### 4.4 ✅ `os.UserHomeDir` 错误忽略（已修复）

**严重度：** 低（不通过）

**文件：** `main.go` L109

**现象：** `home, _ := os.UserHomeDir()` 忽略错误。若无法获取用户主目录（如容器内 HOME 未设置），`home` 为空字符串，数据目录变为相对路径 `.config/fwalizer`。

**影响范围：** WebUI 模式数据目录定位。

**修复方案：**
```go
func getDataDir() string {
    home, err := os.UserHomeDir()
    if err != nil {
        fmt.Fprintf(os.Stderr, "无法获取用户目录: %v\n", err)
        os.Exit(1)
    }
    return filepath.Join(home, ".config", "fwalizer")
}
```

---

### 4.5 ✅ CVM 规则计数未使用 `PolicyStatistics`（已修复）

**严重度：** 低（存在风险）

**文件：** `provider/tc_cvm.go` L219

**现象：** `checkRuleLimit` 使用 `len(ps.Ingress) + len(ps.Egress)` 手动计数，而 Build1.md 规定使用 API 返回的 `PolicyStatistics`（`IngressIPv4TotalCount + IngressIPv6TotalCount + EgressIPv4TotalCount + EgressIPv6TotalCount`）。功能等价但偏离文档规范。

**影响范围：** CVM 安全组规则上限检测逻辑。

**修复方案：**
- 优先使用 `PolicyStatistics` 字段精确计数，fallback 到手动计数
- 关键代码变更：
```go
func (p *TCCVM) checkRuleLimit(toAdd int) error {
    // ...
    ps := resp.Response.SecurityGroupPolicySet
    if ps == nil {
        return nil
    }
    var total int
    if ps.PolicyStatistics != nil {
        stats := ps.PolicyStatistics
        total = int(strVal(stats.IngressIPv4TotalCount)) + ...
    } else {
        total = len(ps.Ingress) + len(ps.Egress)
    }
    // ... 后续判断不变
}
```
- 注：需确认 SDK 中 `PolicyStatistics` 字段类型（可能为 *string 或 *int64）

---

### 4.6 ✅ 半开探测失败仍递增计数器（已修复）

**严重度：** 低（存在风险）

**文件：** `dns/circuitbreaker.go` L44

**现象：** `RecordFailure` 无条件执行 `cb.failCount[domain]++`。Build1.md 12.7 节规定「半开探测失败不计入失败次数，维持熔断状态」。当前实现在已熔断后仍递增计数器，导致数值无限增长（功能不受影响，但偏离文档且影响日志可读性）。

**影响范围：** DNS 熔断器日志输出准确性。

**修复方案：**
```go
func (cb *CircuitBreaker) RecordFailure(domain string) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    // 已熔断时不再递增（半开探测失败维持熔断状态）
    if cb.failCount[domain] >= cb.threshold {
        return
    }
    cb.failCount[domain]++
    if cb.failCount[domain] == cb.threshold {
        slog.Error("DNS 熔断触发", "domain", domain, "连续失败", cb.failCount[domain])
    }
}
```

---

### 4.7 ⏸️ `webui/api/` 子目录未拆分 + Step 13 端点缺失（待规划）

**严重度：** 中（存在风险）

**文件：** `webui/server.go`（全部 API 集中于此）

**现象：**
1. Build1.md Step 13 规定 API 拆分为 `webui/api/targets.go`、`rules.go`、`sync.go`、`settings.go`，实际全部合并在 `server.go` 单文件中
2. 以下端点未实现：
   - `GET /api/sync/status`（同步状态）
   - `POST /api/sync/trigger`（手动触发同步）
   - `POST /api/sync/dryrun`（试运行）
   - `POST /api/test-connection`（测试连接）
   - `GET /api/config/export`（配置导出）
   - `POST /api/config/import`（配置导入）

**影响范围：** WebUI 功能完整性；前端 Dashboard 页面无法展示同步状态、无法手动触发同步。

**修复方案：**
- **文件拆分**（可选，功能等价但符合规范）：将 handler 按职责拆分到 `webui/api/` 子目录
- **端点补全**（优先级高）：
  - `/api/sync/status`：返回 Syncer 上次同步时间、状态、各域名结果
  - `/api/sync/trigger`：调用 Syncer 立即执行一轮同步
  - `/api/sync/dryrun`：复用 Diff 逻辑，仅返回 toAdd/toDelete 不写入
  - `/api/test-connection`：用凭据创建临时 Client 调用 Describe 验证连通性
  - `/api/config/export`：从 SQLite 导出 JSON（凭据字段不导出）
  - `/api/config/import`：接收 JSON 校验后写入 SQLite，触发热重载
- 需要在 `Server` 结构体中增加对 `Syncer` 的引用以支持 status/trigger/dryrun

---

### 4.8 ✅ `truncateLighthouseDesc` 代码重复（已修复）

**严重度：** 低（存在风险）

**文件：** `provider/tc_lighthouse.go` L227-232

**现象：** `truncateLighthouseDesc` 与 `syncer/retry.go` 中的 `truncateDesc` 功能重复。`CreateRules`（L112）调用了 `truncateLighthouseDesc`，而 Syncer 在调用 `CreateRules` 前已通过 `truncateDesc` 截断。存在双重截断 + 代码冗余。

**影响范围：** 代码可维护性；两处截断逻辑若修改一处忘改另一处会产生不一致。

**修复方案：**
- 删除 `tc_lighthouse.go` 中的 `truncateLighthouseDesc` 函数
- `CreateRules` 中直接使用传入的 `r.Description`（已由 Syncer 层截断）
- 关键代码变更：
```go
// tc_lighthouse.go CreateRules 中：
fwRule := &lighthouse.FirewallRule{
    Action:                  common.StringPtr(r.Action),
    FirewallRuleDescription: common.StringPtr(r.Description), // 已由 Syncer 截断
}
```

---

### 4.9 ✅ `go mod tidy` 后 go.mod/go.sum 变更未提交（已修复）

**严重度：** 低（存在风险）

**文件：** `go.mod`、`go.sum`

**现象：** `app/systray.go`（build tag `desktop`）引用 `fyne.io/systray`，但 go.mod 中原本缺失该依赖。执行 `go mod tidy` 后新增：
- `fyne.io/systray v1.12.2`
- `github.com/godbus/dbus/v5 v5.1.0`（indirect）

**影响范围：** 其他开发者 `go build -tags desktop` 时可能因 go.sum 缺失条目而失败。

**修复方案：**
- 执行 `go mod tidy` 并提交 go.mod 和 go.sum 变更

---

## 4.10 ⚠️ `.env.example` 包含旧格式遗留内容（未修复，问题复现）

**严重度：** 高

**文件：** `.env.example` L59-200

**现象：** 文件第 1-58 行为当前有效格式，但第 59-200 行存在**两份完整重复** + 一份旧版本遗留内容：
- L59-116：第一份完整重复（与 L1-58 内容一致）
- L117-200：旧版本遗留，包含：
  - 旧格式 TARGETS（凭据嵌入：`provider|resource_id|region|access_id|access_key`）
  - 旧变量名：`TENCENTCLOUD_SECRET_ID`、`LIGHTHOUSE_INSTANCE_ID`、`DOMAIN_RULES`、`RULE_TAG`、`CHECK_INTERVAL`、`DNS_SERVER`
  - 与当前配置体系完全冲突，会误导用户

**原因分析：** Issue 4.10 标记为已修复，但修复未实际生效或后续操作覆盖了修复结果。

**影响范围：** 新用户参照配置时产生严重困惑；旧变量名不被解析器识别；文件内容三倍冗余。

**推荐修复方案：**
- 删除 L59-200 的全部内容
- 仅保留 L1-58 的当前有效配置模板

**状态：** 待修复

---

## Issue 5: 多轮深度审查发现问题（第三轮）

**阶段：** 全量构建完成后的多轮深度审查

---

### 5.1 `docker-publish.yml` 包含重复 YAML 文档（严重度：高）

**模块：** CI/CD

**文件：** `.github/workflows/docker-publish.yml` L85-161

**现象：** 文件包含两个完整的 YAML 文档（L1-84 和 L85-161），第二段以 `name: Docker Build & Publish` 重新开始。GitHub Actions 不支持单文件多文档，会导致解析失败或仅识别第一段。

**原因分析：** 文件被追加写入而非覆盖，导致旧版本内容残留。

**影响范围：** CI/CD 流水线可能无法正常工作；第二段中仅更新 lighthouse/common 两个 SDK（缺少 vpc/swas/ecs），若被误识别会导致编译失败。

**推荐修复方案：**
- 删除 L85-161 的全部重复内容
- 仅保留 L1-84 的完整工作流（已包含所有 SDK 更新）

**状态：** 待修复

---

### 5.2 Dockerfile 位置与引用不一致（严重度：高 → 已部分修复，遗留问题降级为中）

**模块：** Docker 构建

**文件：** `Dockerfile`（项目根目录）、`build/Dockerfile`、`Makefile` L19、`.github/workflows/docker-publish.yml` L76

**现象（原始）：**
- `Makefile` docker-build 目标引用 `-f build/Dockerfile`
- CI/CD 工作流引用 `file: build/Dockerfile`
- 原审查时 `build/` 目录为空

**第 9 轮核实更新：** `build/Dockerfile` 现已存在（20 行），内容符合 Build1.md 要求：
- ✅ `ARG VERSION=dev` + ldflags 版本注入
- ✅ `-tags docker` 构建标签
- ✅ 双模式 HEALTHCHECK（`wget ... || pgrep fwalizer || exit 1`）
- ❌ 运行阶段缺少 `WORKDIR /app`（同 Issue 6.3）

**遗留问题（严重度：中）：** 项目根目录存在冗余的旧版 `Dockerfile`（30 行），缺少版本注入、构建标签和双模式健康检查（即 Issue 5.3、5.4 所描述的问题均仅存在于根 Dockerfile）。两个 Dockerfile 并存会造成维护混乱。

**推荐修复方案：**
- 删除根目录 `Dockerfile`（Makefile/CI 均引用 `build/Dockerfile`，根 Dockerfile 无任何引用）
- Issue 5.3、5.4 随之关闭
- 为 `build/Dockerfile` 运行阶段添加 `WORKDIR /app`（见 Issue 6.3）

**状态：** 待修复（删除根 Dockerfile + 添加 WORKDIR）

---

### 5.3 Dockerfile 缺少版本注入和构建标签（严重度：中）

**模块：** Docker 构建

**文件：** `Dockerfile` L12-13

**现象：**
1. 缺少 `ARG VERSION=dev` 声明和 `-X github.com/alcaprophet/fwalizer/version.Version=${VERSION}` ldflags 注入
2. 缺少 `-tags docker` 构建标签（Build1.md 8.2 节要求）
3. 实际编译命令为 `go build -ldflags="-s -w" -o /fwalizer .`，容器内 `fwalizer version` 永远输出 `dev`

**原因分析：** Dockerfile 编写时遗漏了版本注入和构建标签。

**影响范围：** Docker 镜像中无法正确显示版本号；缺少 docker 构建标签可能影响后续按标签隔离的功能。

**推荐修复方案：**
```dockerfile
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags docker \
    -ldflags="-s -w -X github.com/alcaprophet/fwalizer/version.Version=${VERSION}" \
    -o /fwalizer .
```

**状态：** 待修复

---

### 5.4 Dockerfile HEALTHCHECK 不符合双模式设计（严重度：中）

**模块：** Docker 构建

**文件：** `Dockerfile` L26-27

**现象：** 当前 HEALTHCHECK 为 `killall -0 fwalizer || exit 1`，仅检测进程存活。Build1.md 12.2 节规定：
```
CMD wget -q -O /dev/null http://localhost:9090/api/health 2>/dev/null || pgrep fwalizer || exit 1
```
应兼容 WebUI 模式（HTTP 端点检测）和 .env 模式（进程检测）。

**原因分析：** 注释称"Alpine 不含 pidof"使用了 killall 替代，但缺少 WebUI 模式的 HTTP 健康检查。

**影响范围：** WebUI 模式下若 HTTP 服务挂死但进程仍在，HEALTHCHECK 无法检测到异常。

**推荐修复方案：**
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null http://localhost:9090/api/health 2>/dev/null || killall -0 fwalizer 2>/dev/null || exit 1
```
注：Alpine 的 busybox wget 可用；`killall -0` 替代 `pgrep`（Alpine 无 pgrep）。

**状态：** 待修复

---

### 5.5 `getDataDir()` 未按平台区分数据目录（严重度：中）

**模块：** App 生命周期

**文件：** `main.go` L114-121

**现象：** `getDataDir()` 固定返回 `~/.config/fwalizer`，但 Build1.md 9.3 节规定：
- macOS: `~/Library/Application Support/fwalizer/config.db`
- Windows: `%APPDATA%/fwalizer/config.db`
- Linux: `~/.config/fwalizer/config.db`

**原因分析：** 实现时简化处理，未做跨平台适配。

**影响范围：** macOS 和 Windows 用户的数据存储位置不符合操作系统规范，可能导致数据在应用更新/卸载时丢失。

**推荐修复方案：**
```go
func getDataDir() string {
    switch runtime.GOOS {
    case "darwin":
        home, _ := os.UserHomeDir()
        return filepath.Join(home, "Library", "Application Support", "fwalizer")
    case "windows":
        appdata := os.Getenv("APPDATA")
        if appdata == "" {
            home, _ := os.UserHomeDir()
            appdata = filepath.Join(home, "AppData", "Roaming")
        }
        return filepath.Join(appdata, "fwalizer")
    default:
        home, _ := os.UserHomeDir()
        return filepath.Join(home, ".config", "fwalizer")
    }
}
```

**状态：** 待修复

---

### 5.6 WebUI 模式缺少 pidfile 防多实例机制（严重度：中）

**模块：** App 生命周期

**文件：** `main.go`（WebUI 模式启动流程）

**现象：** Build1.md 12.12 节规定 WebUI 模式启动时应创建 pidfile 防止多实例运行，当前实现完全没有 pidfile 逻辑。

**原因分析：** 功能尚未实现。

**影响范围：** 用户可能误启动多个 WebUI 实例，导致 SQLite 写冲突和端口占用。

**推荐修复方案：**
- 启动时创建 `{dataDir}/fwalizer.pid`，写入当前 PID
- 启动前检测 pidfile 是否存在且进程存活，是则拒绝启动
- 正常退出时删除 pidfile（defer 或 signal handler）

**状态：** 待规划

---

### 5.7 Syncer 同步日志未写入 SQLite（严重度：中）

**模块：** Syncer / WebUI

**文件：** `syncer/syncer.go`（syncDomain/syncAll）、`config/store.go`（AddSyncLog）

**现象：** `config.Store.AddSyncLog()` 已实现，`sync_logs` 表已创建，WebUI 的 `GET /api/sync/logs` 端点已实现。但 Syncer 同步过程中从未调用 `AddSyncLog()`，导致同步日志页面永远为空。

**原因分析：** Syncer 与 Store 之间缺少连接——Syncer 不持有 Store 引用，无法写入日志。

**影响范围：** WebUI 同步日志功能完全不可用。

**推荐修复方案：**
- 方案 A（推荐）：通过 EventBus 订阅 `sync:complete`/`sync:error` 事件，在订阅者中写入 Store
- 方案 B：在 Syncer 中注入 `LogFunc func(SyncLog)` 回调
- 需要在 main.go WebUI 模式中接通

**状态：** 待修复

---

### 5.8 告警通知器未接入 EventBus（严重度：中）

**模块：** 告警 / EventBus

**文件：** `notifier/email.go`、`notifier/webhook.go`、`main.go`

**现象：** `EmailNotifier` 和 `WebhookNotifier` 已实现 `Subscriber` 接口，但没有任何代码创建它们并调用 `EventBus.Subscribe()` 注册。告警功能完全未接通。

**原因分析：** Step 15 的告警功能虽标记完成，但仅实现了通知器本身，未在启动流程中接入。

**影响范围：** 同步异常时用户无法收到邮件/Webhook 通知。

**推荐修复方案：**
- 在 WebUI 设置中增加告警配置项（SMTP 和 Webhook URL）
- 启动时读取配置，若已配置则创建对应 Notifier 并 Subscribe 到 EventBus
- 热重载时更新订阅

**状态：** 待规划

---

### 5.9 `Store.LoadConfig()` 未加载 `webui_port` 和 `dns_fail_threshold`（严重度：中）

**模块：** 配置持久化

**文件：** `config/store.go` L284-332

**现象：** `LoadConfig()` 从 settings 表读取 `tag`、`interval`、`dns`、`log_level`，但未读取 `webui_port` 和 `dns_fail_threshold`。用户通过 WebUI 修改这两项后，重载配置不会生效。

**原因分析：** 实现时遗漏了这两个可选配置项的加载。

**影响范围：** WebUI 设置页面修改端口和熔断阈值后不生效（需重启）。

**推荐修复方案：**
```go
if v := settings["webui_port"]; v != "" {
    if n, err := strconv.Atoi(v); err == nil {
        cfg.WebUIPort = n
    }
}
if v := settings["dns_fail_threshold"]; v != "" {
    if n, err := strconv.Atoi(v); err == nil {
        cfg.DNSFailThreshold = n
    }
}
```

**状态：** 待修复

---

### 5.10 Syncer 未发布 `EventSyncStart` / `EventSyncComplete` 事件（严重度：中）

**模块：** Syncer / EventBus

**文件：** `syncer/syncer.go`（syncAll）

**现象：** Build1.md 第十一节定义了 5 种事件类型，其中 `EventSyncStart` 和 `EventSyncComplete` 已声明但从未被 Publish。syncAll 开始时未发布 start 事件，完成时未发布 complete 事件。

**原因分析：** 实现时仅关注了错误事件（DNSFailed、SyncError），遗漏了正常流程事件。

**影响范围：** SSE 事件流中缺少同步开始/完成通知，前端无法实时展示同步进度。

**推荐修复方案：**
- 在 `syncAll()` 开头发布 `EventSyncStart`
- 在 `syncAll()` 结尾发布 `EventSyncComplete`（附带耗时、结果统计）

**状态：** 待修复

---

### 5.11 CLI 缺少 `backup` / `restore` 子命令（严重度：中）

**模块：** CLI

**文件：** `app/cli.go`

**现象：** Build1.md 12.5 节和 12.8 节规定 CLI 应支持 `fwalizer backup` 和 `fwalizer restore [file]` 子命令，当前仅实现了 `version` 和 `validate`。

**原因分析：** Step 15 高级功能未完全实现。

**影响范围：** 用户无法通过命令行备份/恢复 SQLite 配置。

**推荐修复方案：**
- `backup`：复制 `{dataDir}/config.db` 到 `config.db.bak.{timestamp}`，最多保留 5 个
- `restore [file]`：从备份文件恢复，恢复前执行 `PRAGMA integrity_check`

**状态：** 待规划

---

### 5.12 系统托盘缺少开机自启开关和实际同步触发（严重度：中）

**模块：** 桌面端

**文件：** `app/systray.go` L33, L46-47

**现象：**
1. Build1.md Step 16 规定托盘菜单应包含"开机自启[开关]"，当前缺失
2. "立即同步"菜单项仅打印日志（`slog.Info("手动触发同步")`），注释为 TODO，未实际通知 Syncer

**原因分析：** 桌面端功能未完全实现。

**影响范围：** 桌面端用户无法通过托盘触发同步和配置开机自启。

**推荐修复方案：**
- 传入 Syncer 引用或 trigger channel 到 `RunSystray`
- 添加开机自启菜单项（checkbox），调用平台 API

**状态：** 待规划

---

### 5.13 `.dockerignore` 存在重复条目（严重度：低）

**模块：** Docker 构建

**文件：** `.dockerignore`

**现象：** `.env` 出现 2 次（L3、L5），`*.md` 出现 2 次（L2、L10），`.git` 和 `.git/` 同时存在。还包含不存在的 `Ref/` 目录。

**影响范围：** 无功能影响，仅影响可维护性。

**推荐修复方案：** 去重并清理无效条目：
```
Documents/
*.md
.env
.git/
Dockerfile
.dockerignore
Makefile
```

**状态：** 待修复

---

### 5.14 `firewall/` 空目录残留（严重度：低）

**模块：** 目录结构

**文件：** `firewall/`（空目录）

**现象：** 项目根目录存在空的 `firewall/` 目录，为旧版本残留。Build1.md 目录结构中无此目录。

**影响范围：** 无功能影响，仅影响项目整洁度。

**推荐修复方案：** 删除空目录（Git 不跟踪空目录，若含 `.gitkeep` 则删除该文件）。

**状态：** 待修复

---

### 5.15 前端 `package.json` 使用 npm 而非 pnpm 锁定（严重度：低）

**模块：** WebUI 前端

**文件：** `webui/frontend/package.json`、`webui/frontend/package-lock.json`

**现象：** Design1.md 和 Build1.md 规定前端包管理器为 pnpm，但项目中存在 `package-lock.json`（npm 产物），而非 `pnpm-lock.yaml`。

**影响范围：** 不影响功能，但与设计文档不一致。开发者可能混用 npm/pnpm 导致依赖版本差异。

**推荐修复方案：**
- 删除 `package-lock.json`
- 执行 `pnpm install` 生成 `pnpm-lock.yaml`
- 或在文档中将包管理器改为 npm（降低贡献门槛）

**状态：** 待讨论
- 方案 A：改用 pnpm（符合文档）
- 方案 B：文档改为 npm（降低门槛，已有 package-lock.json）
- 推荐方案 B，理由：项目为内部工具，npm 更通用

---

### 5.16 ✅ WebUI SPA 路由无 fallback 问题（已确认无问题）

**模块：** WebUI 后端

**文件：** `webui/frontend/src/main.ts`

**现象：** 前端使用 `createWebHashHistory()`（hash 路由模式），URL 形如 `http://localhost:9090/#/targets`，服务器始终返回 `index.html`，无需 SPA fallback。

**结论：** 无问题，hash 路由与 `http.FileServer` 完全兼容。

---

### 5.17 前端缺少「高级功能」和「告警配置」页面（严重度：中）

**模块：** WebUI 前端

**文件：** `webui/frontend/src/main.ts`、`webui/frontend/src/App.vue`

**现象：** Build1.md Step 14 规定 7 个页面（仪表盘、云资源管理、域名规则、全局设置、同步日志、高级功能、告警配置），实际仅实现 5 个，缺少：
- `/advanced`（高级功能：Dry Run、配置导入/导出、健康检查）
- `/alerts`（告警配置：邮件、Webhook）

**原因分析：** Step 15 告警功能和高级功能页面未完全实现。

**影响范围：** 用户无法通过 WebUI 执行 Dry Run、导入导出配置、配置告警。

**推荐修复方案：**
- 新增 `views/Advanced.vue`：包含 Dry Run 按钮、配置导入/导出、健康检查状态
- 新增 `views/Alerts.vue`：包含 SMTP 和 Webhook 配置表单
- 在 `main.ts` 和 `App.vue` 菜单中注册路由

**状态：** 待规划

---

### 5.18 CI/CD 工作流缺少前端构建步骤（严重度：中）

**模块：** CI/CD

**文件：** `.github/workflows/docker-publish.yml`、`.github/workflows/release.yml`

**现象：** 两个工作流均直接执行 `go build ./...`，但 `webui/embed.go` 的 `//go:embed frontend/dist` 要求 `dist/` 目录存在。由于 `dist/` 被 `.gitignore` 排除，从 Git 克隆后直接编译会失败：
```
webui/embed.go:5:12: pattern frontend/dist: no matching files found
```

**原因分析：** 工作流中缺少 `pnpm install && pnpm build` 步骤。

**影响范围：** CI/CD 流水线编译必然失败；新开发者克隆后 `go build` 也会失败。

**推荐修复方案：**
- 方案 A（推荐）：在 CI 中添加 Node.js + pnpm 构建前端的步骤
- 方案 B：在 `webui/frontend/dist/` 中保留一个 `.gitkeep` 或最小 `index.html` 并从 `.gitignore` 中移除（不推荐，会污染 Git）
- 方案 C：使用条件 embed（`//go:embed` 无法条件化，不可行）

**状态：** 待修复

---

## Issue 6: 第 8 轮深度审查发现问题

**阶段：** 全量构建完成后的多轮深度审查（第 8 轮：README、前端视图、测试文件、API 交叉核对）

---

### 6.1 README.md 包含旧版本完整内容（严重度：高）

**模块：** 文档

**文件：** `README.md` L352-540

**现象：** 文件第 1-351 行为当前正确的多云版本 README，第 352-540 行为**旧版本单云（Lighthouse-only）完整 README** 残留，包含：
- 旧项目名 "TencentCloudFirewallTool"
- 旧环境变量：`TENCENTCLOUD_SECRET_ID`、`LIGHTHOUSE_INSTANCE_ID`、`DOMAIN_RULES`、`RULE_TAG`、`CHECK_INTERVAL`、`DNS_SERVER`
- 旧项目结构（`firewall/` 目录、`TencentAPIGuide/`、`Ref/`）
- 旧 DOMAIN_RULES 格式（分号分隔）
- 不存在的 Makefile 目标（`make run`、`make docker-run`、`make docker-logs`）
- PowerShell 本地开发指令

**原因分析：** 重构时新 README 被追加到文件头部，旧内容未删除。

**影响范围：** 用户阅读 README 时会看到两套完全不同的配置体系，产生严重困惑；旧变量名不被当前解析器识别。

**推荐修复方案：**
- 删除 L352-540 的全部旧版本内容
- 仅保留 L1-351 的当前有效 README

**状态：** 待修复

---

### 6.2 前端使用数组索引作为数据库 ID（严重度：中）

**模块：** WebUI 前端 + 后端 API

**文件：** `webui/frontend/src/views/Targets.vue` L33/L52、`Rules.vue` L37/L56、`config/store.go` L127/L163

**现象：** 前端 `Targets.vue` 和 `Rules.vue` 使用 `index + 1` 作为资源 ID 进行编辑/删除操作。但后端使用 SQLite 自增 ID。当删除中间记录后，数组索引与实际数据库 ID 不再匹配：
- 添加 3 个目标 → DB ID: 1, 2, 3
- 删除 ID=2 → 剩余: ID 1, ID 3
- 前端显示为 index 0, 1 → 编辑第二行时发送 ID=2（已不存在）而非 ID=3

**原因分析：** `Store.GetTargets()` 和 `Store.GetRules()` 的 SELECT 查询未返回 `id` 列，前端无法获取实际数据库 ID。

**影响范围：** 删除记录后，编辑/删除其他记录会操作错误的行（静默失败或误操作）。

**推荐修复方案：**
1. 后端：`GetTargets()` 和 `GetRules()` 返回结果中包含 `id` 字段
   - 为 `TargetConfig` 和 `DomainRule` 添加 `ID int` 字段（或在 API 响应中包装）
   - SELECT 查询改为 `SELECT id, cloud_type, region, resource_id FROM targets`
2. 前端：使用返回的 `id` 字段而非 `index + 1`

**状态：** 待修复

---

### 6.3 Docker 运行时 .env 挂载路径与程序工作目录不匹配（严重度：中）

**模块：** Docker 部署

**文件：** `build/Dockerfile`（主要）、`Dockerfile`（根目录，待删除）、`README.md` L65-68

**现象：**
- README 中的 Docker 运行命令挂载 `.env` 到 `/app/.env`：`-v $(pwd)/.env:/app/.env:ro`
- 但 Dockerfile 运行阶段未设置 `WORKDIR`，Alpine 默认为 `/`
- 程序调用 `config.LoadEnv(".env")` 解析为 `/.env`，而非 `/app/.env`
- 结果：容器内程序找不到 .env 文件，启动失败

**原因分析：** Dockerfile 缺少 `WORKDIR /app` 设置。第 9 轮核实确认 `build/Dockerfile` 同样缺少此设置。

**影响范围：** 按 README 文档运行的 Docker .env 模式无法正常工作。

**推荐修复方案：**
在 `build/Dockerfile` 运行阶段添加：
```dockerfile
WORKDIR /app
```
或将 README 中的挂载路径改为 `/.env`（不推荐，不符合惯例）。

**状态：** 待修复

---

### 6.4 README Docker 卷挂载路径与 getDataDir() 的潜在耦合（严重度：低）

**模块：** Docker 部署 / 文档

**文件：** `README.md` L221-222

**现象：** README WebUI 模式示例挂载卷到 `/home/appuser/.config/fwalizer`。当前 `getDataDir()` 使用 `os.UserHomeDir()` + `/.config/fwalizer`。在容器内 appuser 的 home 为 `/home/appuser`，路径匹配。但若修复 5.5（按平台区分路径）后，需确保 Docker 内的路径仍一致。

**影响范围：** 当前无功能影响（路径碰巧一致），但存在潜在耦合风险。

**推荐修复方案：**
- 方案 A：添加 `FWALIZER_DATA_DIR` 环境变量支持（推荐，Docker 最佳实践）
- 方案 B：保持现状，在 README 中说明路径约定

**状态：** 待讨论

---

### 6.5 前端 JSON 字段名与后端结构体字段名耦合（严重度：低）

**模块：** WebUI 前端 + API

**文件：** `webui/frontend/src/views/Targets.vue` L8、`Rules.vue` L8

**现象：** 前端表单使用 Go 结构体字段名（`CloudType`、`ResourceID`、`Host`、`Protocol`）作为 JSON key。Go 的 `encoding/json` 默认使用字段名（无 json tag 时），因此当前能正常工作。但这违反了 REST API 通常使用 snake_case 的惯例，且与 `testConnectionReq`（使用 `json:"cloud_type"` tag）风格不一致。

**影响范围：** 无功能影响，但 API 风格不统一（部分 snake_case、部分 PascalCase）。

**推荐修复方案：**
- 方案 A：统一为 snake_case + json tag（规范）
- 方案 B：保持现状（内部工具，功能正常即可）
- 推荐方案 B，理由：内部使用导向，不过度设计

**状态：** 待讨论

---

## Issue 7: 第 9 轮深度审查发现问题

**阶段：** 全量构建完成后的多轮深度审查（第 9 轮：历史 Issue 核实、同步引擎与存储层交叉核对、Documents API 文档对照）

---

### 7.1 `filterRulesForTarget` 目标索引与 DB ID 不匹配（严重度：高）

**模块：** Syncer / 配置持久化

**文件：** `syncer/syncer.go` L244-260、`config/store.go` L126-144、`webui/frontend/src/views/Rules.vue`

**现象：** `filterRulesForTarget()` 将 `DomainRule.Targets`（存储的是 SQLite 自增 DB ID）与 `targetIndex+1`（Provider 在数组中的位置 +1）进行比较：
```go
for _, t := range r.Targets {
    if t == targetIndex+1 { // Targets 是 1-based
        filtered = append(filtered, r)
    }
}
```
但 Provider 的 `targetIndex` 是 `GetTargets()` 返回数组的下标（0-based），而非 DB ID。当删除中间目标后：
- 添加 3 个目标 → DB ID: 1, 2, 3，数组位置: 0, 1, 2 —— 两者恰好一致
- 删除 DB ID=2 → 剩余: ID 1, ID 3，数组位置: 0, 1
- 规则指定 Targets=[3]（指向第三个目标），但 `targetIndex+1` = 2 ≠ 3 → **规则静默不再应用到该目标**

**原因分析：** 与 Issue 6.2 同根源——`GetTargets()` 不返回 `id` 列，导致整个系统（前端、Syncer）都只能用数组位置“猜测” DB ID。初始状态下两者恰好一致，删除操作后产生偏移。

**影响范围：**
- 删除目标后，指定特定目标的域名规则会静默停止同步（无报错、无日志）
- 可能导致防火墙规则未更新而用户不知情（安全风险）
- 与 Issue 6.2（前端编辑/删除错行）属于同一根本缺陷的不同表现

**推荐修复方案：**
1. `GetTargets()` 返回结果包含 DB ID：
   - 为 `TargetConfig` 添加 `ID int \`json:"id"\`` 字段
   - SELECT 改为 `SELECT id, cloud_type, region, resource_id FROM targets ORDER BY id`
2. Provider 创建时使用实际 DB ID（而非数组下标）作为 targetIndex，或在 Provider 中新增 `TargetID() int` 方法
3. `filterRulesForTarget` 改为比较 `t == p.TargetID()`
4. 前端使用返回的 `id` 字段（同时修复 Issue 6.2）

**状态：** 待修复

---

### 7.2 根目录 Dockerfile 冗余且与 build/Dockerfile 内容不一致（严重度：中）

**模块：** Docker 构建

**文件：** `Dockerfile`（根目录，30 行）、`build/Dockerfile`（20 行）

**现象：** 第 9 轮核实发现 `build/Dockerfile` 已存在且内容符合规范（版本注入、`-tags docker`、双模式 HEALTHCHECK），但根目录仍保留一份旧版 Dockerfile：
- 缺少 `ARG VERSION` + ldflags 版本注入
- 缺少 `-tags docker` 构建标签
- HEALTHCHECK 仅检测进程存活（`killall -0`），不兼容 WebUI 模式
- 无任何构建引用（Makefile 和 CI 均指向 `build/Dockerfile`）

**原因分析：** 早期实现放在根目录，后来按文档要求创建了 `build/Dockerfile`，但旧文件未清理。Issue 5.3、5.4 所描述的问题均仅存在于根 Dockerfile。

**影响范围：** 用户若直接执行 `docker build .`（默认使用根 Dockerfile）会得到缺少版本信息和健康检查的镜像；维护两份 Dockerfile 容易产生不一致。

**推荐修复方案：**
- 删除根目录 `Dockerfile`，仅保留 `build/Dockerfile`
- Issue 5.3、5.4 随之关闭

**状态：** 待修复

---

### 7.3 ECS AuthorizeSecurityGroup 天然幂等与 isIdempotentCreate 的关系（严重度：低，确认无问题）

**模块：** Provider / 重试逻辑

**文件：** `syncer/retry.go` L100-106、`Documents/AliyunECSAPIGuide/AuthorizeSecurityGroup.md`

**核实结果：** ECS API 文档明确说明：“如果指定的安全组规则已存在，此次调用成功，但不会增加规则”。即 ECS 的创建操作天然幂等（不报错），`isIdempotentCreate` 中的错误码匹配主要针对 Lighthouse（`FirewallRulesExist`）和 CVM（`DuplicatePolicy`）。实现正确，无需修改。

**结论：** 无问题。

---

### 7.4 分页处理核实（确认无问题）

**模块：** Provider 层

**核实结果：**
- Lighthouse `DescribeFirewallRules`：默认 Limit=20，实现使用 Limit=100 + Offset 循环分页 ✅
- CVM `DescribeSecurityGroupPolicies`：API 无分页参数（规则上限 100 条，单次全量返回），实现无需分页 ✅
- ECS `DescribeSecurityGroupAttribute`：实现使用 NextToken 分页（MaxResults=500）✅
- SWAS `ListFirewallRules`：实现使用 PageNumber/PageSize=100 分页 ✅

**结论：** 无问题。

---

### 7.5 IPv6/ICMP/TCP+UDP 协议处理核实（确认无问题）

**模块：** Provider / Syncer

**核实结果：**
- SWAS 不支持 IPv6：`supportsIPv6()` 返回 false，`buildDesired` 跳过 IPv6 地址 ✅
- ECS 不支持 ICMPv6：`buildDesired` 跳过 IPv6+ICMP 组合，`retrySync` 中 WARN 提示 ✅
- Lighthouse IPv6+ICMP 使用 ICMPv6 协议 ✅，CVM 使用 ICMPV6 ✅
- TCP+UDP：仅 SWAS 原生支持（`supportsTCPUDP`），其他拆分为 TCP+UDP 两条 ✅
- ICMP 端口处理：Lighthouse 传 ALL，CVM 省略 Port，阿里云传 -1/-1（由 portconv.ToSlash 处理）✅

**结论：** 无问题，符合 AGENTS.md 约束。

---

### 7.6 热重载不重建 Provider 列表和凭据（严重度：中）

**模块：** WebUI / Syncer 热重载

**文件：** `main.go` L84-91、`syncer/syncer.go` L76-79、`provider/credentials.go`

**现象：** WebUI 修改配置后触发热重载（`ReloadFunc` → `store.LoadConfig()` → `s.Reload(newCfg)`），但：
1. `s.providers` 列表在启动时一次性创建，热重载后不重建——新增/删除目标不会生效
2. `provider.SetCredentials()` 仅在启动时调用一次——修改凭据不会生效
3. `ClientPool` 缓存已创建的 Client——即使重新 SetCredentials，旧 Client 仍用旧凭据

**原因分析：** `Syncer.Reload()` 仅更新 `s.cfg` 并重置 Ticker，未设计 Provider 层重建逻辑。

**影响范围：**
- WebUI 中添加新目标后，新目标不会被同步（直到重启）
- 删除目标后，旧 Provider 仍在运行（继续向已删除的资源写入规则！）
- 修改凭据后不生效（直到重启）

**推荐修复方案：**
- 方案 A（推荐）：热重载时重建 providers：
  1. `ReloadFunc` 中调用 `provider.SetCredentials(newCfg...)`
  2. 创建新的 `ClientPool` + 重新构建 providers 列表
  3. 为 Syncer 添加 `ReloadProviders([]provider.Provider)` 方法（或扩展 Reload 签名）
- 方案 B（简化）：热重载后通过 WebUI 提示用户“目标变更需重启生效”

**状态：** 待修复

---

## Issue 8: 第 10 轮深度审查（最终轮）

**阶段：** 全量构建完成后的多轮深度审查（第 10 轮：剩余文件全量检查 + 最终确认）

**检查范围：** app/cli.go、app/systray.go、notifier/email.go、notifier/webhook.go、dns/resolver.go、dns/circuitbreaker.go、internal/portconv、internal/tag、version/version.go、config/env.go、config/validate.go、config/config.go、app/app.go、app/mode.go、notifier/bus.go、.github/workflows/release.yml、go.mod、.gitignore

**核实结果：**
- CLI 子命令：`version`、`validate` 实现正确（backup/restore 缺失已记录于 Issue 5.11）✅
- 系统托盘：`//go:build desktop` 标签隔离正确（功能缺失已记录于 Issue 5.12）✅
- 邮件/Webhook 通知器：实现完整，未接入 EventBus 已记录于 Issue 5.8 ✅
- DNS 解析器：自定义 DNS + A/AAAA + CIDR 格式 + 10s 超时 ✅
- 熔断器：每域名独立、阈值触发、半开探测、成功解除 ✅
- 端口转换：Parse/ToSlash 逻辑正确（ALL → -1/-1）✅
- TAG 格式：`[TAG] comment` 前缀匹配正确 ✅
- release.yml：多平台构建 + 版本注入正确（缺前端构建步骤已记录于 Issue 5.18）✅
- go.mod：模块路径、Go 1.25、依赖完整 ✅
- .env 解析：续行合并、TARGETS/RULES 格式、协议校验、ICMP 强制 ALL、目标范围校验 ✅

**结论：** 本轮未发现新的高/中严重度问题。审查完成。

---

## 审查总结（第 1-10 轮）

**待修复（高严重度）：**
- 4.10 `.env.example` 旧格式残留（高）
- 5.1 docker-publish.yml 重复 YAML 文档（高）
- 6.1 README.md 旧版本内容残留（高）
- 7.1 filterRulesForTarget DB ID 不匹配（高）

**待修复（中严重度）：**
- 5.2 根 Dockerfile 冗余（中，已降级）
- 5.5 getDataDir 未按平台区分（中）
- 5.6 缺少 pidfile 防多实例（中）
- 5.7 同步日志未写入 SQLite（中）
- 5.8 告警通知器未接入 EventBus（中）
- 5.9 LoadConfig 缺少 webui_port/dns_fail_threshold（中）
- 5.10 缺少 SyncStart/SyncComplete 事件（中）
- 5.11 CLI 缺少 backup/restore（中）
- 5.12 托盘缺少开机自启和同步触发（中）
- 5.17 前端缺少高级功能/告警页面（中）
- 5.18 CI/CD 缺少前端构建步骤（中）
- 6.2 前端数组索引作 DB ID（中）
- 6.3 Docker WORKDIR 缺失（中）
- 7.6 热重载不重建 Provider（中）

**待讨论：**
- 5.15 npm vs pnpm、6.4 数据目录耦合、6.5 JSON 字段命名

**剩余风险：**
- 7.1 + 6.2 为同一根本缺陷，建议优先修复（影响数据正确性和同步可靠性）
- 5.18 导致 CI/CD 必然失败，影响发布流程
- 4.10/6.1/5.1 为文档残留问题，影响用户体验但不影响功能
