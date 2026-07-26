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

**原因分析：** Issue 4.10 曾标记为"已修复"，但修复未实际生效或后续操作覆盖了修复结果。第 1 轮、第 8 轮审查均确认问题仍存在。

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
1. `build` 目标仅编译 Go 源码，不构建前端，生成的二进制不含 WebUI
2. `docker-build` 目标需确认引用正确的 Dockerfile 路径

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

**推荐修复方案：** 删除根目录 `Dockerfile`，仅保留 `build/Dockerfile`。原 Issues 5.3、5.4（版本注入和 HEALTHCHECK）随之关闭。

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

**原因分析：** `Store.GetTargets()` 和 `Store.GetRules()` 的 SELECT 查询未返回 `id` 列，导致整个系统（前端、Syncer）只能用数组位置"猜测" DB ID。初始状态下两者恰好一致，删除操作后产生偏移。

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

**现象描述：** `handleConfigImport` 依次执行 `ClearAll()` → `BatchAddTargets()` → `BatchAddRules()` → settings 写入，这些操作不在同一事务中。若中间步骤失败，数据库处于"已清空但未导入完整"的不一致状态。

**影响范围：** 配置导入中途失败时，用户原有配置丢失且新配置未完全写入。

**推荐修复方案：** 在 `config/store.go` 添加 `WithTransaction` 事务方法，将导入操作包裹在同一事务中，失败时自动回滚。

---

### [WEB-03] TypeScript `any` 类型泛滥

- **严重度：** 低
- **当前状态：** 📋 待规划（低优先级）
- **所属模块：** WebUI 前端
- **涉及文件：** 所有 `.vue` 文件中的 `<script setup lang="ts">` 块
- **原始记录：** Issue 9.3（第 5 轮）

**现象描述：** 所有组件广泛使用 `any` 类型（`ref<any>(...)`、`render(row: any)`），虽然 `tsconfig.json` 中 `strict: true` 已启用，但 `any` 绕过了所有类型检查。

**影响范围：** IDE 无法提供自动补全和编译期类型检查；重构时容易引入 bug。

**推荐修复方案：** 在 `src/types.ts` 中定义 `Target`、`Rule`、`SyncStatus` 等接口，逐步替换 `any` 为具体类型。

**最终裁定：** 标记为低优先级待规划，暂不纳入当前修复周期。当前功能运行正常，类型体操属于代码质量优化而非功能缺陷；待核心功能稳定后再逐步完善。

---

### [WEB-04] Dashboard 使用轮询而非 SSE 获取状态

- **严重度：** 低
- **当前状态：** 📋 待规划
- **所属模块：** WebUI 前端
- **涉及文件：** `webui/frontend/src/views/Dashboard.vue`
- **原始记录：** Issue 9.4（第 5 轮）

**现象描述：** Dashboard 使用 `setInterval(fetchStatus, 5000)` 每 5 秒轮询，而 Logs 页面已正确使用 SSE。导致状态更新有 5 秒延迟，且产生不必要的 HTTP 请求。

**推荐修复方案：** Dashboard 同时使用 SSE 监听 `sync:start`/`sync:complete` 事件来更新状态；保留轮询作为 SSE 断连时的 fallback（延长间隔到 30s）。

**最终裁定：** 标记为待规划，暂不纳入当前修复周期。理由：5 秒轮询对内部工具场景影响极小（每秒仅 0.2 次请求），且 SSE 改造依赖 [COR-01] 事件发布修复先完成。待 sync 事件体系完善后再统一改造。

---

### [WEB-05] WebUI 模式缺少 pidfile 防多实例机制

- **严重度：** 低
- **当前状态：** 📋 待规划
- **所属模块：** App 生命周期
- **涉及文件：** `main.go`（WebUI 模式启动逻辑）
- **原始记录：** Issue 8.2（第 4 轮）/ Issue 5.6（第 7 轮复查）

**现象描述：** Build1.md 12.12 节规定 WebUI 模式启动时应创建 pidfile 防止多实例运行，当前实现完全没有 pidfile 逻辑。

**影响范围：** 用户可能误启动多个 WebUI 实例，导致 SQLite 写冲突和端口占用。

**推荐修复方案：** 启动时在数据目录创建 `fwalizer.pid`，写入当前 PID；启动前检测已有 pidfile 的进程是否存活；正常退出时删除 pidfile。

**最终裁定：** 标记为待规划，暂不纳入当前修复周期。理由：多实例启动场景极少（用户通常通过 systemd/Docker 管理服务），误操作概率低；实现方案明确，可作为 v1.1 体验优化项。

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
- **当前状态：** ✅ 已裁定-保持现状
- **所属模块：** WebUI 前端 + API
- **涉及文件：** `webui/frontend/src/views/Targets.vue`、`Rules.vue`
- **原始记录：** Issue 6.5（第 8 轮）

