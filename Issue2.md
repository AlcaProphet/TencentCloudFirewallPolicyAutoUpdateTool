# FWAlizer WebUI 设计/功能问题记录

> 本文档记录 WebUI 层面的设计优化与功能缺陷。
> **第1-10轮审查中的 UI-01 至 UI-12 已全部修复并经第11-12轮确认，以上为精简结论。以下第11-12轮审查内容保持完整。**

---

## 已修复项精简汇总（第1-10轮）

### [UI-01] 云产品列显示原始标识符而非中文名称
- **严重度：** 低 | **状态：** ✅ 已修复
- **修复：** Targets.vue 为"云产品"列添加 `render` 函数，复用 `cloudOptions` 映射表。

### [UI-02] DNS 服务器默认值应改为国内公共 DNS 并自动补全端口
- **严重度：** 中 | **状态：** ✅ 已修复
- **修复：** `store.go`、`env.go` 默认 DNS 改为 `223.5.5.5`；`dns/resolver.go` 已有端口自动补全逻辑；前端 placeholder 和 `.env.example` 同步更新。

### [UI-03] 设置表单无默认值预填
- **严重度：** 中 | **状态：** ✅ 已修复
- **修复：** `GET /api/settings` 返回合并默认值后的有效配置（tag/interval/dns/log_level），前端 `v-model` 自动显示。

### [UI-04] 日志级别应改为下拉选择组件
- **严重度：** 低 | **状态：** ✅ 已修复
- **修复：** Settings.vue 的 log_level 从 NInput 改为 NSelect（debug/info/warn/error）。

### [UI-05] 时间戳格式不可读
- **严重度：** 中 | **状态：** ✅ 已修复
- **修复：** Logs.vue 添加 `formatTime()` 渲染函数，显示本地时区格式（含 UTC 偏移）。

### [UI-06] 历史记录中"目标"和"域名"字段始终为空
- **严重度：** 高 | **状态：** ✅ 已修复
- **修复：** syncer.go 在 `syncDomain()` 成功路径发布逐域名事件（含 provider/domain）；logwriter.go 过滤全局事件并为逐域名事件正确提取字段。

### [UI-07] 实时事件文案不可读
- **严重度：** 中 | **状态：** ✅ 已修复
- **修复：** Logs.vue 添加事件类型中文映射和智能详情渲染。

### [UI-08] 缺少实时日志流输出模块
- **严重度：** 中 | **状态：** ✅ 已修复
- **修复：** 新增 `webui/api/logstream.go`（SSE + slog 广播 Handler）+ 前端可折叠实时日志面板。

### [UI-09] 启动日志应显示完整访问地址
- **严重度：** 低 | **状态：** ✅ 已修复
- **修复：** server.go 启动日志改为 `"访问地址", "http://"+addr`。

### [UI-10] 全局设置 placeholder 冗余
- **严重度：** 低 | **状态：** ✅ 已修复
- **修复：** Settings.vue 移除 TAG/间隔/DNS 输入框的 placeholder 属性（值已由 API 预填）。

### [UI-11] 时间应显示本地时区而非固定 UTC
- **严重度：** 中 | **状态：** ✅ 已修复
- **修复：** `formatTime()` 使用 `Date` 本地时间 + `getTimezoneOffset()` 自动检测时区偏移，列标题改为"时间"。

### [UI-12] 同步日志目标字段应仅显示资源 ID
- **严重度：** 中 | **状态：** ✅ 已修复
- **修复：** logwriter.go 从 `provider.Name()` 格式（如 `tc_lighthouse(lhins-xxx)`）中提取括号内资源 ID。

---

# 第11轮深度审查（2026-07-26）：全项目逐文件合规检查

> 本轮基于 Design1.md、Build1.md、AGENTS.md 对项目全部纳入版本控制的文件进行逐文件审查，
> 同时复核 Issue1.md 中所有"待修复"和"已修复（待复核）"项的当前代码状态。

---

## 6. 构建与 Docker 配置

### [R11-01] `.dockerignore` 存在严重重复条目（[DOC-03] 修复未生效）

- **严重度：** 中
- **当前状态：** 待修复
- **所属模块：** Docker 构建
- **涉及文件：** `.dockerignore`（19行）
- **原始记录：** Issue1.md [DOC-03] 曾标记为"✅ 已修复（待复核）"，但代码中修复未实际生效

