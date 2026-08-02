# AGENTS.md — FWAlizer AI 编码指令

> 本文档是给 AI 编码助手的指令集，也是项目**唯一的强要求文档**（详见「十二、文档体系与优先级」）。
> 项目设计方向见 [Design1.md](./Design1.md)、[Design2.md](./Design2.md) 与 [Design3.md](./Design3.md)（设计构想，非强制），当前构建方案见 [Build4.md](./Build4.md)，历史构建与问题记录见 [Build1.md](./Build1.md)、[Build2.md](./Build2.md)、[Build3.md](./Build3.md)、[Issue1.md](./Issue1.md)、[Issue2.md](./Issue2.md)、[Issue3.md](./Issue3.md)。

---

## 一、项目基本信息

- **模块路径**：`github.com/alcaprophet/fwalizer`
- **Go 版本**：`go 1.25`
- **文档定位与优先级**：编码前先阅读本文件（强要求）。设计构想见 [Design1.md](./Design1.md)、[Design2.md](./Design2.md) 与 [Design3.md](./Design3.md)（非强制，供参考）；详细构建方案见 [Build4.md](./Build4.md)（当前）与 [Build1.md](./Build1.md)、[Build2.md](./Build2.md)、[Build3.md](./Build3.md)（历史归档）；错误与修复记录见 [Issue1.md](./Issue1.md)、[Issue2.md](./Issue2.md)、[Issue3.md](./Issue3.md)（历史归档）

---

## 二、核心编码原则

### 简单轻量化

- 功能做减法，不引入不必要的抽象和重型框架
- 优先使用 Go 标准库（`net/http`、`encoding/json`、`time.Ticker`、`log/slog`）
- 不使用 gin/echo 等 HTTP 框架，不使用外部 cron 库
- 单二进制分发，无运行时依赖

### 不过度防御

- 聚焦核心场景，不对极端边界做过度防御性编程
- 合理假设输入有效性，不做无意义的 nil 检查链
- 错误处理到位即可，不堆叠冗余的 fallback 逻辑

### 安全设计（内部使用导向）

- 项目以**内部使用**为设计前提，WebUI 不针对公开访问设计
- 网络安全边界由用户自己控制（防火墙、VPN、反向代理等）
- WebUI 默认绑定 `127.0.0.1`，端口通过 `WEBUI_PORT` 配置（默认 `60200`，若被占用自动在 50000–65535 范围随机选择可用端口；参见 `webui/server.go` 的 `findAvailablePort`）
- Docker 用户通过 `-p` 自行决定暴露范围
- 凭据通过独立环境变量传入，不与资源声明混合

### 开箱即用

- 最小化前置依赖，首次运行即可工作
- 配置有合理默认值，无必填项阻塞启动
- WebUI 模式下自动打开浏览器，无需手动操作

### 不确定时主动提问

- 遇到模糊需求、多种可行方案或技术取舍时，使用提问工具询问用户
- 提问时**必须附上推荐选项**并简要说明理由
- 不要在假设下自行决定关键设计

---

## 三、防火墙规则操作约束

- **绝不**使用全量覆盖类 API（如 Lighthouse 的 `ModifyFirewallRules`、CVM 的“重置安全组规则”），会误删非本工具管理的规则
- **只用**增量添加 + 精确删除（各云对应 API 详见 Build1.md 第五节）
- 所有由本工具创建的规则，通过对应的描述字段标记，格式：
  ```
  [TAG] comment
  ```
  示例：`[auto-dns] 生产API`
- 不同云厂商的规则标识字段不同（详见 Build1.md），均以 `[TAG]` 前缀识别
- 删除时“规则已不存在”视为成功（幂等），不报错
- 添加时“规则已存在”视为成功，WARN 日志并跳过
- 支持协议：TCP / UDP / TCP+UDP / **ICMP**（ICMP 时端口由各 Provider 按 API 要求处理：Lighthouse 传 ALL，阿里云传 -1/-1，CVM 省略 Port 字段）
- **TCP+UDP 协议拆分：** 仅阿里云 SWAS 原生支持 TCP+UDP，Lighthouse/CVM/ECS 均不支持，由 `buildDesired()` 自动拆分为 TCP + UDP 两条规则
- **IPv6+ICMP 处理：** Lighthouse 使用 ICMPv6 协议，CVM 使用 ICMPV6 协议，ECS 不支持（AuthorizeSecurityGroup 无 ICMPv6，直接跳过并 WARN）
- 端口格式：单端口、逗号分隔、范围（`8000-8010`）、`ALL`
- 腾讯云 CVM 安全组规则上限 **100 条**，接近上限时停止新增并告警

