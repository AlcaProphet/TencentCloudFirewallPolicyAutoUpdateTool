# FWAlizer 问题修复与功能构建（Build2）

> 依据 [Issue2.md](./Issue2.md) 全部待修复项（R11-01~R11-07）和待规划项（WEB-06/FEA-02/FEA-03/FEA-06）的实施计划。
> 参考 [Build1.md](./Build1.md) 的分步构建规范和代码标准，遵循 [AGENTS.md](./AGENTS.md) 的编码约束。
>
> 覆盖情况：R11 修复 7 项（Step 1–7）+ 待规划 4 项（Step 8–11），共 11 步。

---

## 构建进度追踪

| Step | 编号 | 内容 | 严重度 | 状态 |
|------|------|------|--------|------|
| 1 | R11-01 | `.dockerignore` 去重 | 🟡 中 | ☑ 已完成，待验收 |
| 2 | R11-02 | 配置导入事务保护 | 🔴 高 | ☑ 已完成，待验收 |
| 3 | R11-03 | 前端数组索引→DB ID | 🔴 高 | ☑ 已完成，待验收 |
| 4 | R11-04 | CI/CD 前端构建步骤 | 🔴 高 | ☑ 已完成，待验收 |
| 5 | R11-05 | .env 模式 0-based 规则过滤 | 🟡 中 | ☑ 已完成，待验收 |
| 6 | R11-06 | 移除 `app.Run` mode 参数 | ⚪ 低 | ☑ 已完成，待验收 |
| 7 | R11-07 | README DNS 默认值同步 | ⚪ 低 | ☑ 已完成，待验收 |
| 8 | WEB-06 | `/advanced` + `/alerts` 页面 + 告警 API | 🟡 中 | ☑ 已完成，待验收 |
| 9 | FEA-02 | 告警通知器接入 EventBus | 🟡 中 | ☑ 已完成，待验收 |
| 10 | FEA-03 | CLI `backup` / `restore` | 🟡 中 | ☑ 已完成，待验收 |
| 11 | FEA-06 | systray 同步触发 + 开机自启 | ⚪ 低 | ☑ 已完成，待验收 |

> 状态标记：☐ 未开始 / ◧ 进行中 / ☑ 已完成

---

## 分步构建计划

> 原则：每一步完成后均可编译、可测试。不跳步，不并行多步。
> AI 执行指令：每次仅执行一个 Step，完成后运行验收命令，确认通过后再进入下一步。
> **先修复后构建**：Step 1–7（修复）→ Step 8–11（构建）。

---

### Step 1：R11-01 `.dockerignore` 去重

**目标：** 清理 `.dockerignore` 重复条目，保留 7 行核心排除项。

**前置条件：** 无

**涉及文件：** `.dockerignore`

**操作：**
1. 读取 `.dockerignore`（当前 19 行）
2. 保留 L1-7（`Documents/`、`*.md`、`.env`、`.git/`、`Dockerfile`、`.dockerignore`、`Makefile`）
3. 删除 L8-19（全部重复条目和不存在的 `Ref/` 目录）

**验收：**
```bash
cat .dockerignore
# 预期输出：仅 7 行，无重复，无 Ref/
```

---

### Step 2：R11-02 配置导入事务保护（🔴 高）

**目标：** `handleConfigImport` 的事务回调真正使用 `tx` 执行 SQL，确保导入失败时完整回滚。

**前置条件：** Step 1 完成

**涉及文件：** `config/store.go`、`webui/api/settings.go`

**操作：**

#### 2.1 `config/store.go` — 新增 6 个 Tx 变体方法

在现有 `ClearAll`、`AddTarget`、`AddRule`、`SetSetting` 方法之后，以及 `BatchAddTargets`、`BatchAddRules` 之后，新增对应的 Tx 版本：