**现象描述：** 文件 19 行内容严重重复：
- L1-7（核心排除项）：`Documents/`、`*.md`、`.env`、`.git/`、`Dockerfile`、`.dockerignore`、`Makefile`
- L8-16（全部为 L1-7 的重复副本，顺序打乱）
- L12-13：额外的 `.env` 和 `.git`（无尾部斜杠）重复
- L17：额外的 `*.md` 重复
- L18：不存在的 `Ref/` 目录引用

共计 `.env` 出现 4 次、`*.md` 出现 3 次、`.git` 相关出现 3 次、`Dockerfile` 出现 2 次、`.dockerignore` 出现 2 次、`Makefile` 出现 2 次、`Documents/` 出现 2 次。

**原因分析：** Issue1.md [DOC-03] 描述的去重修复（精简为 7 行核心排除项）在代码中未生效。当前 19 行内容为原始未去重状态。

**影响范围：** 无功能影响，但增加维护负担和用户困惑。

**推荐修复方案：** 删除 L8-19，保留 L1-7 的核心排除项（已涵盖所有必要排除规则）。

---

## 7. 数据安全与事务

### [R11-02] `handleConfigImport` 事务未实际保护 Store 操作

- **严重度：** 高
- **当前状态：** 待修复
- **所属模块：** WebUI 后端 / 配置持久化
- **涉及文件：** `webui/api/settings.go`（L99-141）、`config/store.go`（L237-273）

**现象描述：** `handleConfigImport` 使用 `d.Store.WithTransaction(func(tx *sql.Tx) error { ... })` 包裹导入操作，但回调内部调用的 `ClearAll()`、`BatchAddTargets()`、`BatchAddRules()`、`SetSetting()` 全部使用 `s.db.Exec()`（直接操作 `*sql.DB`），而非事务对象 `tx.Exec()`。

**原因分析：** Go `database/sql` 中 `*sql.DB.Begin()` 返回的 `*sql.Tx` 会占用一个连接。`s.db.Exec()` 使用连接池中的**另一个**连接，因此不在同一事务中。结果：
- `ClearAll()` 立即提交清空操作（不可回滚）
- 若后续 `BatchAddTargets`/`BatchAddRules` 失败，事务回滚不会撤销 `ClearAll()`
- 用户原有配置丢失且新配置未写入 = **不可逆数据丢失**

**影响范围：** 配置导入中途失败时数据库处于"已清空但未导入完整"的不一致状态，原有配置永久丢失。

**推荐修复方案：**
- **方案 A（推荐）：** 让 `ClearAll`、`BatchAddTargets`、`BatchAddRules` 接受可选的 `*sql.Tx` 参数（事务版本），或新增 `ClearAllTx(tx *sql.Tx)` 等事务变体
- **方案 B：** 在 `handleConfigImport` 中内联所有 SQL 操作（绕过 Store 方法，直接使用 tx）

**详细修复步骤：**
```go
// store.go：新增事务版本的 ClearAll
func (s *Store) ClearAllTx(tx *sql.Tx) error {
    exec := s.db.Exec
    if tx != nil {
        exec = func(query string, args ...any) (sql.Result, error) {
            return tx.Exec(query, args...)
        }
    }
    _, err := exec("DELETE FROM targets; DELETE FROM rules; DELETE FROM settings;")
    return err
}
// 同理为 AddTarget / AddRule / SetSetting 增加 Tx 版本或使用 executor 接口
```

---

## 8. 前端数据一致性

### [R11-03] 前端 Targets.vue / Rules.vue 仍使用数组索引代替数据库 ID

- **严重度：** 高
- **当前状态：** 待修复
- **所属模块：** WebUI 前端
- **涉及文件：** `webui/frontend/src/views/Targets.vue`（L33, L52）、`webui/frontend/src/views/Rules.vue`（L37, L56）
- **原始记录：** Issue1.md [WEB-01]（三层联动修复方案已验证，但前端部分未实施）

**现象描述：** 
- `Targets.vue` L33：`openEdit(row, index)` → `editingId.value = index + 1`（使用数组下标）
- `Targets.vue` L52：`deleteTarget(index)` → `fetch('/api/targets/${index + 1}')`（使用数组下标）
- `Rules.vue` L37/L56：同样的 `index + 1` 模式

而后端 API `/api/targets/{id}` 期待的是数据库 ID（SQLite autoincrement 自增整数）。删除中间记录后，数组下标与 DB ID 不再对应，前端会操作错误的行。

