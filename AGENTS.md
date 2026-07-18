# AGENTS.md — FWAlizer 项目规范

## 项目概述

**FWAlizer**（Firewall DNS Synchronizer）是一个运行在 Docker 容器中的轻量级自动化工具。它通过 DNS 解析指定域名的 IP 地址，并自动将解析结果同步到腾讯云 Lighthouse 实例的防火墙白名单中。

## 技术栈

| 层 | 技术 | 说明 |
|---|------|------|
| 语言 | **Go 1.25+** | 单一二进制，无运行时依赖 |
| SDK | `github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse` | 腾讯云国内站 Go SDK |
| 部署 | **Docker** (alpine 3.20) | 多阶段构建，最终镜像约 14MB |
| CI/CD | **GitHub Actions** | 自动编译 + 推送到 ghcr.io |
| 配置 | **.env** 文件 | 通过 `docker run --env-file` 传入 |

## 核心设计约束

### 1. 防火墙规则操作约束

- **绝不**使用 `ModifyFirewallRules`（会全量覆盖，误删非本工具管理的规则）
- **只用** `CreateFirewallRules`（增量添加）和 `DeleteFirewallRules`（精确删除）
- 所有由本工具创建的规则，必须通过 `FirewallRuleDescription` 字段标记，格式：
  ```
  [RULE_TAG:hostname] Protocol Port
  ```
  其中 `RULE_TAG` 由 `.env` 中 `RULE_TAG` 变量指定（默认：`auto-dns`）

### 2. DNS 解析约束

- 使用 **自定义 DNS 服务器**（`.env` 中 `DNS_SERVER` 指定，如 `8.8.8.8:53`）
- 通过 Go `net.Resolver` 的 `Dial` 函数指定上游 DNS
- 同时解析 **A 记录（IPv4）** 和 **AAAA 记录（IPv6）**
- IPv4 写入 `CidrBlock` 字段（格式：`1.2.3.4/32`）
- IPv6 写入 `Ipv6CidrBlock` 字段（格式：`2001:db8::1/128`）
- 域名解析失败时：记录 WARN 日志，保留现有规则不变（不删除）

### 3. 定时调度约束

- 使用 Go 标准库 `time.Ticker`，不依赖外部 cron
- 间隔由 `.env` 中 `CHECK_INTERVAL` 控制（如 `5m`、`30m`、`1h`）
- 优雅退出：收到 `SIGTERM`/`SIGINT` 后，完成当前轮次再退出

### 4. 乐观锁约束

- 利用 API 返回的 `FirewallVersion` 检测并发冲突
- 写入失败时自动重试（最多 3 次，指数退避）
- 重试步骤：重新 Describe → 重新 diff → 重新 Create/Delete

### 5. Docker 约束

- 基础镜像：`alpine:3.20`
- 编译镜像：`golang:1.25-alpine`
- `CGO_ENABLED=0` 静态编译
- 以非 root 用户运行（`adduser -D appuser`）
- 仅暴露 stdout 日志（`docker logs` 查看）
- 支持 `HEALTHCHECK`

### 6. 配置约束

- 所有配置通过 `.env` 环境变量传入
- `.env` 文件绝不提交 Git
- 提供 `.env.example` 模板
- 密钥必须使用腾讯云 **CAM 子账号 + 最小权限**

### 7. GitHub Actions 约束

- 推送 tag（如 `v1.0.0`）时自动构建并推送 Docker 镜像到 **ghcr.io**
- 镜像命名：`ghcr.io/<owner>/fwalizer:<tag>`
- 构建平台：`linux/amd64`
- PR 时仅编译检查，不推送镜像

## 配置项参考 (.env)

```env
TENCENTCLOUD_SECRET_ID=     # 必填，腾讯云 SecretId
TENCENTCLOUD_SECRET_KEY=    # 必填，腾讯云 SecretKey
LIGHTHOUSE_INSTANCE_ID=    # 必填，目标 Lighthouse 实例 ID
LIGHTHOUSE_REGION=         # 必填，实例地域（如 ap-guangzhou）
DOMAIN_RULES=              # 必填，域名规则，格式见下
RULE_TAG=auto-dns          # 可选，规则描述前缀
CHECK_INTERVAL=5m          # 可选，检查间隔
DNS_SERVER=8.8.8.8:53      # 可选，DNS 服务器
```

### DOMAIN_RULES 格式

```
host|protocol|ports|action;host|protocol|ports|action;...
```

- `host`: 域名（如 `api.example.com`）
- `protocol`: `TCP` / `UDP` / `TCP+UDP`
- `ports`: 逗号分隔端口号（如 `443,80`）或 `*`（所有端口）
- `action`: `ACCEPT` / `DROP`

示例：
```
DOMAIN_RULES=api.example.com|TCP|443,80|ACCEPT;cdn.example.com|TCP|443|ACCEPT
```

## 项目结构

```
TencentCloudFirewallTool/
├── .env.example                 # 配置模板
├── .gitignore
├── AGENTS.md                    # 本文件
├── Dockerfile
├── Makefile
├── README.md
├── go.mod
├── go.sum
├── main.go                      # 入口
├── .github/
│   └── workflows/
│       └── docker-publish.yml   # CI/CD
├── config/
│   └── config.go                # .env 解析与校验
├── dns/
│   └── resolver.go              # DNS 解析（自定义服务器）
├── firewall/
│   ├── client.go                # Lighthouse SDK 封装
│   ├── rule.go                  # 规则对比 & diff 逻辑
│   └── sync.go                  # 同步调度主循环
└── Ref/                         # API 文档 & SDK 参考
    └── tencentcloud-sdk-go/     # 腾讯云 Go SDK（本地克隆）
```

## 开发约定

- **模块路径**：`github.com/<owner>/fwalizer`（开源仓库地址）
- **Go 版本**：`go 1.25`
- **SDK 依赖**：`github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse`
- **无外部框架依赖**：不使用 gin/echo 等 HTTP 框架，不使用 cron 库，尽量用标准库
- **所有错误必须处理**，不可忽略 `error` 返回值
- **日志使用 `log/slog`**（Go 1.21+ 内置结构化日志）
- **注释使用中文**（面向国内开发者）