```go
// ClearAllTx 在事务中清空所有配置
func (s *Store) ClearAllTx(tx *sql.Tx) error {
    _, err := tx.Exec("DELETE FROM targets; DELETE FROM rules; DELETE FROM settings;")
    return err
}

// AddTargetTx 在事务中添加目标
func (s *Store) AddTargetTx(tx *sql.Tx, t TargetConfig) error {
    _, err := tx.Exec(
        "INSERT INTO targets (cloud_type, region, resource_id) VALUES (?, ?, ?)",
        string(t.CloudType), t.Region, t.ResourceID,
    )
    return err
}

// AddRuleTx 在事务中添加域名规则
func (s *Store) AddRuleTx(tx *sql.Tx, r DomainRule) error {
    targetsJSON, _ := json.Marshal(r.Targets)
    enableIPv6 := 0
    if r.EnableIPv6 {
        enableIPv6 = 1
    }
    _, err := tx.Exec(
        "INSERT INTO rules (host, protocol, ports, action, targets, comment, enable_ipv6) VALUES (?, ?, ?, ?, ?, ?, ?)",
        r.Host, r.Protocol, r.Ports, r.Action, string(targetsJSON), r.Comment, enableIPv6,
    )
    return err
}

// SetSettingTx 在事务中写入单项配置
func (s *Store) SetSettingTx(tx *sql.Tx, key, value string) error {
    _, err := tx.Exec(
        "INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)",
        key, value,
    )
    return err
}

// BatchAddTargetsTx 在事务中批量添加目标
func (s *Store) BatchAddTargetsTx(tx *sql.Tx, targets []TargetConfig) error {
    for _, t := range targets {
        if err := s.AddTargetTx(tx, t); err != nil {
            return err
        }
    }
    return nil
}

// BatchAddRulesTx 在事务中批量添加规则
func (s *Store) BatchAddRulesTx(tx *sql.Tx, rules []DomainRule) error {
    for _, r := range rules {
        if err := s.AddRuleTx(tx, r); err != nil {
            return err
        }
    }
    return nil
}
```

#### 2.2 `webui/api/settings.go` — 替换 `handleConfigImport` 回调内调用

将 L111-133 事务回调内的 4 处调用替换为 Tx 版本：

| 原调用 | 替换为 |
|--------|--------|
| `d.Store.ClearAll()` | `d.Store.ClearAllTx(tx)` |
| `d.Store.BatchAddTargets(imp.Targets)` | `d.Store.BatchAddTargetsTx(tx, imp.Targets)` |
| `d.Store.BatchAddRules(imp.Rules)` | `d.Store.BatchAddRulesTx(tx, imp.Rules)` |
| `d.Store.SetSetting(k, v)` | `d.Store.SetSettingTx(tx, k, v)` |

**验收：**
```bash
go build ./...
go vet ./...
```

---

### Step 3：R11-03 前端数组索引→DB ID（🔴 高）

**目标：** `Targets.vue` 和 `Rules.vue` 的编辑/删除操作使用 `row.id` 替代 `index + 1`。

**前置条件：** Step 2 完成

**涉及文件：** `webui/frontend/src/views/Targets.vue`、`webui/frontend/src/views/Rules.vue`

**操作：**

#### 3.1 `Targets.vue`

| 行号 | 修改 |
|------|------|
| L32 | `function openEdit(row: any, index: number)` → `function openEdit(row: any)` |
| L33 | `editingId.value = index + 1` → `editingId.value = row.id` |
| L51 | `async function deleteTarget(index: number)` → `async function deleteTarget(row: any)` |
| L52 | `` fetch(`/api/targets/${index + 1}`) `` → `` fetch(`/api/targets/${row.id}`) `` |
| L79-86 | render 函数内：`onClick: () => openEdit(row, index)` → `onClick: () => openEdit(row)`；`onClick: () => deleteTarget(index)` → `onClick: () => deleteTarget(row)` |

#### 3.2 `Rules.vue`

| 行号 | 修改 |
|------|------|
| L36 | `function openEdit(row: any, index: number)` → `function openEdit(row: any)` |
| L37 | `editingId.value = index + 1` → `editingId.value = row.id` |
| L55 | `async function deleteRule(index: number)` → `async function deleteRule(row: any)` |
| L56 | `` fetch(`/api/rules/${index + 1}`) `` → `` fetch(`/api/rules/${row.id}`) `` |
| L76-84 | render 函数内：同 Targets.vue 模式修改 |

**验收：**
```bash
cd webui/frontend && npm run build && cd ../..
go build ./...
```

---

### Step 4：R11-04 CI/CD 前端构建步骤（🔴 高）

**目标：** 两个 CI workflow 在 Go 编译前先构建前端，确保 `//go:embed frontend/dist` 能找到文件。

