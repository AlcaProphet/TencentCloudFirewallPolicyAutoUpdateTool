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

### 3.6 main.go 使用 fmt.Println

**修复：** WebUI 模式启动提示改为 `slog.Info()`，导出 `app.InitLogger()` 供 main.go 调用