**现象描述：** 前端表单使用 Go 结构体字段名（PascalCase，如 `CloudType`、`ResourceID`）作为 JSON key，而部分 API 使用 snake_case（如 `testConnectionReq` 的 `json:"cloud_type"` tag）。Go `encoding/json` 对无 tag 字段做大小写不敏感匹配，当前能正常工作，但 API 风格不统一。

**推荐修复方案：**
- 方案 A：统一为 snake_case + json tag（规范）
- 方案 B（推荐）：保持现状（内部工具，功能正常即可）

**最终裁定：** 选方案 B。保持现状不做修改。理由：Go `encoding/json` 大小写不敏感匹配确保功能正常；统一风格需同时改动前后端，风险高于收益；内部工具不过度设计。

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
- **当前状态：** 📋 待规划（低优先级）
- **所属模块：** CVM Provider
- **涉及文件：** `provider/tc_cvm.go`
- **原始记录：** Issue 11.2（第 7 轮）

**现象描述：** `checkRuleLimit` 优先使用 `PolicyStatistics` 精确统计（含 IPv4/IPv6），但 fallback 到手动计数时仅使用 `len(ps.Ingress) + len(ps.Egress)`，未包含 `Ipv6Ingress`/`Ipv6Egress`。

**影响范围：** 仅当 API 未返回 `PolicyStatistics` 且存在 IPv6 规则时，规则计数不准确。此情况极少见。

**推荐修复方案：** fallback 路径补充 `len(ps.Ipv6Ingress) + len(ps.Ipv6Egress)`。

**最终裁定：** 标记为低优先级待规划，暂不修复。理由：触发条件极为罕见（需同时满足 API 未返回 PolicyStatistics 且存在 IPv6 规则），当前无已知用户触发此路径；修复方案明确（一行代码），可在下次触及该文件时顺手修复。

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

### [FEA-04] 项目 `README.md` 需清理旧内容并补充

- **严重度：** 中
- **当前状态：** 待规划
- **所属模块：** 文档
- **涉及文件：** `README.md`
- **原始记录：** Issue 12.3（第 8 轮）

**现象描述：** 仓库中存在 `README.md`，但包含旧版本残留（见 [DOC-02]），且缺少面向用户的纯净项目介绍和快速开始指南。

**推荐修复方案：** 清理旧内容后（[DOC-02]），补充项目简介、功能特性、快速开始（Docker / 二进制 / 源码）、配置说明、运行模式等。

---

### [FEA-05] 测试覆盖缺口

- **严重度：** 低
- **当前状态：** 📋 待规划
- **所属模块：** 测试
- **涉及文件：** 多个（覆盖缺口）
- **原始记录：** Issue 12.5（第 8 轮）

**现象描述：** 已有测试覆盖 `config`、`dns`、`internal/portconv`、`internal/tag`、`notifier`、`provider/common` 六个包。缺失覆盖：四个 Provider 实现文件（`tc_lighthouse.go` / `tc_cvm.go` / `ali_swas.go` / `ali_ecs.go`）、`syncer/`、`webui/api/`、`config/store.go`。

**推荐修复方案：** 优先对 `provider/common.go` 的 Diff 逻辑补充边界用例；Provider 实现层测试依赖云 SDK Mock，可留待后期；Syncer 层可通过注入 Mock Provider + Mock Resolver 进行单元测试。

**最终裁定：** 标记为待规划，暂不纳入当前修复周期。理由：已有测试覆盖了核心逻辑包（Diff、熔断、端口转换、事件总线）；Provider/Syncer 层测试依赖 Mock 基础设施（需先建立 Mock SDK Client 体系），投入产出比在 MVP 阶段不高；可在 v1.1 建立 Mock 体系后逐步补充。

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
- **当前状态：** ✅ 已裁定-移除mode参数
- **所属模块：** App 生命周期
- **涉及文件：** `app/app.go`
- **原始记录：** Issue 5.3（第 1 轮）

**现象描述：** `func Run(cfg *config.Config, mode Mode) error` 接收 `mode` 参数但从未使用。原因：Phase 2 后 WebUI 模式逻辑移到 `main.go` 内联，`app.Run` 退化为仅服务于 `.env` 模式的函数。

**推荐修复方案：**
- 方案 A（推荐）：移除 `mode` 参数，精简为 `Run(cfg *config.Config) error`
- 方案 B：保留参数但在函数内根据 mode 做差异化行为