**原因分析：** Issue1.md [WEB-01] 的四步修复方案中，第1步（结构体增加 ID 字段）、第2步（SELECT 含 id）、第3步（Provider 使用 DB ID）、第4步（Syncer 直接比较 DB ID）已全部实施。但**前端修改被遗漏**——`openEdit`/`deleteTarget` 仍然使用 `index + 1`。

**影响范围：** 删除中间记录后，编辑/删除操作会指向错误的数据库行；用户可能在不知情下修改/删除了错误的目标或规则。

**推荐修复方案：**
- `Targets.vue` L33：`editingId.value = row.id`
- `Targets.vue` L52：`fetch('/api/targets/${row.id}')`（需传入 row 而非 index）
- `Rules.vue` L37：`editingId.value = row.id`
- `Rules.vue` L56：`fetch('/api/rules/${row.id}')`（需传入 row 而非 index）

需同步调整 `openEdit`、`deleteTarget`、`deleteRule` 的函数签名从 `(row, index)` 改为 `(row)`（或仅传 `row.id`）。

---

## 9. CI/CD 构建链路

### [R11-04] CI/CD 流水线缺少前端构建步骤（[BLD-03]/[BLD-04] 确认仍存在）

- **严重度：** 高
- **当前状态：** 待修复
- **所属模块：** CI/CD
- **涉及文件：** `.github/workflows/docker-publish.yml`（L43 `go build -v ./...`）、`.github/workflows/release.yml`（L39 `go build -v ./...`）
- **原始记录：** Issue1.md [BLD-03]、[BLD-04]

**现象描述：** 两个 CI workflow 均包含 `go build -v ./...` 编译检查步骤，但缺少 Node.js 环境安装和 `npm ci && npm run build` 前端构建步骤。`webui/embed.go` 使用 `//go:embed frontend/dist`，而 `dist/` 目录被 `.gitignore` 排除，checkout 后不存在。`go build` 会因 embed 找不到文件而失败。

**原因分析：** [BLD-03] 和 [BLD-04] 的修复方案已在 Issue1.md 详细说明，但未实施。`docker-publish.yml` 的 Docker 构建步骤（使用 `build/Dockerfile`）可内部处理前端构建，但之前的 `go build ./...` 编译检查无法通过。`release.yml` 直接编译多平台二进制，完全无前端构建环节。

**影响范围：** CI/CD 流水线编译失败；Release 二进制不含 WebUI 前端。

**推荐修复方案：** 
在 Go 编译步骤前添加：
```yaml
- name: 设置 Node.js
  uses: actions/setup-node@v4
  with:
    node-version: '20'
- name: 构建前端
  run: cd webui/frontend && npm ci && npm run build
```

---

## 10. .env 模式 Provider ID 语义不一致

### [R11-05] `.env` 模式 `app.Run` 使用数组索引导致规则过滤不匹配

- **严重度：** 中
- **当前状态：** 待修复
- **所属模块：** App 生命周期 / Syncer
- **涉及文件：** `app/app.go`（L33）、`syncer/syncer.go`（L282-297）、`provider/registry.go`（L26）

**现象描述：** 
- `app.Run` L33：`provider.NewProvider(t, i, pool)`，`i` 为 0-based 数组下标（0, 1, 2...）
- `syncer.go` `filterRulesForTarget` L282-297：比较 `r.Targets` 中的值与 `targetDBID`
- `.env` 解析（`parseTargetNums`）校验 RULES targets 为 1-based（`n < 1 || n > max`）
- 结果：Provider 0 的 `TargetIndex()` 返回 0，Rule targets=[1,3] 永远不匹配 → 指定 target 的规则对第一个目标不生效

**示例：**
```
TARGETS=tc_lighthouse|lhins-abc|ap-guangzhou, tc_cvm|sg-def|ap-shanghai
RULES=api.example.com|TCP|443|ACCEPT|2|API
```
- Provider[0] targetIndex=0，Rule target=2 → 不匹配
- Provider[1] targetIndex=1，Rule target=2 → 不匹配（应为 Provider[1] 匹配 target 2）

**原因分析：** [WEB-01] 修复将 Syncer 改为直接比较 DB ID，但这要求 Provider 的 `TargetIndex()` 与 Rule 的 `Targets` 值语义一致。WebUI 模式两者均为 DB ID（一致），但 `.env` 模式 Provider 用 0-based 下标、Rule 用 1-based 编号（不一致）。