**前置条件：** Step 3 完成

**涉及文件：** `.github/workflows/docker-publish.yml`、`.github/workflows/release.yml`

**操作：**

#### 4.1 `docker-publish.yml`

在 L42 `编译检查` 步骤之前（即 `- name: 编译检查` 那一行之前）插入：

```yaml
      - name: 设置 Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: 构建前端
        run: cd webui/frontend && npm ci && npm run build
```

> **注意：** Docker 镜像构建步骤（L70-84）通过 `build/Dockerfile` 内部已有前端构建阶段。此处仅为让 L43 的 `go build -v ./...` 编译检查通过。插入位置在 `更新所有 SDK 到最新版` 步骤（L31-40）之后、`编译检查` 步骤（L42）之前。

#### 4.2 `release.yml`

在 L37 `编译检查 + 测试` 步骤之前插入同样的两步。前端构建一次即可供后续所有 GOOS/GOARCH 编译使用。

**验收：**
```bash
# 手动检查 YAML 缩进正确（GitHub Actions 无本地运行环境）
# 确认插入的步骤与已有步骤缩进一致（2 空格）
# 或用 Python 验证 YAML 语法：python3 -c "import yaml; yaml.safe_load(open('.github/workflows/docker-publish.yml'))"
```

---

### Step 5：R11-05 `.env` 模式 0-based 规则过滤（🟡 中）

**目标：** RULES 的 targets 编号从 1-based 改为 0-based，与 Provider 数组索引语义一致。

**前置条件：** Step 4 完成

**涉及文件：** `config/env.go`、`config/env_test.go`、`.env.example`、`README.md`

**操作：**

#### 5.1 `config/env.go` `parseTargetNums`

| 行号 | 修改 |
|------|------|
| L248 | 注释：`// parseTargetNums 解析 "1,3" → []int{1,3}` → `// parseTargetNums 解析 "0,2" → []int{0,2}` |
| L257 | `if n < 1 \|\| n > max` → `if n < 0 \|\| n >= max` |
| L258 | `"targets 编号 %d 超出范围 [1,%d]"` → `"targets 编号 %d 超出范围 [0,%d]"` |
| L258 | 参数 `n, max` → `n, max-1` |

#### 5.2 `config/env_test.go` `TestParseEnv_Normal`

| 行号 | 修改 |
|------|------|
| L18 | 测试数据：`ACCEPT\|2\|VPN接入` → `ACCEPT\|1\|VPN接入` |
| L54 | 断言：`cfg.DomainRules[1].Targets[0] != 2` → `cfg.DomainRules[1].Targets[0] != 1`，want 消息同步改为 `want 1` |

> 原因：targetCount=2，0-based 有效范围 [0,1]，原值 2 会触发越界错误。断言必须与测试数据同步修改，否则测试仍会失败。

`TestParseEnv_TargetNumOutOfRange` 测试 `ACCEPT\|5` 在 0-based 下仍会越界（max=1），无需修改。

#### 5.3 `.env.example`

| 行号 | 修改 |
|------|------|
| L35 | targets 字段说明：`应用到的目标编号（逗号分隔）` → `应用到的目标编号（从 0 开始，逗号分隔）` |
| L40 | 示例：`ACCEPT\|2\|VPN接入` → `ACCEPT\|1\|VPN接入` |
| L41 | 示例：`ACCEPT\|1,3\|游戏端口` → `ACCEPT\|0,2\|游戏端口`（4 个目标，0-based 有效范围 [0,3]） |

#### 5.4 `README.md` RULES 语法段落

| 行号 | 修改 |
|------|------|
| L150 | `\| targets \| 目标编号（留空或 \`*\` = 全部） \| \`1,3\` \|` → `\| targets \| 目标编号（从 0 开始，留空或 \`*\` = 全部） \| \`0,2\` \|` |
| L159 | 注释：`# 指定目标（仅第 2 个 TARGETS 条目）` → `# 指定目标（仅第 2 个 TARGETS 条目，编号从 0 开始）` |
| L160 | 示例：`ACCEPT\|2\|VPN接入` → `ACCEPT\|1\|VPN接入` |
| L163 | 示例：`ACCEPT\|1,3\|游戏端口` → `ACCEPT\|0,2\|游戏端口` |
| L173 | 示例：`ACCEPT\|2\|VPN` → `ACCEPT\|1\|VPN` |