**最终裁定：** 选方案 A。移除 `mode` 参数，精简函数签名。理由：参数自 Phase 2 后从未使用，保留死代码违反"功能做减法"原则；若未来需要差异化行为再加回参数不影响兼容性（仅一处调用）。

---

### [DSC-02] CVM `checkRuleLimit` 存在重复 API 调用

- **严重度：** 低
- **当前状态：** ✅ 已裁定-保持现状
- **所属模块：** CVM Provider
- **涉及文件：** `provider/tc_cvm.go`
- **原始记录：** Issue 6.2（第 2 轮）

**现象描述：** `CreateRules` 调用 `checkRuleLimit`，后者独立调用 `DescribeSecurityGroupPolicies`。但 Syncer 的 `retrySync` 在调用 `CreateRules` 前已通过 `GetRules()` 获取了所有规则。`checkRuleLimit` 再次调用相同 API 造成额外网络开销。

**推荐修复方案：**
- 方案 A：将规则计数检查上提到 `retrySync` 层，复用 `GetRules()` 返回的结果
- 方案 B（推荐）：保持现状，额外 API 调用作为安全冗余

**最终裁定：** 选方案 B。保持现状不做修改。理由：CVM 查询 API 配额充足（100 次/秒），额外一次调用几乎无影响；方案 A 会模糊 Provider 自治边界（Provider 应自己保证不超过安全组上限），保持现状更简洁且防御性更强。

---

### [DSC-03] `testConnection` 是否复用应用级 `ClientPool`

- **严重度：** 低
- **当前状态：** ✅ 已裁定-注入复用ClientPool
- **所属模块：** WebUI API
- **涉及文件：** `webui/api/targets.go`
- **原始记录：** Issue 11.3（第 7 轮）

**现象描述：** `handleTestConnection` 创建全新的 `provider.NewClientPool()`，而非复用应用级 pool。测试连接的 SDK client 与同步引擎使用的 client 不是同一个实例。

**推荐修复方案：**
- 方案 A（推荐）：将 `ClientPool` 注入 `Deps` 中（或通过 `Syncer` 接口暴露）以复用
- 方案 B：保持现状（差异极小，不值得增加耦合）

**最终裁定：** 选方案 A。将应用级 `ClientPool` 注入 `Deps` 中统一复用。理由：确保测试连接与同步引擎使用完全一致的 SDK client 配置，消除"测试通过但同步失败"的潜在排查盲区；注入方式增加的耦合度可控。

---

### [DSC-04] HTTP server 优雅退出必要性

- **严重度：** 低
- **当前状态：** ✅ 已裁定-保持现状
- **所属模块：** App 生命周期
- **涉及文件：** `main.go`
- **原始记录：** Issue 10.7（第 6 轮）

**现象描述：** WebUI 模式下 HTTP server goroutine 没有显式的 `Shutdown()` 调用，随进程退出被强制终止。优雅退出仅覆盖了 Syncer（完成当前轮次），HTTP server 无等价处理。

**推荐修复方案：**
- 方案 A：使用 `context.WithCancel` + `http.Server.Shutdown()` 实现优雅退出
- 方案 B（推荐）：保持现状（内部工具场景下可接受）

**最终裁定：** 选方案 B。保持现状不做修改。理由：项目为内部工具，WebUI 仅有少量管理请求，不存在长时间运行的 HTTP 连接；实现优雅退出需额外处理 context 传递和 shutdown 超时，增加复杂度但收益极小，不符合"简单轻量化"原则。

---

### [DSC-05] 前端包管理器：npm vs pnpm

- **严重度：** 低
- **当前状态：** ✅ 已裁定-选npm，文档同步更新
- **所属模块：** WebUI 前端
- **涉及文件：** `webui/frontend/package.json`、`package-lock.json`、Design1.md、Build1.md
- **原始记录：** Issue 5.15（第 7 轮复查）

**现象描述：** Design1.md 和 Build1.md 规定前端包管理器为 pnpm，但项目中使用 `package-lock.json`（npm 产物）而非 `pnpm-lock.yaml`。

**推荐修复方案：**
- 方案 A：改用 pnpm（符合文档）
- 方案 B（推荐）：文档改为 npm（降低门槛，已有 `package-lock.json`）

**最终裁定：** 选方案 B。保留 npm，将 Design1.md 和 Build1.md 中包管理器要求从 pnpm 改为 npm。理由：npm 为 Node.js 内置工具，零门槛；项目中 `package-lock.json` 已存在，CI/Makefile/Dockerfile 构建命令无需改动；项目为内部工具，npm 的依赖管理能力已足够。配套操作：同步更新设计文档中的包管理器描述。