---

## 四、DNS 解析约束

- 使用自定义 DNS 服务器（`DNS` 环境变量指定）
- 通过 Go `net.Resolver` 的 `Dial` 函数指定上游 DNS
- 同时解析 **A 记录（IPv4）** 和 **AAAA 记录（IPv6）**
- IPv4 → `CidrBlock` 字段（格式 `1.2.3.4/32`）
- IPv6 → `Ipv6CidrBlock` 字段（格式 `2001:db8::1/128`）
- 域名解析失败：记录 WARN 日志，保留现有规则不变（不删除）
- 超时统一为 **10s**（连接 + 整体，可通过 `DNS_TIMEOUT` 配置）
- 渐进式熔断：连续失败达阈值后熔断，半开状态每轮探测一次，成功后解除
- ⚠️ 仅支持单台服务器场景（少量 IP），不支持 CDN 等返回大量 IP 的域名

---

## 五、同步调度约束

- 使用 `time.Ticker`，不依赖外部 cron
- 间隔由 `INTERVAL` 环境变量控制（如 `5m`、`30m`、`1h`）
- 优雅退出：收到 `SIGTERM`/`SIGINT` 后，完成当前轮次再退出
- 支持配置热重载（WebUI 修改后通过 channel 通知 Syncer）
- 同步全局开关（`SYNC_ENABLED`/`sync_enabled`，默认 true）：暂停时 ticker 与手动 trigger 均不触发同步；模拟测试与连接测试不受影响（独立于 Run() 主循环）

---

## 六、乐观锁与重试

- 每次写入前重新拉取最新规则状态（Describe → Diff → Create/Delete）
- 不传入版本号参数（Lighthouse 的 `FirewallVersion`、CVM 的 `Version`，由云 API 自行管理）
- 写入失败自动重试（最多 3 次，指数退避）
- 重试时重新走完整流程：Describe → Diff → Create/Delete

---

## 七、API 频率限制

- 不同云厂商频率限制不同，取对应间隔（详见 Build1.md）
- 同一云厂商内串行处理（共享配额），域名之间加入间隔
- 不同云厂商可并行同步（API 配额独立）

---

## 八、Docker 约束

- 基础镜像：`alpine:3.20`
- 编译镜像：`golang:1.25-alpine`
- `CGO_ENABLED=0` 静态编译（Docker 构建）
- 非 root 用户运行（`adduser -D appuser`）
- 日志输出到 stdout（Text 格式，`docker logs` 查看）
- 支持 `HEALTHCHECK`（WebUI 模式用 HTTP 端点，`.env` 模式用进程检测）

---

## 九、配置约束

- `.env` 文件**不提交 Git**
- 提供 `.env.example` 模板
- 密钥使用云厂商 **CAM 子账号 + 最小权限**
- 凭据按云厂商独立环境变量（不嵌入 TARGETS）

---

## 十、CI/CD 约束

- 推送 tag（如 `v1.0.0`）时自动构建 Docker 镜像推送到 **ghcr.io**
- 镜像命名：`ghcr.io/alcaprophet/fwalizer:<tag>`
- 构建平台：`linux/amd64`
- PR 时仅编译检查，不推送镜像

---

## 十一、代码规范

