# FWAlizer — 防火墙 DNS 自动同步工具

**FWAlizer**（Firewall DNS Synchronizer）是一个轻量级自动化工具：定时解析指定域名的 IP 地址，自动同步到云防火墙/安全组白名单中。专为域名 IP 频繁变动的场景设计（如动态 DNS、API 网关、VPN 入口）。

---

## 核心特性

- **多云支持**：腾讯云 Lighthouse / CVM，阿里云轻量云（SWAS）/ ECS，四款云产品统一管控
- **双模式运行**：`.env` 文件驱动（适合服务器/容器）或 WebUI 可视化管理（适合桌面端）
- **增量同步**：仅操作带 `[TAG]` 标记的规则，绝不覆盖手动配置的防火墙规则
- **DNS 熔断保护**：连续解析失败达阈值后自动熔断，半开探测恢复，避免误删规则
- **乐观锁重试**：每次写入前重新拉取最新状态，最多 3 次指数退避重试
- **跨云并行**：不同云厂商并行同步，同厂商内串行（避免触发频率限制）
- **单二进制分发**：前端 WebUI 编译进二进制，无运行时依赖
- **Docker 就绪**：Alpine 基础镜像，非 root 运行，健康检查

---

## 快速开始

### 环境要求

| 工具 | 版本 | 用途 |
|------|------|------|
| Go | 1.25+ | 编译（仅开发时需要） |
| Node.js | 22+（可选） | 前端开发/重新构建时需要 |
| Docker | 20+（可选） | 容器化部署 |

### 最简配置

复制 `.env.example` 为 `.env`，填入必要项即可运行：

```bash
cp .env.example .env
```

最简 `.env`（仅一台腾讯云 Lighthouse + 一条规则）：

```env
TARGETS=tc_lighthouse|lhins-abc123|ap-guangzhou

TC_ACCESS_ID=AKIDxxxxxxxx
TC_ACCESS_KEY=xxxxxxxx

RULES=api.example.com|TCP|443|ACCEPT||生产API
```

### 编译与运行

```bash
# 编译
make build

# 验证配置
./fwalizer validate .env

# 启动同步
./fwalizer
```

### Docker 一键运行

```bash
docker run -d --name fwalizer --restart=always \
  -v $(pwd)/.env:/app/.env:ro \
  ghcr.io/alcaprophet/fwalizer:latest
```

---

## 配置说明

所有配置通过 `.env` 文件设置（`KEY=VALUE` 格式，支持 `#` 注释和 `\` 续行）。

### 基础配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `TARGETS` | （必填） | 云资源目标列表，格式见下方 |
| `RULES` | （必填） | 域名规则列表，格式见下方 |
| `TAG` | `auto-dns` | 规则标记前缀，用于识别本工具创建的规则 |
| `INTERVAL` | `5m` | DNS 检查间隔（如 `30s`、`5m`、`1h`） |
| `LOG_LEVEL` | `info` | 日志级别：`debug` / `info` / `warn` / `error` |
| `FWALIZER_MODE` | 自动检测 | 强制运行模式：`env` / `webui` |

### DNS 配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DNS` | `8.8.8.8:53` | 上游 DNS 服务器地址 |
| `DNS_TIMEOUT` | `10s` | DNS 解析超时时间 |
| `DNS_FAIL_THRESHOLD` | `5` | 连续失败多少次后触发熔断 |

### 腾讯云凭据