---

### [DSC-06] Docker 数据目录路径与 `getDataDir()` 的潜在耦合

- **严重度：** 低
- **当前状态：** ✅ 已裁定-添加FWALIZER_DATA_DIR环境变量
- **所属模块：** Docker 部署
- **涉及文件：** `README.md`、`main.go`、`docker-compose.yml.example`（新增）
- **原始记录：** Issue 6.4（第 8 轮）

**现象描述：** README WebUI 模式示例挂载卷到 `/home/appuser/.config/fwalizer`。当前 `getDataDir()` 使用 `os.UserHomeDir()` + `/.config/fwalizer`，路径碰巧一致。但若修复 [FEA-01]（按平台区分路径）后，需确保 Docker 内的路径仍一致。

**推荐修复方案：**
- 方案 A（推荐）：添加 `FWALIZER_DATA_DIR` 环境变量支持，Docker 通过该变量显式指定路径
- 方案 B：保持现状，在 README 中说明路径约定

**最终裁定：** 选方案 A。在 `main.go` 的 `getDataDir()` 中优先读取 `FWALIZER_DATA_DIR` 环境变量；新增 `docker-compose.yml.example` 文件演示用法。理由：这是 Docker 最佳实践（通过环境变量配置路径），同时解决 [FEA-01] 跨平台路径与 Docker 的耦合问题；方案 B 在修复 [FEA-01] 后会 break。配套操作：创建 `docker-compose.yml.example` 文件，在其中体现 `FWALIZER_DATA_DIR` 用法。

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

**结论：** 第 4 轮审查确认所有 6 个缺失端点已全部补齐（含 SSE 实时推送），API 层功能完整。文件拆分至 `webui/api/` 子目录作为可选优化项保留。

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

- **错误处理合规：** 所有 error 均已处理；全项目统一使用 `log/slog`；注释使用中文；未使用全量覆盖 API；仅操作入站规则
- **安全合规：** WebUI 绑定 `127.0.0.1`；凭据通过独立环境变量传入；配置导出自动剔除凭据
- **核心逻辑合规：** 频率限制符合文档；幂等处理正确（删除"已不存在"视为成功，添加"已存在"WARN 跳过）；重试语义正确（每次重试重新 Describe → Diff → Create/Delete）；TCP+UDP 拆分正确；IPv6+ICMP 处理正确
- **SPA 路由：** 前端使用 `createWebHashHistory()`（hash 路由），与 `http.FileServer` 完全兼容，无需 fallback
- **ECS 幂等性：** ECS `AuthorizeSecurityGroup` 天然幂等（"规则已存在时调用成功但不增加规则"），`isIdempotentCreate` 的错误码匹配主要针对 Lighthouse/CVM，实现正确
- **四云分页：** Lighthouse（Limit=100 + Offset）、CVM（无分页参数，单次全量）、ECS（NextToken，MaxResults=500）、SWAS（PageNumber/PageSize=100），四家分页处理均正确
- **协议处理：** SWAS 跳过 IPv6；ECS 跳过 ICMPv6；Lighthouse/CVM ICMP 协议名正确；TCP+UDP 拆分正确；ICMP 端口处理正确
- **第 10 轮全面检查：** CLI 子命令、系统托盘、邮件/Webhook 通知器、DNS 解析器、熔断器、端口转换、TAG 格式、release.yml、go.mod、.env 解析等均通过核查，无新发现的高/中严重度问题

---

# 三、汇总表

## 按严重度统计

| 严重度 | 待修复 | 已裁定（待实施） | 已裁定（无需改动） | 待规划 | 已关闭 |
|--------|--------|-----------------|-------------------|--------|--------|
| 高 | 8 | — | — | — | 4 |
| 中 | 10 | — | — | 5 | 2 |
| 低 | 3 | 4 | 3 | 6 | 4 |
| **合计** | **21** | **4** | **3** | **11** | **10** |

> 注：「已裁定（待实施）」指方案已确定但仍需编码/文档变更的项目；「已裁定（无需改动）」指裁定为保持现状的项目。

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
| [FEA-04] | 项目 README 需清理并补充 | 待规划 |

### 低严重度（16 项）