**注意事项：**
- `app/app.go` L33（`for i, t := range cfg.Targets` — `i` 本就是 0-based）**无需修改**，与新 0-based RULES 语义一致
- 此为 `.env` 模式的**破坏性变更**，需在 `.env.example` 和 `README.md` 中同步标注

**验收：**
```bash
go test ./config/... -v
# 预期：全部 PASS，无 FAIL
```

---

### Step 6：R11-06 移除 `app.Run` mode 参数（⚪ 低）

**目标：** 移除 `app.Run` 中从未使用的 `mode Mode` 参数。

**前置条件：** Step 5 完成

**涉及文件：** `app/app.go`、`main.go`

**操作：**

| 文件 | 行号 | 修改 |
|------|------|------|
| `app/app.go` | L16 | `func Run(cfg *config.Config, mode Mode) error` → `func Run(cfg *config.Config) error` |
| `main.go` | L125 | `app.Run(cfg, mode)` → `app.Run(cfg)` |

**验收：**
```bash
go build ./...
go vet ./...
```

---

### Step 7：R11-07 README DNS 默认值同步（⚪ 低）

**目标：** README 中 DNS 默认值从 `8.8.8.8:53` 改为 `223.5.5.5`。

**前置条件：** Step 6 完成

**涉及文件：** `README.md`

**操作：**

| 行号 | 修改 |
|------|------|
| L91 | `8.8.8.8:53` → `223.5.5.5`，描述改为 `"上游 DNS 服务器地址（端口 :53 自动补全）"` |

**验收：**
```bash
grep '8.8.8.8' README.md
# 预期：无输出（或仅在历史说明中出现）
```

---

### Step 8：WEB-06 `/advanced` + `/alerts` 页面 + 告警 API（🟡 中）

**目标：** 新增两个前端页面和告警配置的后端 API。

**前置条件：** Step 1–7 全部完成

**涉及文件：**
- 新增 `webui/frontend/src/views/Advanced.vue`
- 新增 `webui/frontend/src/views/Alerts.vue`
- 新增 `webui/api/alerts.go`
- 修改 `webui/frontend/src/main.ts`
- 修改 `webui/frontend/src/App.vue`
- 修改 `config/store.go`（新增 alert 表 + CRUD）
- 修改 `config/config.go`（新增 `AlertEmailConfig`、`AlertWebhookConfig` 结构体）
- 修改 `webui/api/deps.go`（注册新路由）

**操作：**

#### 8.1 `config/config.go` — 新增结构体

在 `DomainRule` 结构体之后添加：

```go
// AlertEmailConfig SMTP 邮件告警配置
type AlertEmailConfig struct {
    Enabled  bool   `json:"enabled"`
    Host     string `json:"host"`
    Port     string `json:"port"`
    Username string `json:"username"`
    Password string `json:"password"`
    FromAddr string `json:"from_addr"`
    ToAddr   string `json:"to_addr"`
}

// AlertWebhookConfig Webhook 告警配置
type AlertWebhookConfig struct {
    Enabled bool   `json:"enabled"`
    URL     string `json:"url"`
}
```

#### 8.2 `config/store.go` — 新增 alert 表和 CRUD

**Schema（在 `initTables` 中添加）：**
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

**新增 Store 方法：**
```go
func (s *Store) GetAlertEmail() (*AlertEmailConfig, error) { /* SELECT * FROM alert_email WHERE id=1 */ }
func (s *Store) SaveAlertEmail(cfg *AlertEmailConfig) error { /* INSERT OR REPLACE INTO alert_email */ }
func (s *Store) GetAlertWebhook() (*AlertWebhookConfig, error) { /* SELECT * FROM alert_webhook WHERE id=1 */ }
func (s *Store) SaveAlertWebhook(cfg *AlertWebhookConfig) error { /* INSERT OR REPLACE INTO alert_webhook */ }
```