**影响范围：** `.env` 模式下，使用指定 targets 的规则（非 `*`/空）无法正确匹配到目标；只有空 targets（全部目标）的规则正常工作。

**推荐修复方案：**
- **方案 A（推荐）：** `app.Run` L33 改为 `provider.NewProvider(t, i+1, pool)`，使 `.env` 模式也使用 1-based ID，与 Rule targets 语义一致
- **方案 B：** 修改 RULES 解析将 targets 转为 0-based（需同时更新 `.env.example` 文档和校验范围）

---

## 11. 代码清理

### [R11-06] `app.Run` 的 `mode` 参数未被使用

- **严重度：** 低
- **当前状态：** 待修复
- **所属模块：** App 生命周期
- **涉及文件：** `app/app.go`（L16）
- **原始记录：** Issue1.md [DSC-01] "✅ 已裁定-移除mode参数"

**现象描述：** `func Run(cfg *config.Config, mode Mode) error` 接收 `mode` 参数，但函数体内（L17-51）从未引用 `mode`。自 Phase 2 后 WebUI 模式逻辑移到 `main.go` 内联，`app.Run` 仅服务 `.env` 模式。

**推荐修复方案：** 移除 `mode` 参数，签名改为 `Run(cfg *config.Config) error`；同步修改 `main.go` L125 调用为 `app.Run(cfg)`。

---

## 12. 文档与代码不一致

### [R11-07] README DNS 默认值与代码不一致

- **严重度：** 低
- **当前状态：** 待修复
- **所属模块：** 文档
- **涉及文件：** `README.md`（L91）

**现象描述：** README L91 表格中 DNS 默认值显示为 `8.8.8.8:53`，而代码实际默认值为 `223.5.5.5`（`config/env.go` L30、`config/store.go` L331）。

**推荐修复方案：** 将 README L91 的 DNS 默认值改为 `223.5.5.5`（并注明端口自动补全 :53）。

---

## 13. 功能缺失（已确认待规划项）

### [R11-08] CI/CD 和前端构建多个已知问题确认

以下 Issue1.md 中的已知问题经本轮逐文件核查，确认仍处于待修复/待规划状态：

| 编号 | 问题 | 当前状态 | 本轮确认 |
|------|------|---------|----------|
| [BLD-01] | 前端构建产物 `dist/` 缺失 | 待修复 | 确认：`dist/` 不在仓库中，需 CI/Makefile/Docker 构建生成 |
| [BLD-03] | CI/CD 缺少前端构建 | 待修复 | 见 [R11-04] |
| [BLD-04] | Release 缺少前端构建 | 待修复 | 见 [R11-04] |
| [WEB-06] | 前端缺少高级功能/告警页面 | 待规划 | 确认：`main.ts` 仅有 5 个路由，缺少 `/advanced`、`/alerts` |
| [FEA-03] | CLI 缺少 `backup`/`restore` | 待规划 | 确认：`cli.go` 仅有 `version` 和 `validate` |
| [FEA-06] | systray 缺少开机自启和同步触发 | 待规划 | 确认：`systray.go` L47 仍有 `// TODO` |
| [FEA-02] | 告警通知器未接入 EventBus | 待规划 | 确认：`email.go`/`webhook.go` 已实现但无注册代码 |

---

## 14. 本轮确认已修复的 Issue1.md 项

以下项目经逐文件核查，确认已在代码中正确实施：

