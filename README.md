# FWAlizer — 防火墙 DNS 自动同步工具

**FWAlizer**（Firewall DNS Synchronizer）是一个轻量级自动化工具：定时解析指定域名的 IP 地址，自动同步到云防火墙/安全组白名单中。专为域名 IP 频繁变动的场景设计（如动态 DNS、API 网关、VPN 入口）。

---

## 核心特性

- **多云支持**：腾讯云 Lighthouse / CVM，阿里云轻量云（SWAS）/ ECS，四款云产品统一管控
- **双模式运行**：`.env` 文件驱动（适合服务器/容器）或 WebUI 可视化管理（适合桌面端）
- **增量同步**：仅操作带 `[TAG]` 标记的规则，绝不覆盖手动配置的防火墙规则
- **DNS 熔断保护**：连续解析失败达阈值后自动熔断，半开探测自动恢复，避免误删规则
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
| Node.js | 20+（可选） | 前端开发/重新构建时需要 |
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
| `DNS` | `223.5.5.5` | 上游 DNS 服务器地址（端口 :53 自动补全） |
| `DNS_TIMEOUT` | `10s` | DNS 解析超时时间（WebUI 模式下在「全局设置」页面可配置） |
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
| `WEBUI_PORT` | `60200` | WebUI 监听端口（绑定 127.0.0.1，若被占用自动在 50000–65535 范围随机选择可用端口） |
| `FWALIZER_DATA_DIR` | 各平台标准路径 | WebUI 数据存储目录（SQLite 数据库位置） |

---

## 运行模式

### .env 模式（服务器/容器部署）

当检测到 `TARGETS` 环境变量或 `.env` 文件时自动进入。纯 headless 运行，日志输出到 stdout。

```bash
./fwalizer                    # 从 .env 加载配置并启动同步
./fwalizer validate .env      # 仅校验配置，不启动
./fwalizer version            # 显示版本号
./fwalizer backup             # 备份 WebUI 数据库
./fwalizer restore <文件>     # 从备份恢复数据库
```

### WebUI 模式（桌面/可视化管理）

无 `.env` 文件时自动进入 WebUI 模式，配置存储在 SQLite 数据库中。

- 默认地址：`http://127.0.0.1:60200`（端口被占用时自动选择可用端口）
- 数据路径（自动选择）：
  - macOS：`~/Library/Application Support/fwalizer/config.db`
  - Linux：`~/.config/fwalizer/config.db`
  - Windows：`%APPDATA%\fwalizer\config.db`
- 可通过 `FWALIZER_DATA_DIR` 环境变量自定义数据目录
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
| targets | 目标编号（从 0 开始，留空或 `*` = 全部） | `0,2` |
| comment | 可选备注 | `生产API` |

### 示例

```env
# 单域名 + 多端口
RULES=api.example.com|TCP|443,80|ACCEPT||生产API

# 指定目标（仅第 2 个 TARGETS 条目，编号从 0 开始）
RULES=vpn.example.com|UDP|1194|ACCEPT|1|VPN接入

# 端口范围 + 多目标
RULES=game.example.com|TCP|8000-8010|ACCEPT|0,2|游戏端口

# ICMP（Ping），端口自动设为 ALL
RULES=ping.example.com|ICMP|ALL|ACCEPT||允许Ping

# TCP+UDP（仅阿里云 SWAS 原生支持，其他云自动拆分为两条规则）
RULES=voice.example.com|TCP+UDP|5060|ACCEPT||SIP语音

# 多条规则组合（逗号分隔 + 反斜杠换行）
RULES=api.example.com|TCP|443|ACCEPT||API, \
      vpn.example.com|UDP|1194|ACCEPT|1|VPN
```

> **注意**：仅支持单台服务器场景（DNS 返回少量 IP），不支持 CDN 等返回大量 IP 的域名。

---

## CLI 数据备份与恢复

WebUI 模式下所有配置存储在 SQLite 数据库中，可通过 CLI 命令备份和恢复。**WebUI 模式启动时自动检测 pidfile，防止重复运行多个实例**（若已有实例运行会报错并退出）。

```bash
# 备份（自动生成时间戳文件名，保留最新 5 个）
./fwalizer backup

# 恢复（需先停止运行中的 FWAlizer）
./fwalizer restore config.db.bak.20260727_120000
```