| 变量 | 说明 |
|------|------|
| `TC_ACCESS_ID` | 腾讯云 API 密钥 ID（[获取地址](https://console.cloud.tencent.com/cam/capi)） |
| `TC_ACCESS_KEY` | 腾讯云 API 密钥 Key |

### 阿里云凭据

| 变量 | 说明 |
|------|------|
| `ALI_ACCESS_ID` | 阿里云 AccessKey ID（[获取地址](https://ram.console.aliyun.com/manage/ak)） |
| `ALI_ACCESS_KEY` | 阿里云 AccessKey Secret |

### WebUI 配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `WEBUI_PORT` | `9090` | WebUI 监听端口（绑定 127.0.0.1） |

---

## 运行模式

### .env 模式（服务器/容器部署）

当检测到 `TARGETS` 环境变量或 `.env` 文件时自动进入。纯 headless 运行，日志输出到 stdout。

```bash
./fwalizer                    # 从 .env 加载配置并启动同步
./fwalizer validate .env      # 仅校验配置，不启动
./fwalizer version            # 显示版本号
```

### WebUI 模式（桌面/可视化管理）

无 `.env` 文件时自动进入 WebUI 模式，配置存储在 SQLite 数据库中。

- 默认地址：`http://127.0.0.1:9090`
- 数据路径：`~/.config/fwalizer/config.db`
- 支持通过浏览器添加/编辑/删除云资源目标和域名规则
- 修改配置后自动热重载，无需重启

---

## RULES 语法

格式：`host|protocol|ports|action|targets|comment`

| 字段 | 说明 | 示例 |
|------|------|------|
| host | 要解析的域名 | `api.example.com` |
| protocol | 协议：`TCP` / `UDP` / `TCP+UDP` / `ICMP` | `TCP` |
| ports | 端口：单端口、逗号分隔、范围、`ALL` | `443,80` |
| action | 动作：`ACCEPT`（允许）/ `DROP`（拒绝） | `ACCEPT` |
| targets | 目标编号（留空或 `*` = 全部） | `1,3` |
| comment | 可选备注 | `生产API` |

### 示例

```env
# 单域名 + 多端口
RULES=api.example.com|TCP|443,80|ACCEPT||生产API

# 指定目标（仅第 2 个 TARGETS 条目）
RULES=vpn.example.com|UDP|1194|ACCEPT|2|VPN接入

# 端口范围 + 多目标
RULES=game.example.com|TCP|8000-8010|ACCEPT|1,3|游戏端口

# ICMP（Ping），端口自动设为 ALL
RULES=ping.example.com|ICMP|ALL|ACCEPT||允许Ping

# TCP+UDP（仅阿里云 SWAS 原生支持，其他云自动拆分为两条规则）
RULES=voice.example.com|TCP+UDP|5060|ACCEPT||SIP语音

# 多条规则组合（逗号分隔 + 反斜杠换行）
RULES=api.example.com|TCP|443|ACCEPT||API, \
      vpn.example.com|UDP|1194|ACCEPT|2|VPN
```

> **注意**：仅支持单台服务器场景（DNS 返回少量 IP），不支持 CDN 等返回大量 IP 的域名。

---

## 多云 API 权限

建议使用子账号 + 最小权限策略。

### 腾讯云（CAM 子账号）

| 云产品 | 所需权限 |
|--------|----------|
| Lighthouse | `QcloudLighthouseFullAccess`（或自定义：`DescribeFirewallRules` + `CreateFirewallRules` + `DeleteFirewallRules`） |
| CVM（VPC） | `QcloudVPCFullAccess`（或自定义：`DescribeSecurityGroupPolicies` + `CreateSecurityGroupPolicies` + `DeleteSecurityGroupPolicies`） |

### 阿里云（RAM 子账号）

| 云产品 | 所需权限 |
|--------|----------|
| 轻量云（SWAS） | `AliyunSWASFullAccess`（或自定义：`swas:ListFirewallRules` + `swas:CreateFirewallRules` + `swas:DeleteFirewallRules`） |
| ECS | `AliyunECSFullAccess`（或自定义：`ecs:DescribeSecurityGroupAttribute` + `ecs:AuthorizeSecurityGroup` + `ecs:RevokeSecurityGroup`） |

---

## Docker 部署

### 拉取镜像

```bash
docker pull ghcr.io/alcaprophet/fwalizer:latest
```

### .env 模式运行

```bash
docker run -d --name fwalizer --restart=always \
  -v $(pwd)/.env:/app/.env:ro \
  ghcr.io/alcaprophet/fwalizer:latest
```

### WebUI 模式运行

```bash
docker run -d --name fwalizer --restart=always \
  -p 9090:9090 \
  -v fwalizer-data:/home/appuser/.config/fwalizer \
  ghcr.io/alcaprophet/fwalizer:latest
```

### docker-compose 示例

```yaml
version: "3.8"
services:
  fwalizer:
    image: ghcr.io/alcaprophet/fwalizer:latest
    container_name: fwalizer
    restart: always
    volumes:
      - ./.env:/app/.env:ro   # .env 模式
    # 或 WebUI 模式：
    # ports:
    #   - "9090:9090"
    # volumes:
    #   - fwalizer-data:/home/appuser/.config/fwalizer

# volumes:
#   fwalizer-data:
```

### 本地构建镜像

```bash
make docker-build
docker run --rm fwalizer version
```

---

## 开发指南

### 目录结构

```
fwalizer/
├── main.go                  # 入口：模式判定 + 启动
├── app/                     # 应用生命周期（CLI、模式检测、日志初始化）
├── config/                  # 配置模型、.env 解析器、SQLite 存储、校验
├── dns/                     # DNS 解析器 + 熔断器
├── provider/                # 多云抽象层（接口 + 四家 Provider 实现）
├── syncer/                  # 同步引擎（主循环、重试、频率控制）
├── notifier/                # 事件总线 + 告警（邮件、Webhook）
├── webui/                   # WebUI 后端（HTTP API + 前端 embed）
│   └── frontend/            # Vue 3 + Vite + Naive UI 前端源码
├── internal/                # 内部工具（端口转换、标签解析）
├── version/                 # 版本信息（ldflags 注入）
└── build/                   # Dockerfile
```

### 本地开发步骤

```bash
# 1. 克隆项目
git clone https://github.com/alcaprophet/fwalizer.git
cd fwalizer

# 2. 编译后端
make build

# 3. 运行测试
make test

# 4. 代码检查
make vet
```

### 前端开发

```bash
cd webui/frontend

# 安装依赖
npm install

# 启动开发服务器（热重载，代理 API 到 127.0.0.1:9090）
npm run dev

# 构建生产版本（输出到 dist/，go:embed 会自动包含）
npm run build
```

### 构建完整二进制（含前端）

```bash
cd webui/frontend && npm run build && cd ../..
make build
./fwalizer    # 访问 http://127.0.0.1:9090
```

---

## 常见问题（FAQ）

### 1. 启动报错 "加载 .env 失败"？

确保项目根目录存在 `.env` 文件。可以先复制模板：`cp .env.example .env`，然后填入实际配置。

### 2. 如何确认规则是否同步成功？

查看日志输出（默认 `info` 级别会打印每次同步结果）。也可设置 `LOG_LEVEL=debug` 查看详细的 DNS 解析结果和 Diff 计算过程。

### 3. 本工具会不会删除我手动添加的防火墙规则？

**不会。** 本工具仅操作描述字段以 `[TAG]`（默认 `[auto-dns]`）开头的规则。手动创建的规则不带此标记，完全不受影响。

### 4. DNS 解析失败时会删除已有规则吗？

**不会。** DNS 解析失败时保留现有规则不变（仅记录 WARN 日志）。连续失败达到阈值（默认 5 次）后触发熔断，暂停该域名同步，不影响其他域名。

### 5. 支持 IPv6 吗？

腾讯云 Lighthouse、CVM 和阿里云 ECS 支持 IPv6（AAAA 记录）。阿里云轻量云（SWAS）不支持 IPv6，解析到的 IPv6 地址会自动跳过。

### 6. 阿里云 SWAS 支持 DROP 规则吗？

**不支持。** 阿里云轻量云的 `CreateFirewallRules` API 无 Policy 字段，创建的规则均为 accept。配置 DROP 时会记录 WARN 日志并跳过。

### 7. WebUI 模式如何切换为 .env 模式？

设置环境变量 `FWALIZER_MODE=env`，或确保 `TARGETS` 环境变量存在（程序会自动检测并进入 .env 模式）。

---

## 许可证

[MIT License](./LICENSE)
# 🔐 FWAlizer — Firewall DNS Synchronizer

**FWAlizer** 是一个轻量级自动化工具，运行在 Docker 容器中，通过 DNS 解析指定域名的 IP 地址，并自动将解析结果同步到**腾讯云 Lighthouse 防火墙**白名单中。

> 适用于：域名 IP 频繁变动的场景（如动态 DNS、API 网关），避免手动更新防火墙规则。

---

## ✨ 特性

- 🚀 **自动同步**：定时 DNS 解析 → 自动更新防火墙白名单
- 📦 **Docker 一键部署**：`docker run -d --env-file .env ghcr.io/alcaprophet/fwalizer:latest`
- 🔒 **安全隔离**：仅管理标记为 `[auto-dns]` 的规则，不影响其他防火墙规则
- 🌐 **IPv4 + IPv6**：同时支持 A 记录和 AAAA 记录
- 🛡️ **乐观锁**：自动处理版本冲突，最多重试 3 次
- ⏱️ **频率保护**：内置 500ms API 调用间隔，遵守腾讯云 10次/秒 限制
- 📋 **灵活配置**：支持 TCP/UDP、多端口、ACCEPT/DROP 动作
- 🏥 **健康检查**：内置 Docker HEALTHCHECK

---

## 🚀 快速开始

### 1. 准备配置文件

```bash
cp .env.example .env
# 编辑 .env 填入你的配置
```

`.env` 内容示例：

```env
TENCENTCLOUD_SECRET_ID=AKIDxxxxxxxx
TENCENTCLOUD_SECRET_KEY=xxxxxxxxxxxxxxxx
LIGHTHOUSE_INSTANCE_ID=lhins-xxxxxxxx
LIGHTHOUSE_REGION=ap-guangzhou
DOMAIN_RULES=api.example.com|TCP|443,80|ACCEPT;cdn.example.com|TCP|443|ACCEPT
CHECK_INTERVAL=5m
```

### 2. 拉取并运行

```bash
docker run -d \
  --name fwalizer \
  --env-file .env \
  --restart=always \
  ghcr.io/alcaprophet/fwalizer:latest
```

### 3. 查看日志

```bash
docker logs -f fwalizer
```

---

## ⚙️ 配置说明

| 环境变量 | 必填 | 默认值 | 说明 |
|---------|------|--------|------|
| `TENCENTCLOUD_SECRET_ID` | ✅ | - | 腾讯云 SecretId |
| `TENCENTCLOUD_SECRET_KEY` | ✅ | - | 腾讯云 SecretKey |
| `LIGHTHOUSE_INSTANCE_ID` | ✅ | - | Lighthouse 实例 ID |
| `LIGHTHOUSE_REGION` | ✅ | - | 实例地域（如 `ap-guangzhou`） |
| `DOMAIN_RULES` | ✅ | - | 域名规则，见下方格式 |
| `RULE_TAG` | ❌ | `auto-dns` | 规则描述前缀 |
| `CHECK_INTERVAL` | ❌ | `5m` | 检查间隔（`30s`/`1m`/`5m`/`1h`） |
| `DNS_SERVER` | ❌ | `8.8.8.8:53` | 上游 DNS 服务器 |

### DOMAIN_RULES 格式

```
host|protocol|ports|action[|comment];host|protocol|ports|action[|comment];...
```

| 字段 | 可选值 | 说明 |
|------|--------|------|
| `host` | 域名 | 如 `api.example.com` |
| `protocol` | `TCP` / `UDP` / `TCP+UDP` | 协议类型 |
| `ports` | `443,80` / `ALL` | 端口号（逗号分隔）或 `ALL` 全部 |
| `action` | `ACCEPT` / `DROP` | 匹配动作 |
| `comment` | 任意文本 | 可选备注，直接拼接在 `[RULE_TAG]` 后 |

> ⚠️ **注意**：
> - 请确保每个域名仅指向单台服务器（少量 IP），本工具**不支持返回大量 IP 的 CDN 域名**。
> - 防火墙规则描述（`FirewallRuleDescription`）受腾讯云 API 限制 ≤ 64 字节，请勿使用超长域名或备注。
> - 多个域名之间已内置 500ms API 频率保护间隔，遵守腾讯云 10次/秒 速率限制。

生成的防火墙规则描述格式：`[RULE_TAG]备注内容`，如 `[auto-dns]生产API`。无备注时仅保留 `[auto-dns]`。

示例：
```env
# 放行 api.example.com 的 TCP 443 和 80 → 描述: [auto-dns]生产API
# 放行 cdn.example.com 的 TCP+UDP 443 → 描述: [auto-dns]
DOMAIN_RULES=api.example.com|TCP|443,80|ACCEPT|生产API;cdn.example.com|TCP+UDP|443|ACCEPT
```

---

## 🏗️ 本地开发

### 无需 Docker 直接运行

```powershell
# 1. 创建并编辑 .env
copy .env.example .env

# 2. 加载环境变量（注意 -Encoding UTF8 避免中文乱码）
Get-Content -Encoding UTF8 .env | Where-Object { $_ -match '^\s*[^#]' -and $_ -match '=' } | ForEach-Object {
    $name, $value = $_ -split '=', 2
    Set-Item -Path "env:$($name.Trim())" -Value $value.Trim()
}

# 3. 运行
go run .
```

### Make 命令（需要 Git Bash 或 WSL2）

```bash
# 编译
make build

# 本地运行
make run

# 构建 Docker 镜像
make docker-build

# Docker 本地运行
make docker-run

# 查看日志
make docker-logs

# 运行测试
make test
```

### Go 命令

```bash
go build -ldflags="-s -w" -o fwalizer .   # 编译
go test -v ./...                            # 测试
go vet ./...                                # 代码检查
```

---

## 📁 项目结构

```
TencentCloudFirewallTool/
├── .env.example                 # 配置模板
├── Dockerfile                   # 多阶段构建
├── Makefile                     # 开发命令
├── main.go                      # 入口
├── config/
│   └── config.go                # .env 解析与校验
├── dns/
│   └── resolver.go              # DNS 解析（自定义服务器）
├── firewall/
│   ├── client.go                # Lighthouse SDK 封装
│   ├── rule.go                  # 规则对比 & diff 逻辑
│   └── sync.go                  # 定时同步主循环
├── TencentAPIGuide/             # 腾讯云 Lighthouse API & Go SDK 官方文档
├── .github/workflows/
│   └── docker-publish.yml       # CI/CD（自动构建 + 推送 ghcr.io）
└── Ref/                         # API 文档 & SDK 参考
```

---

## 🔒 安全建议

1. **使用 CAM 子账号**：不要使用主账号密钥，创建子账号并授予最小权限（`QcloudLighthouseFullAccess`）
2. **不提交 `.env`**：`.gitignore` 已配置忽略，密钥不会进入版本控制
3. **限制 IP 访问**：在 CAM 中配置 IP 限制策略
4. **定期轮换密钥**

---

## 📄 License

MIT © 2024