| 编号 | 问题 | 验证结果 |
|------|------|----------|
| [DOC-01] | `.env.example` 旧格式残留 | **已修复**：59 行，格式干净，DNS=223.5.5.5 |
| [DOC-02] | `README.md` 旧版本残留 | **已修复**：352 行，无旧项目名，内容一致 |
| [DOC-04] | `firewall/` 空目录残留 | **已修复**：目录不存在 |
| [COR-01] | sync:start/complete 事件未发布 | **已修复**：`syncAll()` 和 `syncDomain()` 均发布事件 |
| [COR-02] | 熔断 IsOpen 未跳过同步 | **已修复**：`syncDomain` L228-231 有 `return` |
| [COR-03] | truncateDesc 缺失 SWAS 限制 | **已修复**：含 `CloudAliSWAS: maxLen=50` 分支 |
| [COR-04] | strVal 位置不当 | **已修复**：已移至 `common.go` |
| [COR-05] | 同步日志未写入 SQLite | **已修复**：`StoreLogWriter` 已订阅 EventBus |
| [COR-06] | LoadConfig 缺少配置项 | **已修复**：已读取 `webui_port` 和 `dns_fail_threshold` |
| [COR-07] | 热重载不重建 Provider | **已修复**：`ReloadFunc` 中重建 providers 和凭据 |
| [FEA-01] | getDataDir 未按平台区分 | **已修复**：支持 `FWALIZER_DATA_DIR` + 平台路径 |
| [FEA-07] | WebUI 缺少凭据配置 | **已修复**：Settings.vue 有凭据输入框 |
| [DKR-01] | 根 Dockerfile 冗余 | **已修复**：根 Dockerfile 已删除 |
| [DKR-02] | Dockerfile 缺少前端构建 | **已修复**：含 `frontend-builder` 阶段 |
| [DKR-03] | 缺少 WORKDIR | **已修复**：`WORKDIR /app` |
| [BLD-05] | Makefile 缺少 frontend | **已修复**：`build` 依赖 `frontend` 目标 |
| [BLD-02] | docker-publish.yml 重复 | **已修复**：85 行单文档，引用 `build/Dockerfile` |
| [WEB-08] | 三个页面空白 | **已修复**：字段名 + NMessageProvider + nil slice |
| [DSC-05] | 前端包管理器 | **已修复**：使用 npm，文档待同步（Design1.md/Build1.md 仍写 pnpm） |
| [DSC-06] | Docker 数据目录耦合 | **已修复**：`FWALIZER_DATA_DIR` + `docker-compose.yml.example` |

---

## 审查总结

### 本轮（第11轮）新发现问题

| 严重度 | 数量 | 编号 |
|--------|------|------|
| 高 | 3 | [R11-02]、[R11-03]、[R11-04] |
| 中 | 2 | [R11-01]、[R11-05] |
| 低 | 2 | [R11-06]、[R11-07] |
| **合计** | **7** | |

### 剩余风险

1. **数据安全**（最高风险）：[R11-02] 配置导入事务无效 — 用户执行导入失败后永久丢失配置
2. **数据一致性**（高风险）：[R11-03] 前端仍用数组索引 — 删除中间记录后操作错误行
3. **构建链路断裂**（高风险）：[R11-04] CI/CD 无法编译 — 发布产物不含 WebUI
4. **功能正确性**（中风险）：[R11-05] .env 模式规则过滤错误 — 指定 targets 的规则不生效
5. **配置卫生**（低风险）：[R11-01] .dockerignore 重复、[R11-06] 未使用参数、[R11-07] README 文档不一致

### 终止判定

本轮发现 **3 个高严重度 + 2 个中严重度** 新问题，**不满足连续一轮无高/中严重度新问题的终止条件**。建议在修复本轮问题后继续第12轮审查。

---

> **审查说明**：本轮（第11轮）为全项目逐文件深度审查，覆盖 Go 后端、前端源码、配置文件、构建脚本、CI/CD、Docker、文档共 50+ 文件。与 Issue1.md 记载的第1-10轮审查衔接（第10轮终止后新增）。

---

# 第12轮审查（2026-07-26）：深度边缘验证与合规扫描

> 本轮聚焦于测试文件、Documents 目录、Provider 边界条件、错误处理完整性、安全配置合规。
> 与第11轮形成互补，第11轮为广度扫描，本轮为深度验证。

---

## 15. 本轮验证通过的合规项

### 15.1 测试覆盖

- 已有 7 个测试文件覆盖 `config`、`dns`、`internal/portconv`、`internal/tag`、`notifier`、`provider/common` 六个包
- 全部测试通过：`go test ./...` 返回 `ok`，无失败
- `common_test.go` 覆盖 OwnedRules、Diff 新增/不变/删除/域名隔离、TCP+UDP 拆分、SWAS IPv6 跳过、ECS ICMPv6 跳过、ClientPool 复用 — 共 8 个测试用例
- `circuitbreaker_test.go` 覆盖熔断触发、重置、半开探测不计入 — 共 3 个测试用例
- `env_test.go` 覆盖正常解析、续行、默认值、非法协议/动作/编号/Provider、空内容、ICMP 强制 ALL、凭据校验、空规则 — 共 10 个测试用例

