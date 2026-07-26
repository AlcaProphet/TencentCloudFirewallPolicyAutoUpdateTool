# FWAlizer 问题追踪

> 第11-12轮逐文件审查（2026-07-26），以 Design1.md、Build1.md、AGENTS.md 为基准。
> 第1-10轮已修复项及独有待规划项见 [Issue1.md](./Issue1.md)。

---

## 当前待修复项

| 编号 | 问题 | 严重度 | 状态 |
|------|------|--------|------|
| [R11-02](#r11-02-handleconfigimport-事务未实际保护-store-操作) | 配置导入事务未保护 Store 操作 | 🔴 高 | 待修复 |
| [R11-03](#r11-03-前端-targetsvue--rulesvue-仍使用数组索引代替数据库-id) | 前端仍用数组索引代替 DB ID | 🔴 高 | 待修复 |
| [R11-04](#r11-04-cicd-流水线缺少前端构建步骤bld-03bld-04-确认仍存在) | CI/CD 缺少前端构建步骤 | 🔴 高 | 待修复 |
| [R11-01](#r11-01-dockerignore-存在严重重复条目doc-03-修复未生效) | .dockerignore 严重重复 | 🟡 中 | 待修复 |
| [R11-05](#r11-05-env-模式-apprun-使用数组索引导致规则过滤不匹配) | .env 模式规则过滤不匹配 | 🟡 中 | 待修复 |
| [R11-06](#r11-06-apprun-的-mode-参数未被使用) | app.Run mode 参数未使用 | ⚪ 低 | 待修复 |
| [R11-07](#r11-07-readme-dns-默认值与代码不一致) | README DNS 默认值不一致 | ⚪ 低 | 待修复 |
| [WEB-06] | 前端缺少 /advanced、/alerts 页面 | 🟡 中 | 待规划 |
| [FEA-02] | 告警通知器未接入 EventBus | 🟡 中 | 待规划 |
| [FEA-03] | CLI 缺少 backup/restore | 🟡 中 | 待规划 |
| [FEA-06] | systray 缺少同步触发和开机自启 | ⚪ 低 | 待规划 |

---

## 第11轮新发现问题

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

**实施计划**（已验证代码基线一致）：
- 文件：`.dockerignore`，当前 19 行
- 操作：删除 L8-19（含重复条目和不存在目录 `Ref/`），保留 L1-7
- 验证：`cat .dockerignore` 仅含 7 行核心排除项，无重复

---

### [R11-02] `handleConfigImport` 事务未实际保护 Store 操作

- **严重度：** 高
- **当前状态：** 待修复
- **所属模块：** WebUI 后端 / 配置持久化
- **涉及文件：** `webui/api/settings.go`（L99-141）、`config/store.go`（L237-273）
- **原始记录：** Issue1.md [WEB-02]（第4轮首次发现配置导入缺少事务保护，第11轮确认 WithTransaction 包装未实际生效）

**现象描述：** `handleConfigImport` 使用 `d.Store.WithTransaction(func(tx *sql.Tx) error { ... })` 包裹导入操作，但回调内部调用的 `ClearAll()`、`BatchAddTargets()`、`BatchAddRules()`、`SetSetting()` 全部使用 `s.db.Exec()`（直接操作 `*sql.DB`），而非事务对象 `tx.Exec()`。

**原因分析：** Go `database/sql` 中 `*sql.DB.Begin()` 返回的 `*sql.Tx` 会占用一个连接。`s.db.Exec()` 使用连接池中的**另一个**连接，因此不在同一事务中。结果：
- `ClearAll()` 立即提交清空操作（不可回滚）
- 若后续 `BatchAddTargets`/`BatchAddRules` 失败，事务回滚不会撤销 `ClearAll()`
- 用户原有配置丢失且新配置未写入 = **不可逆数据丢失**

**影响范围：** 配置导入中途失败时数据库处于"已清空但未导入完整"的不一致状态，原有配置永久丢失。

**推荐修复方案：**
- **方案 A（推荐）：** 让 `ClearAll`、`BatchAddTargets`、`BatchAddRules` 接受可选的 `*sql.Tx` 参数（事务版本），或新增 `ClearAllTx(tx *sql.Tx)` 等事务变体
- **方案 B：** 在 `handleConfigImport` 中内联所有 SQL 操作（绕过 Store 方法，直接使用 tx）

**实施计划**（已验证：`WithTransaction` 已存在但回调内方法均使用 `s.db.Exec()` 而非 `tx.Exec()`）：

1. **`config/store.go`** — 新增 4 个 Tx 变体方法：
   - `ClearAllTx(tx *sql.Tx) error` — 用 `tx.Exec()` 执行 DELETE
   - `AddTargetTx(tx *sql.Tx, t TargetConfig) error` — 用 `tx.Exec()` 插入
   - `AddRuleTx(tx *sql.Tx, r DomainRule) error` — 用 `tx.Exec()` 插入
   - `SetSettingTx(tx *sql.Tx, key, value string) error` — 用 `tx.Exec()` 写入
2. **`webui/api/settings.go`** — `handleConfigImport` 回调内替换调用：
   - `d.Store.ClearAll()` → `d.Store.ClearAllTx(tx)`
   - `d.Store.AddTarget(t)` → `d.Store.AddTargetTx(tx, t)`
   - `d.Store.AddRule(r)` → `d.Store.AddRuleTx(tx, r)`
   - `d.Store.SetSetting(k, v)` → `d.Store.SetSettingTx(tx, k, v)`
3. **`BatchAddTargets`/`BatchAddRules`** — 新增 `BatchAddTargetsTx(tx, targets)` 和 `BatchAddRulesTx(tx, rules)`，内部循环调用对应的 Tx 单条方法
4. 验证：导入中途模拟失败，确认数据库未被清空（事务回滚生效）

---

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

**实施计划**（已验证：后端 `GetTargets`/`GetRules` SELECT 已含 `id` 列，Scan 已捕获 `t.ID`/`r.ID`，API 返回含 `id` 字段）：

1. **`Targets.vue`**：
   - L32：`openEdit(row: any, index: number)` → `openEdit(row: any)`，L33：`editingId.value = row.id`
   - L51：`deleteTarget(index: number)` → `deleteTarget(row: any)`，L52：`fetch('/api/targets/${row.id}')`
   - L79-86：render 函数中 `onClick: () => openEdit(row, index)` → `onClick: () => openEdit(row)`，`onClick: () => deleteTarget(index)` → `onClick: () => deleteTarget(row)`
2. **`Rules.vue`**：
   - L36：`openEdit(row: any, index: number)` → `openEdit(row: any)`，L37：`editingId.value = row.id`
   - L55：`deleteRule(index: number)` → `deleteRule(row: any)`，L56：`fetch('/api/rules/${row.id}')`
   - L76-84：render 函数同 Targets.vue 模式修改
3. 验证：添加 3 个目标 → 删除第 2 个 → 编辑第 3 个，确认操作的是正确行

---

### [R11-04] CI/CD 流水线缺少前端构建步骤（[BLD-03]/[BLD-04] 确认仍存在）

- **严重度：** 高
- **当前状态：** 待修复
- **所属模块：** CI/CD
- **涉及文件：** `.github/workflows/docker-publish.yml`（L43 `go build -v ./...`）、`.github/workflows/release.yml`（L39 `go build -v ./...`）
- **原始记录：** Issue1.md [BLD-01]/[BLD-03]/[BLD-04]（同一根因：前端构建未纳入任何自动化流程，[BLD-01] 为根因——`dist/` 目录缺失，[BLD-03]/[BLD-04] 为 CI/CD 两个表现面）

**现象描述：** 两个 CI workflow 均包含 `go build -v ./...` 编译检查步骤，但缺少 Node.js 环境安装和 `npm ci && npm run build` 前端构建步骤。`webui/embed.go` 使用 `//go:embed frontend/dist`，而 `dist/` 目录被 `.gitignore` 排除，checkout 后不存在。`go build` 会因 embed 找不到文件而失败。

**原因分析：** [BLD-01]/[BLD-03]/[BLD-04] 的修复方案已在 Issue1.md 详细说明，但均未实施。`docker-publish.yml` 的 Docker 构建步骤（使用 `build/Dockerfile`）可内部处理前端构建，但之前的 `go build ./...` 编译检查无法通过。`release.yml` 直接编译多平台二进制，完全无前端构建环节。

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

**实施计划**（已验证：两个 workflow 在 Go 编译步骤前均无前端构建，`build/Dockerfile` 已含 frontend-builder 阶段）：

1. **`.github/workflows/docker-publish.yml`**：在 L42 `编译检查` 步骤前插入 `设置 Node.js` + `构建前端` 步骤（如上 YAML）。注意：Docker 镜像构建步骤（L70-84）通过 `build/Dockerfile` 内部已有前端构建阶段，此处仅为让 L43 的 `go build -v ./...` 编译检查通过
2. **`.github/workflows/release.yml`**：在 L37 `编译检查 + 测试` 步骤前插入 `设置 Node.js` + `构建前端` 步骤。前端构建一次即可供后续所有 GOOS/GOARCH 编译使用
3. 验证：`act push` 或提交 PR 触发 CI，确认编译检查通过

---

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

**实施计划**（用户选择方案B — 0-based，已验证影响范围）：

1. **`config/env.go` `parseTargetNums`**（L248-263）：
   - L257：`if n < 1 || n > max` → `if n < 0 || n >= max`
   - L258：错误消息 `"targets 编号 %d 超出范围 [1,%d]"` → `"targets 编号 %d 超出范围 [0,%d]"`，参数 `n, max` → `n, max-1`
2. **`config/env_test.go`** `TestParseEnv_Normal`（L18）：
   - 测试数据 `ACCEPT|2|VPN接入` → `ACCEPT|1|VPN接入`（targetCount=2，0-based 有效范围 [0,1]，原值 2 会触发越界）
   - `TestParseEnv_TargetNumOutOfRange` 测试 `ACCEPT|5` 在 0-based 下仍会越界（max=1），无需修改
3. **`.env.example`**：RULES 的 targets 字段说明从"从1开始编号"改为"从0开始编号"
4. **`app/app.go`** L33 无需修改（`i` 本就是 0-based），与新的 0-based RULES 语义一致
5. 验证：`go test ./config/...` 全部通过

---

### [R11-06] `app.Run` 的 `mode` 参数未被使用

- **严重度：** 低
- **当前状态：** 待修复
- **所属模块：** App 生命周期
- **涉及文件：** `app/app.go`（L16）
- **原始记录：** Issue1.md [DSC-01] "✅ 已裁定-移除mode参数"

**现象描述：** `func Run(cfg *config.Config, mode Mode) error` 接收 `mode` 参数，但函数体内（L17-51）从未引用 `mode`。自 Phase 2 后 WebUI 模式逻辑移到 `main.go` 内联，`app.Run` 仅服务 `.env` 模式。

**推荐修复方案：** 移除 `mode` 参数，签名改为 `Run(cfg *config.Config) error`；同步修改 `main.go` L125 调用为 `app.Run(cfg)`。

**实施计划**（已验证：函数体内确实无任何 `mode` 引用）：

1. **`app/app.go`** L16：`func Run(cfg *config.Config, mode Mode) error` → `func Run(cfg *config.Config) error`
2. **`main.go`** L125：`app.Run(cfg, mode)` → `app.Run(cfg)`
3. 验证：`go build ./...` 通过

---

### [R11-07] README DNS 默认值与代码不一致

- **严重度：** 低
- **当前状态：** 待修复
- **所属模块：** 文档
- **涉及文件：** `README.md`（L91）

**现象描述：** README L91 表格中 DNS 默认值显示为 `8.8.8.8:53`，而代码实际默认值为 `223.5.5.5`（`config/env.go` L30、`config/store.go` L331）。

**推荐修复方案：** 将 README L91 的 DNS 默认值改为 `223.5.5.5`（并注明端口自动补全 :53）。

**实施计划**（已验证：README L91 为 `8.8.8.8:53`，代码默认值为 `223.5.5.5`）：

1. **`README.md`** L91：`8.8.8.8:53` → `223.5.5.5`，描述改为 `"上游 DNS 服务器地址（端口 :53 自动补全）"`
2. 验证：`grep '8.8.8.8' README.md` 无匹配

---

## 第12轮合规验证（2026-07-26）

> 聚焦测试覆盖、API 合规、安全配置、日志规范、编译检查。

### 测试覆盖

- 已有 7 个测试文件覆盖 `config`、`dns`、`internal/portconv`、`internal/tag`、`notifier`、`provider/common` 六个包
- 全部测试通过：`go test ./...` 返回 `ok`，无失败
- `common_test.go` 覆盖 OwnedRules、Diff 新增/不变/删除/域名隔离、TCP+UDP 拆分、SWAS IPv6 跳过、ECS ICMPv6 跳过、ClientPool 复用 — 共 8 个测试用例
- `circuitbreaker_test.go` 覆盖熔断触发、重置、半开探测不计入 — 共 3 个测试用例
- `env_test.go` 覆盖正常解析、续行、默认值、非法协议/动作/编号/Provider、空内容、ICMP 强制 ALL、凭据校验、空规则 — 共 10 个测试用例

### API 合规性

| 检查项 | 结果 | 证据 |
|--------|------|------|
| 全量覆盖类 API | **合规** — 零使用 | `grep ModifyFirewall\|ResetFirewall\|ReplaceFirewall` 在 provider/ 中零匹配 |
| 仅操作入站规则 | **合规** | CVM 仅用 `Ingress`；ECS 仅 `Direction=ingress`；Lighthouse/SWAS 防火墙无方向概念 |
| Egress 操作 | **合规** — 零使用 | CVM `CreateRules`/`DeleteRules` 仅设置 `Ingress` 字段 |

### 安全合规

| 检查项 | 结果 | 证据 |
|--------|------|------|
| WebUI 绑定地址 | **合规** — `127.0.0.1` | `server.go` L51: `127.0.0.1:%d` |
| 凭据不导出 | **合规** | `settings.go` L75-78: `delete(settings, "tc_access_id")` 等 4 个 key |
| 凭据不导入 | **合规** | `settings.go` L125-126: 跳过四个凭据 key |
| 配置导出无凭据 | **合规** | 同凭据不导出 |
| 无 panic | **合规** | `grep panic` 在 webui/ 中零匹配 |

### 日志规范

| 检查项 | 结果 |
|--------|------|
| syncer/ 包 | **合规** — 全部使用 `log/slog`，无 `fmt.Print` |
| provider/ 包 | **合规** — 全部使用 `log/slog`，无 `fmt.Print` |
| dns/ 包 | **合规** — 无日志输出（纯函数） |
| webui/api/ | **合规** — 3 处 slog（2 Warn + 1 Info），无 `fmt.Print` |
| main.go | **合规** — 启动前错误用 `fmt.Fprintf(os.Stderr)`（slog 未初始化），运行时用 slog |

### 编译与静态检查

- `go vet ./...` — **通过**，零警告
- `go build ./...` — **通过**，零错误
- `go test ./...` — **通过**，所有 6 个测试包 `ok`

### Documents 目录

- 4 个子目录（AliyunECSAPIGuide、AliyunSASAPIGuide、TencentCVMAPIGuide、TencentLighthouseAPIGuide）
- 包含各云厂商 API 参考文档（中文 + 英文），与编码实现中的 API 使用一致
- `.DS_Store` 文件存在（macOS 系统文件），建议通过 .gitignore 排除

---

## 待规划项实施方案

### [WEB-06] 前端缺少 /advanced、/alerts 页面

- **严重度：** 中 | **状态：** 待规划
- **涉及文件：** `webui/frontend/src/main.ts`、`App.vue`、新增 `views/Advanced.vue`、`views/Alerts.vue`、新增 `webui/api/alerts.go`
- **实施范围：** 前后端一起（前端页面 + 告警配置 API）

**后端 API（新增 `webui/api/alerts.go`）：**
- `GET /api/alerts` — 返回 `{email: AlertEmailConfig, webhook: AlertWebhookConfig}`（从 alert_email/alert_webhook 表读取）
- `PUT /api/alerts` — 保存两个配置，写入 alert_email 和 alert_webhook 表，触发 notifyReload

**前端页面：**
- `views/Advanced.vue`：使用 `NTabs` 分 3 个面板，复用现有 `POST /api/sync/dryrun`、`GET /api/config/export`、`POST /api/config/import`、`POST /api/test-connection` 端点
- `views/Alerts.vue`：SMTP 表单（host/port/user/pass/from/to）+ Webhook URL 输入框 + 启用开关，通过 `GET /api/alerts` 加载、`PUT /api/alerts` 保存

**路由与菜单：**
- `main.ts`：新增 `/advanced` 和 `/alerts` 路由
- `App.vue` `menuOptions`：新增"高级功能"和"告警配置"菜单项

**依赖关系：** WEB-06 的 `/alerts` 页面 + `alerts.go` API 为 [FEA-02] 提供配置存储和 UI 基础，需先完成。

---

### [FEA-02] 告警通知器未接入 EventBus

- **严重度：** 中 | **状态：** 待规划
- **涉及文件：** `config/store.go`（新增 alert 表 + CRUD）、`webui/api/alerts.go`（API）、`main.go`（启动时注册）
- **前置依赖：** [WEB-06]（提供 alert_email/alert_webhook 表 + API）
- **存储方案：** 两张独立表（用户选择）

**数据库 Schema（`config/store.go` initTables）：**
```sql
CREATE TABLE IF NOT EXISTS alert_email (
    id INTEGER PRIMARY KEY DEFAULT 1,
    enabled INTEGER DEFAULT 0,
    host TEXT DEFAULT '',
    port TEXT DEFAULT '587',
    username TEXT DEFAULT '',
    password TEXT DEFAULT '',
    from_addr TEXT DEFAULT '',
    to_addr TEXT DEFAULT ''
);
CREATE TABLE IF NOT EXISTS alert_webhook (
    id INTEGER PRIMARY KEY DEFAULT 1,
    enabled INTEGER DEFAULT 0,
    url TEXT DEFAULT ''
);
```

**Store 方法（`config/store.go`）：**
- `GetAlertEmail() (*AlertEmailConfig, error)` — 读取 alert_email 表（单行）
- `SaveAlertEmail(cfg *AlertEmailConfig) error` — UPSERT alert_email 表
- `GetAlertWebhook() (*AlertWebhookConfig, error)` — 读取 alert_webhook 表
- `SaveAlertWebhook(cfg *AlertWebhookConfig) error` — UPSERT alert_webhook 表

**`config/config.go` 新增结构体：**
```go
type AlertEmailConfig struct {
    Enabled  bool   `json:"enabled"`
    Host     string `json:"host"`
    Port     string `json:"port"`
    Username string `json:"username"`
    Password string `json:"password"`
    FromAddr string `json:"from_addr"`
    ToAddr   string `json:"to_addr"`
}
type AlertWebhookConfig struct {
    Enabled bool   `json:"enabled"`
    URL     string `json:"url"`
}
```

**`main.go` 启动时注册（L88 之后）：**
```go
// 读取告警配置并注册 Notifier（若已启用）
if emailCfg, _ := store.GetAlertEmail(); emailCfg != nil && emailCfg.Enabled {
    notifier := notifier.NewEmailNotifier(notifier.EmailConfig{...})
    s.EventBus().Subscribe(notifier.EventSyncError, notifier)
    s.EventBus().Subscribe(notifier.EventDNSFailed, notifier)
}
if webhookCfg, _ := store.GetAlertWebhook(); webhookCfg != nil && webhookCfg.Enabled {
    notifier := notifier.NewWebhookNotifier(webhookCfg.URL)
    s.EventBus().Subscribe(notifier.EventSyncError, notifier)
    s.EventBus().Subscribe(notifier.EventDNSFailed, notifier)
}
```

**热重载**（`ReloadFunc` 内）：重新读取 alert 配置，重建 Notifier 并重新订阅（EventBus 的 Subscribe 追加模式，需注意避免重复订阅；建议先 Unsubscribe 或重置 subscribers map）。

---

### [FEA-03] CLI 缺少 backup/restore

- **严重度：** 中 | **状态：** 待规划
- **涉及文件：** `app/cli.go`、`config/store.go`（新增 `GetDataDir`）
- **getDataDir 方案：** 提取到 config 包（用户选择）

**`config/store.go` 新增：**
```go
func GetDataDir() string {
    if dir := os.Getenv("FWALIZER_DATA_DIR"); dir != "" {
        return dir
    }
    home, _ := os.UserHomeDir()
    switch runtime.GOOS {
    case "darwin":
        return filepath.Join(home, "Library", "Application Support", "fwalizer")
    case "windows":
        return filepath.Join(os.Getenv("APPDATA"), "fwalizer")
    default:
        return filepath.Join(home, ".config", "fwalizer")
    }
}
```
同时修改 `main.go` 的 `getDataDir()` 为调用 `config.GetDataDir()`。

**`app/cli.go` 新增两个 case：**
- `"backup"`：`config.GetDataDir()` → `config.db` → 复制到 `config.db.bak.{YYYYMMDD_HHmmss}` → 清理旧备份保留最新 5 个
- `"restore"`：接受文件路径参数 → 对备份文件执行 SQLite `PRAGMA integrity_check`（需打开备份文件为临时 Store） → 确认无误后复制覆盖原 `config.db`

---

### [FEA-06] systray 缺少同步触发和开机自启

- **严重度：** 低 | **状态：** 待规划
- **涉及文件：** `app/systray.go`、`main.go`
- **架构方案：** 传回调函数（用户选择）
- **平台范围：** macOS + Windows（用户选择）

**同步触发实施：**
1. **`app/systray.go`** `RunSystray` 签名增加：
   ```go
   func RunSystray(openURL string, onSyncTrigger func())
   ```
2. L47 `// TODO` 改为 `onSyncTrigger()`
3. **`main.go`** 调用处：`app.RunSystray(url, func() { s.TriggerSync() })`

**开机自启实施：**

| 平台 | 方案 | 实现 |
|------|------|------|
| macOS | LaunchAgent plist | 写入 `~/Library/LaunchAgents/com.fwalizer.plist`，包含 `ProgramArguments` 指向二进制路径 + `webui` 模式参数 + `RunAtLoad=true`。菜单项切换通过创建/删除 plist 实现 |
| Windows | 注册表 Run 键 | 写入 `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`，键名 `FWAlizer`，值为二进制完整路径 + `webui` 参数。菜单项切换通过增删注册表键实现 |
| Linux | 不实现 | 由用户通过 systemd user service 或 `.desktop` autostart 自行管理 |

**systray 菜单新增：**
```go
mAutoStart := systray.AddMenuItemCheckbox("开机自启", "启动时自动运行", false)
// 点击时切换：创建/删除 plist（macOS）或增删注册表（Windows），并更新 Checked 状态
```

**注意事项：** 需在 `main.go` 启动时检测当前是否已注册开机自启，初始化 Checkbox 状态。二进制路径可通过 `os.Executable()` 获取。Linux 平台下该菜单项隐藏或禁用。

---
