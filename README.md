# 🔐 FWAlizer — Firewall DNS Synchronizer

**FWAlizer** 是一个轻量级自动化工具，运行在 Docker 容器中，通过 DNS 解析指定域名的 IP 地址，并自动将解析结果同步到**腾讯云 Lighthouse 防火墙**白名单中。

> 适用于：域名 IP 频繁变动的场景（如动态 DNS、API 网关），避免手动更新防火墙规则。

---

## ✨ 特性

- 🚀 **自动同步**：定时 DNS 解析 → 自动更新防火墙白名单
- 📦 **Docker 一键部署**：`docker run -d --env-file .env ghcr.io/alcaprophet/fwalizer:latest`
- 🔒 **安全隔离**：仅管理标记为 `[auto-dns:xxx]` 的规则，不影响其他防火墙规则
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
| `comment` | 任意文本 | 可选备注，写入 `FirewallRuleDescription` |

> ⚠️ **注意**：
> - 请确保每个域名仅指向单台服务器（少量 IP），本工具**不支持返回大量 IP 的 CDN 域名**。
> - 防火墙规则描述（`FirewallRuleDescription`）受腾讯云 API 限制 ≤ 64 字节，请勿使用超长域名或备注。
> - 多个域名之间已内置 500ms API 频率保护间隔，遵守腾讯云 10次/秒 速率限制。

示例：
```env
# 放行 api.example.com 的 TCP 443 和 80（备注：生产API）
# 放行 cdn.example.com 的 TCP+UDP 443
DOMAIN_RULES=api.example.com|TCP|443,80|ACCEPT|生产API;cdn.example.com|TCP+UDP|443|ACCEPT
```

---

## 🏗️ 本地开发

> **Windows 用户注意**：Makefile 中的命令（`rm`、`docker` 等）需要 **WSL2** 或 **Git Bash** 环境运行。也可直接用 Go 命令替代：
> ```powershell
> go build -ldflags="-s -w" -o fwalizer.exe .
> docker build -t fwalizer .
> ```

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

# 测试
make test
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