> 备份和恢复仅适用于 WebUI 模式；`.env` 模式直接复制 `.env` 文件即可。

---

## 桌面端系统托盘

在 macOS 或 Windows 上使用 `-tags desktop` 编译后，启动 WebUI 模式会自动显示系统托盘图标：

- 托盘菜单：状态指示 / 打开配置面板 / 立即同步 / 开机自启 / 退出
- 启动后自动打开浏览器进入 WebUI
- 点击「退出」会等待当前同步轮次完成后再退出进程

```bash
CGO_ENABLED=1 go build -tags desktop -o fwalizer .
./fwalizer
```

> macOS 开机自启通过 LaunchAgent plist 实现，Windows 通过注册表 Run 键实现。

---

## 告警通知

在 WebUI 的「告警配置」页面中，可配置邮件（SMTP）和 Webhook 两种通知方式。启用后在发生同步错误或 DNS 解析失败时自动推送告警：

- **邮件告警**：支持标准 SMTP（如 QQ 邮箱、163 邮箱、企业邮箱）
- **Webhook 告警**：支持钉钉、飞书、Slack 三种渠道（在告警配置页选择「通知渠道」），自动适配各平台消息格式
- 告警配置修改后即时生效（热重载），无需重启

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
  -p 60200:60200 \
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
    #   - "60200:60200"
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

# 启动开发服务器（热重载，代理 API 到 127.0.0.1:60200）
npm run dev

# 构建生产版本（输出到 dist/，go:embed 会自动包含）
npm run build
```

### 构建完整二进制（含前端）

```bash
cd webui/frontend && npm run build && cd ../..
make build
./fwalizer    # 访问 http://127.0.0.1:60200
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

**不会。** DNS 解析失败时保留现有规则不变（仅记录 WARN 日志）。连续失败达到阈值（默认 5 次）后触发熔断，暂停该域名同步。**熔断后每轮同步仍会尝试一次半开探测**，解析成功后自动解除熔断并恢复正常同步，无需重启。

### 5. 支持 IPv6 吗？

腾讯云 Lighthouse、CVM 和阿里云 ECS 支持 IPv6（AAAA 记录）。阿里云轻量云（SWAS）不支持 IPv6，解析到的 IPv6 地址会自动跳过。

### 6. 阿里云 SWAS 支持 DROP 规则吗？

**不支持。** 阿里云轻量云的 `CreateFirewallRules` API 无 Policy 字段，创建的规则均为 accept。配置 DROP 时会记录 WARN 日志并跳过。

### 7. WebUI 模式如何切换为 .env 模式？

设置环境变量 `FWALIZER_MODE=env`，或确保 `TARGETS` 环境变量存在（程序会自动检测并进入 .env 模式）。

### 8. 如何备份和恢复 WebUI 配置？

使用 `./fwalizer backup` 备份 SQLite 数据库，`./fwalizer restore <文件>` 恢复。恢复前需先停止 FWAlizer 进程。

### 9. 桌面端托盘不出现？

桌面端功能需要通过 `-tags desktop` 编译标签启用（需 CGO）。标准 `go build` 和 Docker 镜像不包含托盘功能。

### 10. 如何配置告警通知？

启动 WebUI 模式后，在左侧菜单进入「告警配置」页面，填写 SMTP 或 Webhook 信息并启用即可。Webhook 支持在配置页选择「通知渠道」（钉钉/飞书/Slack），程序会自动适配各平台的消息格式。配置保存后即时生效。

### 11. 后端服务端口被占用怎么办？

默认端口 `60200` 被占用时，程序会自动在 `50000–65535` 范围内随机选择一个可用端口，并在日志中输出 WARN 提示和实际端口号。您也可以显式设置 `WEBUI_PORT` 环境变量指定其他端口。

### 12. WebUI 模式能否同时启动多个实例？

**不能。** 程序通过 pidfile（`<数据目录>/fwalizer.pid`）检测已有实例，若检测到另一个 FWAlizer 进程正在运行，会拒绝启动并提示 PID。这避免了多实例操作同一 SQLite 数据库可能引起的问题。

---

## 许可证

[MIT License](./LICENSE)