**参考代码（GetAlertEmail + SaveAlertEmail）：**
```go
func (s *Store) GetAlertEmail() (*AlertEmailConfig, error) {
    var cfg AlertEmailConfig
    var enabled int
    err := s.db.QueryRow("SELECT enabled, host, port, username, password, from_addr, to_addr FROM alert_email WHERE id = 1").
        Scan(&enabled, &cfg.Host, &cfg.Port, &cfg.Username, &cfg.Password, &cfg.FromAddr, &cfg.ToAddr)
    if err == sql.ErrNoRows {
        return &AlertEmailConfig{}, nil
    }
    if err != nil {
        return nil, err
    }
    cfg.Enabled = enabled != 0
    return &cfg, nil
}

func (s *Store) SaveAlertEmail(cfg *AlertEmailConfig) error {
    enabled := 0
    if cfg.Enabled {
        enabled = 1
    }
    _, err := s.db.Exec(
        `INSERT OR REPLACE INTO alert_email (id, enabled, host, port, username, password, from_addr, to_addr)
         VALUES (1, ?, ?, ?, ?, ?, ?, ?)`,
        enabled, cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.FromAddr, cfg.ToAddr,
    )
    return err
}
```

> `GetAlertWebhook` / `SaveAlertWebhook` 同理，字段更少。

#### 8.3 `webui/api/alerts.go` — 新增告警配置 API

```go
package api

import (
    "encoding/json"
    "net/http"

    "github.com/alcaprophet/fwalizer/config"
)

type alertsResponse struct {
    Email   *config.AlertEmailConfig   `json:"email"`
    Webhook *config.AlertWebhookConfig `json:"webhook"`
}

func (d *Deps) handleGetAlerts(w http.ResponseWriter, r *http.Request) {
    emailCfg, err := d.Store.GetAlertEmail()
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    webhookCfg, err := d.Store.GetAlertWebhook()
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, alertsResponse{Email: emailCfg, Webhook: webhookCfg})
}

func (d *Deps) handlePutAlerts(w http.ResponseWriter, r *http.Request) {
    var req alertsResponse
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "请求格式错误")
        return
    }
    if req.Email != nil {
        if err := d.Store.SaveAlertEmail(req.Email); err != nil {
            writeError(w, http.StatusInternalServerError, err.Error())
            return
        }
    }
    if req.Webhook != nil {
        if err := d.Store.SaveAlertWebhook(req.Webhook); err != nil {
            writeError(w, http.StatusInternalServerError, err.Error())
            return
        }
    }
    d.notifyReload()
    writeJSON(w, http.StatusOK, map[string]string{"message": "保存成功"})
}
```

#### 8.4 `webui/api/deps.go` — 注册新路由

在 `Register` 方法末尾（L58 之后）添加：

```go
mux.HandleFunc("GET /api/alerts", d.handleGetAlerts)
mux.HandleFunc("PUT /api/alerts", d.handlePutAlerts)
```

#### 8.5 `views/Advanced.vue` — 高级功能页面

使用 `NTabs` 分 3 个面板：
- **Dry Run**：`NButton` 触发 `POST /api/sync/dryrun`，`NDataTable` 展示 `{toAdd, toDelete}` 结果
- **配置导入/导出**：导出 `NButton` 触发 `GET /api/config/export`（下载 JSON）；导入 `NUpload` + `NButton` 触发 `POST /api/config/import`
- **连接测试**：`NSelect` 选云产品 + `NInput` 填资源ID/地域 + `NButton` 触发 `POST /api/test-connection` + 结果显示

#### 8.6 `views/Alerts.vue` — 告警配置页面

分两个区域：
- **邮件告警**：`NSwitch`（启用）+ 6 个 `NInput`（host/port/user/pass/from/to），密码字段 `type="password"`
- **Webhook 告警**：`NSwitch`（启用）+ `NInput`（URL）
- 保存按钮触发 `PUT /api/alerts`，加载时 `GET /api/alerts`

#### 8.7 路由与菜单

| 文件 | 修改 |
|------|------|
| `main.ts` | routes 数组新增 `{ path: '/advanced', component: () => import('./views/Advanced.vue') }` 和 `{ path: '/alerts', component: () => import('./views/Alerts.vue') }` |
| `App.vue` | `menuOptions` 数组新增 `{ label: '高级功能', key: '/advanced' }` 和 `{ label: '告警配置', key: '/alerts' }` |

