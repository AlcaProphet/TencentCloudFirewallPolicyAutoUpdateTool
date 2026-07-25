# FWAlizer 设计大纲（Design1）

> 面向人类的架构设计与决策文档。技术实现细节见 [Build1.md](./Build1.md)，AI 编码指令见 [AGENTS.md](./AGENTS.md)。

---

## 一、项目定位

**FWAlizer**（Firewall DNS Synchronizer）是一个轻量级自动化工具：通过 DNS 解析指定域名的 IP 地址，自动同步到云防火墙/安全组白名单中。

核心特性：**跨平台单二进制**、**多云支持**、**WebUI 配置面板**、**Docker 与桌面端双部署**。

---

## 二、需求总览

| 序号 | 需求 | 说明 |
|------|------|------|
| 1 | 多实例多地域 | 同时管理多台云资源，跨地域 |
| 2 | **多云支持** | 腾讯云 Lighthouse / CVM，阿里云 ECS / 轻量云 |
| 3 | WebUI 管理面板 | 内部使用，简化配置，不对外开放 |
| 4 | 跨平台桌面端 | Linux / Windows / macOS |
| 5 | 轻量化 | 低资源占用，单二进制分发 |
| 6 | 可扩展 | 邮件告警、Webhook 等未来功能 |

---

## 三、语言选型

**结论：继续使用 Go。**

核心理由：
1. 现有代码 100% Go，零重写成本
2. 原生交叉编译，单二进制跨平台
3. `embed` 包将前端静态资源编译进二进制
4. 不引入 Node.js/Python 运行时，保持轻量
5. 桌面端方案成熟（系统托盘 + 浏览器 WebUI）
6. 腾讯云和阿里云均提供官方 Go SDK

---

## 四、运行模式设计

程序自动检测运行模式，决定配置来源：

| 运行环境 | 配置来源 | WebUI | 说明 |
|---------|---------|-------|------|
| Docker + 传入 .env | .env 解析 | 不启动 | 纯 headless |
| Docker + 未传 .env | SQLite | 启动 | 通过 WebUI 配置 |
| 桌面端直接运行 | SQLite | 启动 | 系统托盘 + WebUI |

### 模式判定逻辑

- 检测 `TARGETS` 环境变量是否存在 → `.env` 模式
- 否则进入 WebUI 模式（自动区分 Docker 或桌面端）
- 可通过 `FWALIZER_MODE=env|webui` 强制覆盖

### 关键设计决策

- `.env` 模式与 WebUI 模式**互斥**：前者不写 SQLite，后者不读 `.env`
- `.env` 模式定位为"高级/运维模式"
- WebUI 默认绑定 `127.0.0.1`，仅本机访问
- 端口通过 `WEBUI_PORT` 配置，默认 `9090`

### 配置存储双轨制

```
EnvLoader (.env 解析) ──→ Config ──→ Syncer
StoreLoader (SQLite)  ──→ Config ──→ Syncer（支持热重载）
```

---

## 五、多端部署策略

### 5.1 Docker

- `.env` 模式：纯 headless，stdout 日志
- WebUI 模式：暴露端口，SQLite 挂载 `/data`
- HEALTHCHECK：WebUI 模式用 HTTP 端点检测

### 5.2 桌面端

**方案：系统托盘 + 浏览器 WebUI（推荐）**

理由：
- 符合"后台常驻服务"定位，无需独立窗口
- 系统托盘提供右键菜单（打开配置面板、立即同步、查看日志、退出）
- 浏览器 WebUI 与 Docker WebUI 零额外开发成本
- 开机自启（Windows Startup / macOS Login Items）
- 体积小（~12MB）

托盘菜单交互：
```
● 运行中 / ● 同步中 / ● 异常
├── 打开配置面板
├── 立即同步
├── 开机自启 [开关]
└── 退出
```

**开机自启**：仅 Windows 和 macOS 支持，作为用户可选设置项（WebUI「全局设置」中开关），**默认不启用**。
- Windows：写入注册表 `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
- macOS：生成 `~/Library/LaunchAgents/com.fwalizer.agent.plist`
- Linux：不提供内置支持（用户自行配置 systemd user unit）

**CGO 说明：** 托盘库 `fyne.io/systray` 在 macOS 需要 CGO，桌面端构建 `CGO_ENABLED=1`，与 Docker 的纯静态编译（`CGO_ENABLED=0`）通过 Go build tags 分离。

---

## 六、多云支持架构

### 6.1 支持的云产品

| 云厂商 | 产品 | 防火墙机制 | 操作对象 |
|--------|------|-----------|----------|
| 腾讯云 | Lighthouse 轻量云 | 实例级防火墙 | `InstanceId` |
| 腾讯云 | CVM 云服务器 | 安全组 | `SecurityGroupId` |
| 阿里云 | 轻量应用服务器 | 实例级防火墙 | `InstanceId` + `RegionId` |
| 阿里云 | ECS 云服务器 | 安全组 | `SecurityGroupId` + `RegionId` |

### 6.2 核心差异

- **实例级 vs 安全组级**：实例级防火墙规则直接绑定实例；安全组可关联多实例
- **端口格式**：各云不一致（逗号分隔 / 范围 / 斜杠格式），需统一转换
- **频率限制**：腾讯云 Lighthouse 10次/秒、CVM 创建50次/秒，阿里云轻量云 100次/60秒，差异显著
- **规则数量**：腾讯云 CVM 单安全组上限 100 条，需检查后写入
- **规则标识**：均通过描述字段 + `[TAG]` 前缀识别本工具创建的规则
- **IPv6**：腾讯云独立 `Ipv6CidrBlock` 字段，阿里云 ECS 独立 `Ipv6SourceCidrIp`，阿里云轻量云不支持 IPv6

### 6.3 Provider 抽象设计

核心思想：定义统一 `Provider` 接口，各云厂商实现该接口。Syncer 仅依赖接口，不感知具体云厂商。

```
Syncer ──→ Provider 接口
              ├── TencentLighthouseProvider
              ├── TencentCVMProvider
              ├── AliyunSWASProvider
              └── AliyunECSProvider