| 编号 | 问题 | 状态 |
|------|------|------|
| [DOC-03] | `.dockerignore` 重复条目 | 待修复 |
| [DOC-04] | `firewall/` 空目录残留 | 待修复 |
| [COR-04] | `strVal` 工具函数位置不当 | 待修复 |
| [DSC-01] | `app.Run` mode 参数未被使用 | ✅ 已裁定-移除mode参数 |
| [DSC-05] | 前端包管理器 npm vs pnpm | ✅ 已裁定-选npm，文档同步更新 |
| [DSC-06] | Docker 数据目录路径耦合 | ✅ 已裁定-添加FWALIZER_DATA_DIR环境变量 |
| [DSC-03] | `testConnection` 复用 ClientPool | ✅ 已裁定-注入复用ClientPool |
| [DSC-02] | CVM `checkRuleLimit` 重复 API 调用 | ✅ 已裁定-保持现状 |
| [DSC-04] | HTTP server 优雅退出必要性 | ✅ 已裁定-保持现状 |
| [WEB-07] | 前端 JSON 字段命名风格不统一 | ✅ 已裁定-保持现状 |
| [COR-08] | CVM `checkRuleLimit` IPv6 计数不完整 | 📋 待规划（低优先级） |
| [WEB-03] | TypeScript `any` 类型泛滥 | 📋 待规划（低优先级） |
| [WEB-04] | Dashboard 轮询而非 SSE | 📋 待规划 |
| [WEB-05] | pidfile 防多实例缺失 | 📋 待规划 |
| [FEA-05] | 测试覆盖缺口 | 📋 待规划 |
| [FEA-06] | systray 缺少开机自启和同步触发 | 📋 待规划 |

> 注：原「待讨论」分类中的 7 项（[DSC-01]~[DSC-06]、[WEB-07]）已全部完成裁定，并入低严重度清单。

---

## 已裁定事项汇总

以下 7 项已完成最终裁定，**无需再次讨论**：

| 编号 | 裁定结论 | 后续动作 |
|------|---------|---------|
| [DSC-01] | 移除 `mode` 参数 | 修改 `app/app.go` 函数签名 |
| [DSC-02] | 保持现状 | 无 |
| [DSC-03] | 注入复用 ClientPool | 修改 `webui/api/targets.go`，ClientPool 注入 Deps |
| [DSC-04] | 保持现状 | 无 |
| [DSC-05] | 选 npm | 更新 Design1.md、Build1.md 中包管理器描述 |
| [DSC-06] | 添加 `FWALIZER_DATA_DIR` 环境变量 | 修改 `main.go` 的 `getDataDir()`；创建 `docker-compose.yml.example` |
| [WEB-07] | 保持现状 | 无 |

---

## 核心发现与优先修复建议

1. **构建链路断裂**（最高优先级）：前端无法被打包进二进制是最严重的问题，影响所有分发渠道（Docker、Release、本地编译）。建议优先修复 [BLD-01] + [BLD-03] + [BLD-04] + [DKR-02] + [BLD-05]，可统一在一次变更中完成。前端构建命令统一使用 `npm ci && npm run build`（见 [DSC-05] 裁定）。

2. **数据一致性风险**（高优先级）：[WEB-01] 前端数组索引 vs DB ID 错位问题影响数据正确性和同步可靠性，删除记录后规则可能静默停止同步，存在安全风险。建议紧随构建链路修复后处理。

3. **文档残留清理**（中优先级）：[DOC-01]、[DOC-02] 两处文档残留直接影响新用户体验，可批量清理。同步更新 Design1.md、Build1.md 中包管理器为 npm（[DSC-05] 裁定）。

4. **Docker 规范清理**（中优先级）：[DKR-01] 删除根 Dockerfile + [DKR-03] 添加 WORKDIR + [DSC-06] 创建 docker-compose.yml.example，清理 Docker 配置冗余并建立环境变量驱动的数据目录配置。

5. **核心功能补全**：熔断器实际生效（[COR-02]）、同步事件发布（[COR-01]）、热重载重建（[COR-07]）对系统可靠性有直接影响，建议在功能补全阶段集中处理。同步日志写入（[COR-05]）待裁定后跟进。

6. **已裁定待实施项**（低优先级，随上述批次顺手处理）：移除 `app.Run` mode 参数（[DSC-01]）、ClientPool 注入 Deps（[DSC-03]）、getDataDir 读 FWALIZER_DATA_DIR（[DSC-06]）。

---

> **审查终止说明**：第 9 轮审查未发现高/中严重度新问题，满足终止条件。第 10 轮最终检查确认无遗漏。审查共进行 10 轮，在第 9 轮后自然终止。
>
> **原始审查编排**：Issue 1–4（构建期 + 首轮修复）→ Issue 5–11（第 1–7 轮深度审查）→ Issue 12（第 8 轮交叉验证）→ 第 9 轮（深度复查）→ 第 7–10 轮复查和扩展。多段审查中存在大量交叉引用和重复记录，本文档已完成合并去重。