**验收：**
```bash
cd webui/frontend && npm run build && cd ../..
go build ./...
go vet ./...
# 启动 WebUI 模式：./fwalizer
# 浏览器访问 http://127.0.0.1:9090 → 确认侧栏出现"高级功能"和"告警配置"菜单
# 进入 /alerts → 填写 SMTP 配置 → 保存 → 刷新页面确认配置持久化
```

---

### Step 9：FEA-02 告警通知器接入 EventBus（🟡 中）

**目标：** 启动时从 alert 表读取配置，若已启用则创建 Notifier 并订阅 EventBus。

**前置条件：** Step 8 完成（需要 alert 表和 `/api/alerts` API）

**涉及文件：** `main.go`

**操作：**

在 `main.go` WebUI 启动路径中，L89（`StoreLogWriter` 订阅之后）添加：

```go
// 读取告警配置并注册 Notifier
if emailCfg, err := store.GetAlertEmail(); err == nil && emailCfg != nil && emailCfg.Enabled {
    notifierEmail := notifier.NewEmailNotifier(notifier.EmailConfig{
        Host: emailCfg.Host, Port: emailCfg.Port,
        User: emailCfg.Username, Pass: emailCfg.Password,
        From: emailCfg.FromAddr, To: emailCfg.ToAddr,
    })
    s.EventBus().Subscribe(notifier.EventSyncError, notifierEmail)
    s.EventBus().Subscribe(notifier.EventDNSFailed, notifierEmail)
    slog.Info("邮件告警已启用", "to", emailCfg.ToAddr)
}

if webhookCfg, err := store.GetAlertWebhook(); err == nil && webhookCfg != nil && webhookCfg.Enabled {
    notifierWH := notifier.NewWebhookNotifier(webhookCfg.URL)
    s.EventBus().Subscribe(notifier.EventSyncError, notifierWH)
    s.EventBus().Subscribe(notifier.EventDNSFailed, notifierWH)
    slog.Info("Webhook 告警已启用", "url", webhookCfg.URL)
}
```

**热重载处理：** 在 `ReloadFunc`（L93-114）中同样读取告警配置并重新订阅。注意 EventBus 的 `Subscribe` 是追加模式，需避免重复订阅。建议方案：在 ReloadFunc 中不处理告警订阅重建（告警配置变更频率极低，重启生效即可）。

**验收：**
```bash
go build ./...
# 启动后配置 SMTP → 手动触发同步错误 → 确认收到告警邮件
```

---

### Step 10：FEA-03 CLI `backup` / `restore`（🟡 中）

**目标：** CLI 新增备份/恢复子命令。

**前置条件：** Step 7 完成（所有 R11 修复已完成，.env 模式 0-based 规则过滤和 mode 参数移除均已生效）

**涉及文件：** `config/store.go`（新增 `GetDataDir`）、`app/cli.go`、`main.go`

**操作：**

#### 10.1 `config/store.go` — 提取 `GetDataDir`

将当前 `main.go` 的 `getDataDir()` 函数逻辑提取到 `config` 包：

```go
// GetDataDir 获取数据存储目录
func GetDataDir() string {
    if dir := os.Getenv("FWALIZER_DATA_DIR"); dir != "" {
        return dir
    }
    home, err := os.UserHomeDir()
    if err != nil {
        // 回退到当前目录（极端情况）
        return "."
    }
    switch runtime.GOOS {
    case "darwin":
        return filepath.Join(home, "Library", "Application Support", "fwalizer")
    case "windows":
        appdata := os.Getenv("APPDATA")
        if appdata == "" {
            appdata = filepath.Join(home, "AppData", "Roaming")
        }
        return filepath.Join(appdata, "fwalizer")
    default:
        return filepath.Join(home, ".config", "fwalizer")
    }
}
```

需在 `store.go` 头部增加 `"os"`、`"path/filepath"`、`"runtime"` import。

#### 10.2 `main.go` — 替换本地函数

删除本地 `getDataDir()` 函数（L131-154），将 L42 的 `getDataDir()` 调用改为 `config.GetDataDir()`。

删除后 `"runtime"` import 不再被使用，需同步移除（`"path/filepath"` 仍被 L47 使用，保留）。

#### 10.3 `app/cli.go` — 新增 `backup` / `restore`

在 `RunCLI` 的 switch 中增加两个 case：