### 15.2 API 合规性

| 检查项 | 结果 | 证据 |
|--------|------|------|
| 全量覆盖类 API | **合规** — 零使用 | `grep ModifyFirewall\|ResetFirewall\|ReplaceFirewall` 在 provider/ 中零匹配 |
| 仅操作入站规则 | **合规** | CVM 仅用 `Ingress`；ECS 仅 `Direction=ingress`；Lighthouse/SWAS 防火墙无方向概念 |
| Egress 操作 | **合规** — 零使用 | CVM `CreateRules`/`DeleteRules` 仅设置 `Ingress` 字段 |

### 15.3 安全合规

| 检查项 | 结果 | 证据 |
|--------|------|------|
| WebUI 绑定地址 | **合规** — `127.0.0.1` | `server.go` L51: `127.0.0.1:%d` |
| 凭据不导出 | **合规** | `settings.go` L75-78: `delete(settings, "tc_access_id")` 等 4 个 key |
| 凭据不导入 | **合规** | `settings.go` L125-126: 跳过四个凭据 key |
| 配置导出无凭据 | **合规** | 同凭据不导出 |
| 无 panic | **合规** | `grep panic` 在 webui/ 中零匹配 |

### 15.4 日志规范

| 检查项 | 结果 |
|--------|------|
| syncer/ 包 | **合规** — 全部使用 `log/slog`，无 `fmt.Print` |
| provider/ 包 | **合规** — 全部使用 `log/slog`，无 `fmt.Print` |
| dns/ 包 | **合规** — 无日志输出（纯函数） |
| webui/api/ | **合规** — 3 处 slog（2 Warn + 1 Info），无 `fmt.Print` |
| main.go | **合规** — 启动前错误用 `fmt.Fprintf(os.Stderr)`（slog 未初始化），运行时用 slog |

### 15.5 编译与静态检查

- `go vet ./...` — **通过**，零警告
- `go build ./...` — **通过**，零错误
- `go test ./...` — **通过**，所有 6 个测试包 `ok`

### 15.6 Documents 目录

- 4 个子目录（AliyunECSAPIGuide、AliyunSASAPIGuide、TencentCVMAPIGuide、TencentLighthouseAPIGuide）
- 包含各云厂商 API 参考文档（中文 + 英文），与编码实现中的 API 使用一致
- `.DS_Store` 文件存在（macOS 系统文件），建议通过 .gitignore 排除

---

## 16. 本轮确认仍需修复的已知问题

| 编号 | 问题 | 严重度 | 状态 |
|------|------|--------|------|
| [R11-02] | 配置导入事务未保护 Store 操作 | 高 | 待修复 |
| [R11-03] | 前端仍用数组索引代替 DB ID | 高 | 待修复 |
| [R11-04] | CI/CD 缺少前端构建步骤 | 高 | 待修复 |
| [R11-01] | .dockerignore 严重重复 | 中 | 待修复 |
| [R11-05] | .env 模式规则过滤不匹配 | 中 | 待修复 |
| [R11-06] | app.Run mode 参数未使用 | 低 | 待修复 |
| [R11-07] | README DNS 默认值不一致 | 低 | 待修复 |
| [WEB-06] | 前端缺少 /advanced、/alerts 页面 | 中 | 待规划 |
| [FEA-06] | systray 缺少同步触发和开机自启 | 低 | 待规划 |
| [FEA-03] | CLI 缺少 backup/restore | 中 | 待规划 |

---

## 17. 审查终止判定

### 第11轮：发现 3 高 + 2 中 → 不满足终止条件
### 第12轮：发现 0 高 + 0 中 → **满足终止条件**

第12轮深度验证未发现新的高/中严重度问题，所有测试通过，go vet 零警告。第11轮发现的 5 个高/中问题已全部记录并附修复方案。

**审查结论：** 项目在核心逻辑层（Provider 抽象、Syncer 引擎、DNS 解析、熔断、EventBus、API 端点）实现质量高，与 Build1.md/Design1.md 规范一致性良好。主要剩余风险集中在构建链路（前端构建未自动化）和个别数据安全/一致性缺陷（事务无效、前端索引错位）。

**建议：** 优先修复 [R11-02]（数据安全）和 [R11-03]（数据一致性），再统一处理构建链路问题（[R11-04] + [BLD-03]/[BLD-04]）。
