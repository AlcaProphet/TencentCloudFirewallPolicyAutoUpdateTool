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