```go
case "backup":
    dataDir := config.GetDataDir()
    src := filepath.Join(dataDir, "config.db")
    ts := time.Now().Format("20060102_150405")
    dst := filepath.Join(dataDir, fmt.Sprintf("config.db.bak.%s", ts))
    if err := copyFile(src, dst); err != nil {
        fmt.Fprintf(os.Stderr, "备份失败: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("备份成功: %s\n", dst)
    // 清理旧备份（保留最新 5 个）
    cleanOldBackups(dataDir, 5)
    return true

case "restore":
    if len(args) < 3 {
        fmt.Fprintf(os.Stderr, "用法: fwalizer restore <备份文件路径>\n")
        os.Exit(1)
    }
    backupFile := args[2]
    dataDir := config.GetDataDir()
    src := backupFile
    dst := filepath.Join(dataDir, "config.db")
    // 验证备份文件完整性（用 sqlite 打开执行 PRAGMA integrity_check）
    if err := verifyBackup(src); err != nil {
        fmt.Fprintf(os.Stderr, "备份文件校验失败: %v\n", err)
        os.Exit(1)
    }
    if err := copyFile(src, dst); err != nil {
        fmt.Fprintf(os.Stderr, "恢复失败: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("恢复成功，请重启 FWAlizer")
    return true
```

需新增辅助函数 `copyFile`、`cleanOldBackups`、`verifyBackup`（`verifyBackup` 使用 `sql.Open("sqlite", backupFile)` 打开备份文件并执行 `PRAGMA integrity_check`）。`cli.go` 需在 import 中增加 `"time"`、`"path/filepath"`、`"database/sql"`、`_ "modernc.org/sqlite"`（`"fmt"` 已存在）。

**验收：**
```bash
go build ./...
./fwalizer backup              # 应生成 config.db.bak.{timestamp}
./fwalizer restore config.db.bak.xxx  # 应提示恢复成功
```

---

### Step 11：FEA-06 systray 同步触发 + 开机自启（⚪ 低）

**目标：** "立即同步"菜单项实际触发同步；新增"开机自启"菜单项（macOS + Windows）。

**前置条件：** Step 10 完成（main.go 已使用 `config.GetDataDir()`）

**涉及文件：** `app/systray.go`、新增 `app/systray_stub.go`、`syncer/syncer.go`、`main.go`

**操作：**

#### 11.1 `app/systray.go` — 签名增加回调参数

```go
func RunSystray(openURL string, onSyncTrigger func()) {
    systray.Run(func() {
        onSystrayReady(openURL, onSyncTrigger)
    }, func() {
        onSystrayExit()
    })
}

func onSystrayReady(openURL string, onSyncTrigger func()) {
    // ... 现有菜单 ...
    // L47 处：
    // slog.Info("手动触发同步")
    // // TODO: 通过 channel 通知 Syncer 立即同步
    // 改为：
    onSyncTrigger()
    // ...
}
```

**新增"开机自启"菜单项：**
```go
// 在 mSync 之后添加
systray.AddSeparator()
mAutoStart := systray.AddMenuItemCheckbox("开机自启", "启动时自动运行", isAutoStartEnabled())
// 在 goroutine 的 select 中新增：
case <-mAutoStart.ClickedCh:
    if mAutoStart.Checked() {
        disableAutoStart()
        mAutoStart.Uncheck()
    } else {
        enableAutoStart()
        mAutoStart.Check()
    }
```

**“退出”菜单优雅关闭：**

因 `syscall.Kill` 在 Windows 不可用，采用跨平台 channel 方案：

1. `app/systray.go` 新增包级变量：
```go
var quitCh = make(chan struct{})

// QuitCh 返回退出信号 channel（供 main.go 监听）
func QuitCh() <-chan struct{} { return quitCh }
```

2. `mQuit` 处理逻辑改为：
```go
case <-mQuit.ClickedCh:
    systray.Quit()
    close(quitCh) // 通知主 goroutine 优雅退出
    return
```

3. `app/systray_stub.go` 桩文件同步提供：
```go
var quitCh = make(chan struct{})
func QuitCh() <-chan struct{} { return quitCh }
```

4. `syncer/syncer.go` 新增公共方法（因 `doneCh` 为包内私有，main.go 无法直接访问）：
```go
// Wait 等待 Syncer 完全退出（Stop 后调用）
func (s *Syncer) Wait() { <-s.doneCh }
```