- **所有 error 必须处理**，不可忽略返回值
- 日志使用 `log/slog`（Go 1.21+ 内置结构化日志）
- 注释使用**中文**（面向国内开发者）
- 遵守 `Documents/` 中的 API 文档要求（参数格式、字段长度限制、频率限制）
- 多云抽象基于 Provider 接口 + 工厂注册模式（详见 Build1.md）
- 桌面端系统托盘功能**已搁置**（代码归档至 `desktop/`，详见 [FutureDesktopDevelop.md](./FutureDesktopDevelop.md)）
- 日志多路复用器 `MultiHandler` 统一定义在 `app/logutil.go`（消除与 `webui/api/logstream.go` 的重复）
- WebUI 模式通过 pidfile（`config/pidfile.go` + 平台文件）防止多实例运行
- 事件类型：全局同步完成用 `EventSyncComplete`，逐域名同步完成用 `EventDomainSyncComplete`（定义于 `notifier/bus.go`）
- 同步全局开关：`POST /api/sync/pause|resume` 端点（先写 DB 后通知 Syncer）；`SyncStatus.enabled` 字段；前端「运行测试」页（路由 `/run-test`）统一承载模拟测试与连接测试（`DryRunResponse{results, warnings}` 包装、`to_add`/`to_delete` 为规则明细数组）

---

## 十二、文档体系与优先级

### 12.1 文档定位与优先级（本文件为唯一强要求）

| 文档类型 | 文件 | 定位 | 约束力 |
|---------|------|------|--------|
| **强要求** | **AGENTS.md（本文件）** | AI 编码指令与约束 | **唯一强要求，尽量不违背** |
| 设计构想 | [Design1.md](./Design1.md) / [Design2.md](./Design2.md) / [Design3.md](./Design3.md) | 设计大方向、架构构想、决策记录 | 非强制，供参考 |
| 构建方案 | [Build1.md](./Build1.md) / [Build2.md](./Build2.md) / [Build3.md](./Build3.md)（历史归档）、[Build4.md](./Build4.md)（当前） | 详细的分步构建方案与验收命令 | 非强制，执行建议 |
| 问题记录 | [Issue1.md](./Issue1.md) / [Issue2.md](./Issue2.md) / [Issue3.md](./Issue3.md)（历史归档） | 记录的错误与修复方案 | 非强制，经验参考 |

**执行规则：**

- 只有 **AGENTS.md** 是强要求文档，其他类型文档均为**设计取向，不是强规则**，不需要严格遵守
- **Design 文档**描述设计大方向与构想；**Build 文档**描述详细的构建方案；**Issue 文档**记录错误与修复方案
- 若 Design / Build / Issue 文档之间存在冲突，或与 AGENTS.md 冲突：**提示用户并让用户做决策**，不擅自选择遵守哪一份
- 若构想本身存在冲突，同样**提示用户**，由用户决策
- Design1.md 等设计文档中的内容不是强制性规定，仅是设计类构想

### 12.2 文档清单

| 文档 | 目标读者 | 内容 | 状态 |
|------|---------|------|------|
| AGENTS.md（本文件） | AI 编码助手 | 编码指令与约束（**唯一强要求**） | 活跃 |
| [Design1.md](./Design1.md) | 人类（开发者/用户） | 架构设计、需求、决策、路线图（设计构想） | 活跃 |
| [Design2.md](./Design2.md) | 人类（开发者/用户） | 同步全局开关 + 运行测试页设计（设计构想） | 活跃 |
| [Design3.md](./Design3.md) | 人类（开发者/用户） | WebUI 体验优化与同步日志修复设计（设计构想） | 活跃 |
| [Build4.md](./Build4.md) | 开发者 | 当前构建方案：WebUI 体验优化 + 同步日志修复（Step 1-8） | 活跃 |
| [Build3.md](./Build3.md) | 开发者 | 同步全局开关 + 运行测试页构建（Step 1-13，已全部验收通过） | 历史归档 |
| [Build1.md](./Build1.md) | 开发者 | 原始构建计划与技术实现细节（Step 1-16，已全部完成，技术参考） | 历史归档 |
| [Build2.md](./Build2.md) | 开发者 | 修复与功能构建计划（Step 1-11，已全部验收通过） | 历史归档 |
| [Issue3.md](./Issue3.md) | 开发者 | 第13-15轮审查问题与修复记录 | 历史归档 |
| [Issue2.md](./Issue2.md) | 开发者 | 第11-12轮审查问题与合规验证记录 | 历史归档 |
| [Issue1.md](./Issue1.md) | 开发者 | 第1-10轮审查历史问题精简归档 | 历史归档 |
| [FutureDesktopDevelop.md](./FutureDesktopDevelop.md) | 开发者 | 桌面端功能搁置记录与后续重启思路 | 归档 |