```

工厂模式 + 自注册：各 Provider 在 `init()` 中注册，新增云厂商仅需实现接口。

### 6.4 同步并发策略

- **跨云厂商并行**：不同云厂商 API 配额独立
- **同云厂商串行**：共享 API 配额，各域名间加间隔
- ClientPool 连接池：相同云厂商 + Region + 凭据的 Target 共享 SDK Client

---

## 七、WebUI 功能规划

### 技术选型

| 层 | 技术 | 理由 |
|---|------|------|
| 后端 | Go `net/http` + `encoding/json` | 标准库，无外部依赖 |
| 前端 | **Vue 3** | 轻量，CDN 引入 |
| 嵌入 | Go `embed` | 编译进二进制 |
| 存储 | **SQLite** (`modernc.org/sqlite`) | 纯 Go、单文件 |

### 功能页面

```
├── 仪表盘（实例状态概览、规则统计）
├── 云资源管理（添加/编辑/删除、连接测试）
├── 域名规则管理（可视化编辑、绑定云资源、DNS 测试）
├── 全局设置（DNS、间隔、TAG）
├── 同步日志（最近 N 条记录）
├── 高级功能（Dry Run、配置导入/导出、健康检查）
└── 告警配置（邮件、Webhook）
```

---

## 八、可扩展性设计

### 事件总线（Event Bus）

解耦同步事件与消费者：
```
EventBus ──→ Subscriber
  ├── sync:start / sync:complete / sync:error
  ├── rule:changed
  └── dns:failed
```

### 扩展插件

```
notifier/
├── email.go    （SMTP 邮件告警）
├── webhook.go  （钉钉/飞书/Slack）
└── telegram.go（Telegram Bot）
```

---

## 九、实施路线

| Phase | 内容 | 周期 |
|-------|------|------|
| **Phase 1** | Provider 抽象 + 多云基础 + 配置重设计 + CLI | 1-2 周 |
| **Phase 2** | WebUI + 配置持久化 + 热重载 | 1-2 周 |
| **Phase 3** | 告警 + 高级功能（EventBus、Dry Run、导入导出） | 1-2 周 |
| **Phase 4** | 桌面端 + 打磨（系统托盘、打包、日志持久化） | 1-2 周 |

---

## 十、技术决策速览

| 决策点 | 选择 |
|--------|------|
| 后端语言 | Go（维持） |
| 配置默认模式 | WebUI + SQLite 优先，.env 为高级模式 |
| .env 模式触发 | 检测 `TARGETS` 环境变量 |
| 多云抽象 | Provider 接口 + 工厂注册 |
| 统一规则模型 | RuleInfo / RuleAction |
| Web 后端 | `net/http` 标准库 |
| Web 前端 | Vue 3 |
| 配置存储 | SQLite（`modernc.org/sqlite`） |
| 桌面方案 | 系统托盘 + 浏览器 WebUI |
| 系统托盘库 | `fyne.io/systray` |
| 构建分离 | Go build tags + CGO 差异 |
| 开机自启 | Windows + macOS，用户设置项，默认关闭 |
| 配置热重载 | channel 通知 Syncer |
| 健康检查 | `/api/health` HTTP 端点 |
| 扩展架构 | Event Bus + Subscriber |
| WebUI 安全 | 默认绑定 `127.0.0.1` |
| WebUI 端口 | 可配置，默认 `9090` |
| 凭据管理 | 按云厂商独立环境变量 |
| 同步并发策略 | 跨厂商并行，同厂商串行 |
| 幂等设计 | 删除时“规则不存在”视为成功，添加时“规则已存在”跳过 |
| 规则方向 | 仅操作入站（Ingress）规则 |
| 桌面存储路径 | 系统标准路径 |
| 协议支持 | TCP / UDP / TCP+UDP / ICMP |
| 端口输入格式 | 单端口、逗号分隔、范围（`8000-8010`）、`ALL` |
| Dry Run | 执行到 Diff 为止，不实际写入 |
| 配置导入/导出 | JSON 格式 |
| SQLite 并发 | WAL 模式 + 短事务 |
| 进程锁 | WebUI 模式 pidfile 防多实例 |
| EventBus 投递 | 异步，Subscriber 失败仅记日志 |
| 版本号注入 | ldflags `-X github.com/alcaprophet/fwalizer/version.Version` |
| DNS 失败处理 | 渐进式熔断（半开探测恢复） |
| 日志级别 | `LOG_LEVEL` 环境变量 |

---

## 十一、设计原则

本项目遵循以下核心原则：

1. **简单轻量化**：功能做减法，不引入重型框架，优先标准库
2. **内部使用导向**：安全设计以内部使用为主，WebUI 不针对公开访问设计，网络安全边界由用户控制
3. **开箱即用**：最小化前置依赖和配置步骤，首次运行即可工作
4. **不过度防御**：避免过度防御性编程，聚焦核心场景