5. `main.go` 中替换 `syncer.WaitForSignal(s)` 为：
```go
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
select {
case <-sigCh:
case <-app.QuitCh():
}
slog.Info("收到停止信号，等待当前轮次完成...")
s.Stop()
s.Wait()
```

> 效果：点击“退出”→ 托盘消失 → Syncer 完成当前轮次 → 进程退出。与 Ctrl+C 行为一致。
> `main.go` 需增加 `"os/signal"`、`"syscall"` import；`systray.go` 无需额外 import（channel 操作无依赖）。

#### 11.2 `app/systray_stub.go` — 非桌面构建桩（新增文件）

`RunSystray` 仅在 `//go:build desktop` 下编译，但 `main.go` 无构建标签。需新增桩文件确保非桌面构建编译通过：

```go
//go:build !desktop

package app

// RunSystray 非桌面构建下为空操作
func RunSystray(openURL string, onSyncTrigger func()) {}

var quitCh = make(chan struct{})

// QuitCh 非桌面构建下永远不会关闭（无托盘退出按钮）
func QuitCh() <-chan struct{} { return quitCh }
```

#### 11.3 开机自启平台实现

新增辅助函数（可在 `app/autostart.go` 或 `systray.go` 中，需 `//go:build desktop` 标签）：

| 函数 | 平台 | 实现 |
|------|------|------|
| `isAutoStartEnabled()` | macOS | 检查 `~/Library/LaunchAgents/com.fwalizer.agent.plist` 是否存在 |
| | Windows | 检查 `HKCU\Software\Microsoft\Windows\CurrentVersion\Run\FWAlizer` 注册表键 |
| | Linux | 始终返回 false（不实现） |
| `enableAutoStart()` | macOS | 创建 plist 文件，`RunAtLoad=true`，`ProgramArguments` 为 `os.Executable()` 路径 |
| | Windows | 写入注册表 Run 键 |
| `disableAutoStart()` | macOS | 删除 plist 文件 |
| | Windows | 删除注册表 Run 键 |

> 使用 `//go:build desktop` 标签确保仅在桌面端编译。二进制路径通过 `os.Executable()` 获取。

#### 11.4 `main.go` — 传递同步触发回调

在 WebUI 模式的 `go srv.Start()` 之后、`go s.Run()` 之前插入（必须以 goroutine 调用，因 `systray.Run()` 会阻塞）：

```go
url := fmt.Sprintf("http://127.0.0.1:%d", cfg.WebUIPort)
go app.RunSystray(url, func() { s.TriggerSync() })
```

> 注意：`s.TriggerSync()` 在 `api.Syncer` 接口和 `syncer.Syncer` 中均已实现，直接调用即可。
> 非桌面构建下 `RunSystray` 为空操作（由 11.2 桩文件保证编译通过），goroutine 立即返回。
> 主 goroutine 的阻塞点已替换为 `select{sigCh / app.QuitCh()}`（见“退出”优雅关闭第 5 点），Ctrl+C 和托盘退出均触发优雅关闭。

**验收：**
```bash
# 非桌面构建编译检查（验证桩文件生效）
go build ./...
# 桌面构建（仅 macOS/Windows 可执行）
CGO_ENABLED=1 go build -tags desktop -o fwalizer .
go vet ./...
# 手动验证：
# 1. 启动 ./fwalizer → 托盘出现
# 2. 点击"立即同步" → 检查日志确认同步已触发
# 3. 点击"开机自启" → 检查 plist/注册表已写入
# 4. 再次点击"开机自启" → 检查已清除
# 5. 点击"退出" → 确认日志输出"收到停止信号"后进程正常退出
```

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
- Step 8 → Step 9 是强依赖链（alert 表/API → EventBus 注册），必须按序
- Step 10 可在 Step 7 之后的任意位置执行（CLI 不依赖 WebUI），与 Step 8 同改 `config/store.go` 但无逻辑依赖（各自新增独立方法）
- Step 11 依赖 Step 10（main.go 需先完成 `config.GetDataDir()` 提取）
- 每个 Step 完成后必须：`go build ./... && go vet ./...`（前端 Step 3/8 需额外 `npm run build`，Step 11 需 `CGO_ENABLED=1` 和 `-tags desktop`）
