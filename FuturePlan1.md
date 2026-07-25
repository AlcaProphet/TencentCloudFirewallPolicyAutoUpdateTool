# FWAlizer 未来演进方案（FuturePlan1）

## 一、需求总结

| 序号 | 需求 | 说明 |
|------|------|------|
| 1 | 多实例多地域支持 | 同时管理多台实例，跨地域 |
| 2 | **多云支持** | 腾讯云 Lighthouse / CVM，阿里云 ECS / 轻量云 |
| 3 | WebUI 管理面板 | 内部使用，简化配置难度，不对外开放 |
| 4 | 跨平台桌面端 | 同时支持 Linux / Windows / macOS |
| 5 | 轻量化 | 保持低资源占用，单二进制分发 |
| 6 | 可扩展性 | 支持邮件告警、Webhook 等未来功能扩展 |

---

## 二、语言选型分析

### 候选语言对比

| 语言 | 跨平台 | 单二进制 | 桌面 GUI | 生态丰富度 | 轻量性 | 与现有 Go 代码兼容性 |
|------|--------|----------|----------|-----------|--------|---------------------|
| **Go** | ★★★★★ | ★★★★★ | ★★★☆ (Wails) | ★★★★ | ★★★★★ | ★★★★★ |
| Python | ★★★★ | ★☆ (PyInstaller) | ★★★★ (Tkinter/Qt) | ★★★★★ | ★★ | ★☆ |
| Node.js | ★★★★ | ★☆ (pkg/nexe) | ★★★★★ (Electron) | ★★★★★ | ★★ | ★☆ |
| C++ | ★★★★★ | ★★★★★ | ★★★★★ (Qt) | ★★★ | ★★★★ | ☆ |
| PHP | ★★★ | ★☆ | ★☆ | ★★★ | ★★★ | ☆ |

### 推荐方案：继续使用 Go

**理由：**

1. **零改造成本**：现有后端代码 100% Go 编写，无需重写核心逻辑（DNS 解析、防火墙同步、重试机制）
2. **单二进制跨平台**：Go 原生交叉编译 `GOOS=linux/windows/darwin`，一个二进制包含后端 + WebUI 服务
3. **嵌入式 WebUI**：Go 1.16+ 的 `embed` 包可将前端静态资源直接编译进二进制，无需额外文件
4. **轻量级**：不引入 Node.js/Python 运行时，Docker 镜像保持 ~14MB
5. **桌面端方案成熟**：[Wails](https://wails.io)（Go + Web 前端）可打包为 macOS/Windows/Linux 桌面应用，体感接近 Electron 但体积仅 ~10MB（Electron ~150MB）
6. **多云 SDK 齐全**：腾讯云和阿里云均提供官方 Go SDK，无需第三方封装

---

## 三、配置模式重设计（WebUI 优先）

### 核心思路

程序启动时自动检测运行模式，决定配置来源：

| 运行环境 | 配置来源 | WebUI | 说明 |
|---------|---------|-------|------|
| Docker + 传入 .env | .env 解析 | 不启动 | 纯 headless，适配现有 Docker 部署 |
| Docker + 未传 .env | SQLite | 启动 | 首次通过 WebUI 配置，之后持久化 |
| 桌面端直接运行 | SQLite | 启动 | 系统托盘 + WebUI 配置面板 |

### 启动判定逻辑（伪代码）

```go
func determineMode() RunMode {
    // 检测是否有最小必要环境变量（至少一个 TARGET 配置）
    if hasMinimalEnvConfig() {
        return ModeEnvOnly  // 纯 .env 模式，不启动 WebUI
    }
    // 否则进入 WebUI 模式
    if isDocker() {
        return ModeDockerWebUI  // Docker 内启动 WebUI
    }
    return ModeDesktopWebUI     // 桌面端启动 WebUI + 系统托盘
}
```

### 关键设计点

- `.env` 模式标记为"高级/运维模式"，仅当用户主动传入 `TARGETS` 环境变量时激活
- WebUI 模式下，配置持久化到 SQLite（`modernc.org/sqlite`，纯 Go，无 CGO）
- 两种模式互斥：`.env` 模式不写 SQLite，WebUI 模式不读 `.env`（避免配置混乱）
- 可通过 `FWALIZER_MODE=env|webui` 环境变量强制指定模式（覆盖自动检测）
- WebUI 默认监听 `127.0.0.1`（仅本机访问），Docker 用户通过 `-p` 映射自行决定暴露范围
- WebUI 端口可通过 `WEBUI_PORT` 环境变量配置，默认 `9090`

### WebUI 安全策略

- **默认绑定 `127.0.0.1`**：仅允许本机访问，不暴露到局域网
- Docker 部署时，用户通过 `-p 9090:9090` 或 `-p 127.0.0.1:9090:9090` 自行控制暴露范围
- 如需局域网访问，可通过 `-p 0.0.0.0:9090:9090` 显式暴露（用户自行评估风险）
- WebUI 持有云密钥和防火墙控制权，安全优先级高于便利性

### 首次运行体验（桌面端）

- 启动后自动打开浏览器访问 `http://localhost:{PORT}`
- 系统托盘显示提示气泡："FWAlizer 已启动，点击打开配置面板"
- 无需引导向导，用户直接通过 WebUI 表单添加云资源和规则
- SQLite 数据文件存储在各平台标准路径：
  - macOS: `~/Library/Application Support/fwalizer/config.db`
  - Windows: `%APPDATA%/fwalizer/config.db`
  - Linux: `~/.config/fwalizer/config.db`

### 配置存储双轨制

```
+------------------+       +------------------+
|   EnvLoader      |       |   StoreLoader    |
| (解析 .env)       |       | (读取 SQLite)    |
+--------+---------+       +--------+---------+
         |                           |
         v                           v
    +----+----+                 +----+----+
    | Config  |                 | Config  |
    +---------+                 +---------+
              \               /
               v             v
               +----+----+
               | Syncer  |
               +---------+
```

- `EnvLoader`：仅解析 `TARGETS` / `RULES` / `TAG` / `INTERVAL` / `DNS`，生成只读 Config
- `StoreLoader`：从 SQLite 加载 Config，支持热更新（WebUI 修改后通知 Syncer 重载）

---

## 四、多端运行策略

### 4.1 Docker 部署

```
docker run -d --env-file .env ghcr.io/alcaprophet/fwalizer:latest   # .env 模式
docker run -d -p 9090:9090 -v fwalizer-data:/data ghcr.io/...       # WebUI 模式
```

- `.env` 模式：纯 headless 后台进程，stdout 日志
- WebUI 模式：暴露端口（默认 9090），SQLite 挂载到 `/data` 持久化
- HEALTHCHECK：WebUI 模式下使用 `wget -q --spider http://localhost:9090/api/health`，比进程检测更准确

### 4.2 桌面端（Windows / macOS）：系统托盘 + 浏览器 WebUI

**方案对比：**

| 方案 | 优点 | 缺点 | 适合场景 |
|------|------|------|---------|
| **A: 纯浏览器 WebUI** | 开发简单，一套代码 | 用户需手动启动进程、记住端口、浏览器标签页管理 | 开发者自用 |
| **B: Wails 原生窗口** | 原生体验，独立窗口 | 构建复杂，二进制较大(~15MB)，窗口管理增加认知负担 | 需要完整桌面 App |
| **C: 系统托盘 + 浏览器 WebUI** | 静默运行、托盘图标管理、浏览器配置，体感最自然 | 需引入系统托盘库 | **后台工具类软件** |

**推荐方案 C**，理由：

1. **符合"后台小工具"定位**：防火墙同步本质是后台常驻服务，不需要独立窗口
2. **系统托盘是最佳交互入口**：右键菜单提供"打开配置面板"、"立即同步"、"查看日志"、"退出"等操作
3. **浏览器 WebUI 零额外开发成本**：后端 HTTP Server 与 Docker WebUI 完全复用
4. **开机自启**：托盘应用注册为系统启动项（Windows: Startup, macOS: Login Items）
5. **体积小**：Go 二进制 + embed 前端，约 10-12MB

### 4.3 系统托盘交互设计

```
托盘图标 (systray)
├── 状态：● 运行中 / ● 同步中 / ● 异常
├── 打开配置面板 (http://localhost:9090)
├── 立即同步
├── 查看最近日志
├── 开机自启 [开关]
└── 退出
```

技术实现：使用 `fyne.io/systray`（跨平台系统托盘库），约 2MB 开销。

> **⚠️ CGO 依赖说明**：`fyne.io/systray` 在 macOS 上需要 Objective-C 编译器（AppKit），Linux 上需要 `libgtk-3-dev`。
> 桌面端构建**必须** `CGO_ENABLED=1`，与 Docker 的纯静态编译（`CGO_ENABLED=0`）分离。详见 4.4 节。

### 4.4 构建标签分离

使用 Go build tags 控制编译产物，Docker 镜像不包含系统托盘依赖：

```go
// systray.go
//go:build desktop

package app
// 系统托盘相关代码仅在 desktop 构建时包含
```

```bash
# Docker 构建（无系统托盘，纯静态）
CGO_ENABLED=0 go build -tags docker -ldflags="-s -w -X main.version=$VERSION" -o fwalizer .

# 桌面构建（含系统托盘，需要 CGO）
CGO_ENABLED=1 go build -tags desktop -ldflags="-s -w -X main.version=$VERSION" -o fwalizer .
```

**构建差异对比：**

| 维度 | Docker 构建 | 桌面构建 |
|------|-------------|----------|
| CGO | `CGO_ENABLED=0`（纯静态） | `CGO_ENABLED=1`（需要系统托盘原生库） |
| Build Tags | `-tags docker` | `-tags desktop` |
| 包含模块 | core + webui + provider | core + webui + provider + systray |
| 体积 | ~14MB（Docker 镜像内） | ~12-14MB（单二进制） |
| 交叉编译 | 仅 `linux/amd64` | 需各平台工具链（macOS: xcode, Windows: mingw） |

---

## 五、多云支持架构（核心设计）

### 5.1 支持的云产品矩阵

| 云厂商 | 产品 | 防火墙机制 | 资源标识 | Go SDK 模块 |
|--------|------|-----------|----------|-------------|
| 腾讯云 | **Lighthouse 轻量云** | 实例级防火墙 | `InstanceId` (lhins-xxx) | `tencentcloud-sdk-go/tencentcloud/lighthouse` |
| 腾讯云 | **CVM 云服务器** | 安全组 (Security Group) | `SecurityGroupId` (sg-xxx) | `tencentcloud-sdk-go/tencentcloud/vpc` |
| 阿里云 | **轻量应用服务器** | 实例级防火墙 | `InstanceId` + `RegionId` | `alibaba-cloud-sdk-go/services/swas-open` 或 V2 SDK |
| 阿里云 | **ECS 云服务器** | 安全组 (Security Group) | `SecurityGroupId` (sg-xxx) + `RegionId` | `alibaba-cloud-sdk-go/services/ecs` |

### 5.2 各云 API 对比

| 操作 | 腾讯云 Lighthouse | 腾讯云 CVM (安全组) | 阿里云轻量云 | 阿里云 ECS (安全组) |
|------|-------------------|---------------------|-------------|---------------------|
| **查询规则** | `DescribeFirewallRules` | `DescribeSecurityGroupPolicies` | `ListFirewallRules` | `DescribeSecurityGroupAttribute` |
| **添加规则** | `CreateFirewallRules` | `CreateSecurityGroupPolicies` | `CreateFirewallRules` | `AuthorizeSecurityGroup` |
| **删除规则** | `DeleteFirewallRules` | `DeleteSecurityGroupPolicies` | `DeleteFirewallRules` | `RevokeSecurityGroup` |
| **Endpoint** | `lighthouse.tencentcloudapi.com` | `vpc.tencentcloudapi.com` | `swas.{regionId}.aliyuncs.com` | `ecs.{regionId}.aliyuncs.com` |
| **频率限制** | 10次/秒 | 100次/秒 | 100次/60秒 | 不限（按产品整体） |
| **规则标识** | `FirewallRuleDescription` | `PolicyDescription` | `Remark` | `Description` |
| **规则粒度** | 实例绑定 | 安全组绑定（可关联多实例） | 实例绑定 | 安全组绑定（可关联多实例） |

### 5.3 核心差异分析

**规则方向：** 所有云产品的防火墙规则均为**入站方向（Ingress）**。本工具仅管理入站规则，不涉及出站规则。

**实例级防火墙 vs 安全组：**
- **实例级**（Lighthouse / 阿里轻量云）：规则直接绑定到单台实例，操作时传 `InstanceId`
- **安全组级**（CVM / 阿里 ECS）：规则属于安全组，安全组可关联多台实例；操作时传 `SecurityGroupId`，规则变更自动生效于组内所有实例

**规则标识方式（用于识别本工具创建的规则）：**
- 腾讯云 Lighthouse：`FirewallRuleDescription` 以 `[RULE_TAG]` 开头
- 腾讯云 CVM：`PolicyDescription` 以 `[RULE_TAG]` 开头
- 阿里云轻量云：`Remark` 以 `[RULE_TAG]` 开头
- 阿里云 ECS：`Description` 以 `[RULE_TAG]` 开头（限制 256 字节，足够使用）

**端口格式差异：**
- 腾讯云 Lighthouse / 阿里轻量云：`80` 或 `443,80` 或 `ALL`
- 腾讯云 CVM：`80` 或 `8000-8010`（单端口或范围）
- 阿里云 ECS：`80/80` 或 `1/200`（必须用斜杠格式）

**IPv6 支持：**
- 腾讯云 Lighthouse：`Ipv6CidrBlock` 字段
- 腾讯云 CVM：`Ipv6CidrBlock` 字段
- 阿里云轻量云：`SourceCidrIp` 字段（⚠️ **待确认**：阿里轻量云防火墙是否支持 IPv6；若不支持，IPv6 解析结果应跳过并 WARN 日志）
- 阿里云 ECS：`Ipv6SourceCidrIp` 字段（独立于 IPv4 的 `SourceCidrIp`）

### 5.4 Provider 抽象接口设计

```go
// provider/provider.go

package provider

import "github.com/alcaprophet/fwalizer/dns"

// CloudType 云厂商标识
type CloudType string

const (
    CloudTCLighthouse CloudType = "tc_lighthouse"
    CloudTCCVM       CloudType = "tc_cvm"
    CloudAliSWAS     CloudType = "ali_swas"
    CloudAliECS      CloudType = "ali_ecs"
)

// RuleInfo 统一的防火墙规则信息（从云端查询回来的）
type RuleInfo struct {
    Protocol      string // TCP / UDP / ICMP / ALL
    Port          string // 端口（已归一化为 "port" 或 "start-end" 格式）
    CidrBlock     string // IPv4 CIDR
    Ipv6CidrBlock string // IPv6 CIDR
    Action        string // ACCEPT / DROP
    Description   string // 规则描述/备注
    PolicyIndex   string // 规则索引（安全组类需要，防火墙类可为空）
    RuleID        string // 规则唯一 ID（部分 API 需要）
}

// RuleAction 要写入云端的防火墙规则（创建/删除用）
type RuleAction struct {
    Protocol      string
    Port          string // 已转换为对应云厂商的端口格式
    CidrBlock     string
    Ipv6CidrBlock string
    Action        string
    Description   string
}

// TargetConfig 目标云资源配置（凭据通过独立环境变量传入）
type TargetConfig struct {
    CloudType  CloudType
    Region     string
    ResourceID string // InstanceId 或 SecurityGroupId（根据 CloudType 决定含义）
}

// Config 应用配置（凭据按云厂商分离）
type Config struct {
    // 腾讯云凭据
    TCAccessID   string
    TCAccessKey  string
    // 阿里云凭据
    AliAccessID  string
    AliAccessKey string

    Targets     []TargetConfig
    DomainRules []DomainRule
    Tag         string
    Interval    time.Duration
    DNS         string
    LogLevel    string // debug / info / warn / error
}

// Provider 多云防火墙操作接口（核心抽象）
type Provider interface {
    // Name 返回 Provider 名称，用于日志
    Name() string

    // CloudType 返回云厂商标识
    CloudType() CloudType

    // GetRules 获取当前所有防火墙/安全组规则
    // 返回: 规则列表, 总数, 错误
    GetRules() ([]RuleInfo, error)

    // CreateRules 批量添加规则
    CreateRules(rules []RuleAction) error

    // DeleteRules 批量删除规则
    DeleteRules(rules []RuleAction) error

    // ConvertPort 将通用端口格式转换为该云厂商的格式
    // 通用格式: "80" 或 "443,80" 或 "ALL"
    ConvertPort(port string) string

    // TargetIndex 返回该 Provider 在 TARGETS 列表中的编号（从 1 开始）
    // 用于 RULES 中的 targets 列绑定匹配
    TargetIndex() int
}
```

### 5.5 目录结构（重构后）

```
fwalizer/
├── main.go                      # 入口：模式判定 + 启动
├── app/                         # 应用生命周期管理
│   ├── app.go                   # App 结构体，协调各组件
│   ├── mode.go                  # 运行模式检测（env/webui/docker/desktop）
│   ├── systray.go               # 系统托盘（仅桌面端编译，//go:build desktop）
│   └── cli.go                   # CLI 子命令（version / validate / backup / restore）
├── config/
│   ├── config.go                # 配置模型定义
│   ├── env.go                   # .env 解析器
│   ├── store.go                 # SQLite 持久化
│   └── validate.go              # 配置校验
├── dns/
│   └── resolver.go              # DNS 解析（基本不变）
├── provider/                    # 多云抽象层
│   ├── provider.go              # Provider 接口定义
│   ├── registry.go              # Provider 注册表（工厂模式）
│   ├── common.go                # 通用工具（ownedRules、Diff、端口转换等）
│   ├── tc_lighthouse.go          # 腾讯云 Lighthouse 实现
│   ├── tc_cvm.go                # 腾讯云 CVM 安全组实现
│   ├── ali_swas.go              # 阿里云轻量云实现
│   └── ali_ecs.go               # 阿里云 ECS 安全组实现
├── syncer/                      # 同步引擎（从 firewall/ 重命名）
│   ├── syncer.go                # 同步主循环
│   ├── retry.go                 # 重试逻辑（指数退避）
│   └── ratelimit.go             # API 频率控制
├── webui/
│   ├── server.go                # HTTP Server
│   ├── api/                     # REST API handlers
│   │   ├── targets.go           # 云资源 CRUD
│   │   ├── rules.go             # 规则 CRUD
│   │   ├── sync.go              # 同步状态/手动触发
│   │   └── settings.go          # 全局设置
│   └── frontend/                # Vue 3 SPA（embed 嵌入）
│       └── dist/
├── notifier/                    # 事件总线 + 告警
│   ├── bus.go
│   ├── email.go
│   └── webhook.go
├── internal/                    # 内部工具
│   ├── portconv/                # 端口格式转换
│   └── tag/                     # 规则标签生成/解析
├── version/
│   └── version.go               # 版本信息（通过 ldflags 注入）
└── build/
    └── Dockerfile               # Docker 构建文件
```

### 5.6 Provider 工厂注册

```go
// provider/registry.go

package provider

import "fmt"

// Factory 创建 Provider 的工厂函数
type Factory func(cfg TargetConfig) (Provider, error)

var registry = map[CloudType]Factory{}

// Register 注册 Provider 工厂
func Register(cloudType CloudType, factory Factory) {
    registry[cloudType] = factory
}

// NewProvider 根据配置创建对应的 Provider
func NewProvider(cfg TargetConfig) (Provider, error) {
    factory, ok := registry[cfg.CloudType]
    if !ok {
        return nil, fmt.Errorf("不支持的云产品类型: %s", cfg.CloudType)
    }
    return factory(cfg)
}

// 各 Provider 在 init() 中自注册
func init() {
    // 注册由各实现文件的 init() 完成
}
```

### 5.7 通用 Diff 逻辑（provider-agnostic）

```go
// provider/common.go

// ruleKey 统一规则唯一标识
type ruleKey struct {
    protocol      string
    port          string
    cidrBlock     string
    ipv6CidrBlock string
    action        string
}

// OwnedRules 从全量规则中提取本工具管理的规则
func OwnedRules(allRules []RuleInfo, tagPrefix string) []RuleInfo {
    prefix := "[" + tagPrefix + "]"
    var owned []RuleInfo
    for _, r := range allRules {
        if strings.HasPrefix(r.Description, prefix) {
            owned = append(owned, r)
        }
    }
    return owned
}

// Diff 计算需要添加和删除的规则
func Diff(
    resolved []dns.ResolvedIP,
    rule DomainRule,
    desc string,
    existing []RuleInfo,
    portConverter func(string) string,
) (toAdd []RuleAction, toDelete []RuleAction) {
    // ... 通用 diff 逻辑，与云厂商无关
}
```

### 5.8 Syncer 重构（调用 Provider 接口）

```go
// syncer/syncer.go（重构后）

type Syncer struct {
    cfg        *config.Config
    providers  []provider.Provider  // 多个 Provider 实例
    resolver   *dns.Resolver
    configCh   chan *config.Config  // 接收热更新配置
    stopCh     chan struct{}
    wg         sync.WaitGroup
    clientPool *ClientPool          // SDK Client 复用池
}

// ClientPool SDK Client 复用池
// 相同云厂商 + Region + 凭据的多个 Target 共享 Client 实例，避免重复创建连接
type ClientPool struct {
    clients map[string]provider.Provider  // key: cloudType|region|accessID
}

// Reload 热更新配置（WebUI 修改后调用）
func (s *Syncer) Reload(newCfg *config.Config) {
    s.configCh <- newCfg
}

func (s *Syncer) syncAll() {
    // 按云厂商分组并行同步（不同云厂商 API 配额独立）
    // 同一云厂商内串行处理（共享 API 配额）
    var groupWg sync.WaitGroup
    groups := s.groupByCloud()
    for cloudType, providers := range groups {
        groupWg.Add(1)
        go func(ct CloudType, ps []provider.Provider) {
            defer groupWg.Done()
            for _, p := range ps {
                slog.Info("开始同步", "provider", p.Name(), "cloud", ct)
                rules := filterRulesForTarget(s.cfg.DomainRules, p.TargetIndex())
                for _, rule := range rules {
                    s.syncDomain(p, rule)
                    time.Sleep(s.rateLimitInterval(ct)) // 按云厂商的 API 限制调整间隔
                }
            }
        }(cloudType, providers)
    }
    groupWg.Wait()
}

func (s *Syncer) syncDomain(p provider.Provider, rule config.DomainRule) {
    // 1. DNS 解析（带可配置超时，默认 10s）
    resolved, err := s.resolver.Lookup(rule.Host)
    // 2. 获取当前规则
    allRules, err := p.GetRules()
    // 3. 过滤本工具管理的规则
    owned := provider.OwnedRules(allRules, s.cfg.Tag)
    // 4. Diff
    toAdd, toDelete := provider.Diff(resolved, rule, desc, owned, p.ConvertPort)
    // 5. 写入（带重试）
    s.applyWithRetry(p, toAdd, toDelete, 3)
}

// rateLimitInterval 根据云厂商返回 API 调用间隔
func (s *Syncer) rateLimitInterval(cloudType CloudType) time.Duration {
    switch cloudType {
    case CloudAliSWAS:
        return 800 * time.Millisecond  // 阿里云轻量云 100次/60秒，需更保守
    default:
        return 500 * time.Millisecond  // 腾讯云、阿里云 ECS 等
    }
}
```

### 5.9 SDK 依赖

```
// go.mod 新增依赖
require (
    // 腾讯云（现有 + 新增 CVM/VPC）
    github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.3.137
    github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse v1.3.108
    github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc v1.3.xxx    // 新增

    // 阿里云
    github.com/aliyun/alibaba-cloud-sdk-go v1.63.xxx                         // 新增

    // 系统托盘（仅桌面端构建）
    fyne.io/systray v1.11.0                                                  // 新增
)
```

> **注意**：阿里云 V1 Go SDK (`alibaba-cloud-sdk-go`) 已终止支持。短期使用 V1 SDK 的 `services/ecs` 和 `services/swas-open`，后续关注 V2 SDK 进展再迁移。

### 5.10 配置格式（.env）

不再兼容旧 .env，采用全新设计。**凭据与资源声明分离**，避免密钥泄露。

```env
# ─── 云资源目标 ───
# 格式: provider|resource_id|region
# 逗号分隔多个目标，支持反斜杠换行
TARGETS=tc_lighthouse|lhins-abc|ap-guangzhou, \
        tc_cvm|sg-def|ap-shanghai, \
        ali_swas|ace0706b|cn-hangzhou, \
        ali_ecs|sg-ghi|cn-shenzhen

# ─── 云厂商凭据（按厂商分离，避免泄露到进程列表 / docker inspect） ───
TC_ACCESS_ID=       # 腾讯云 SecretId（tc_lighthouse / tc_cvm 共用）
TC_ACCESS_KEY=      # 腾讯云 SecretKey
ALI_ACCESS_ID=      # 阿里云 AccessKeyId（ali_swas / ali_ecs 共用）
ALI_ACCESS_KEY=     # 阿里云 AccessKeySecret

# ─── 域名规则 ───
# 格式: domain|protocol|ports|action|targets|comment
# targets 列：逗号分隔目标编号（从 1 开始），留空或 * 表示全部
# comment 列：可选备注，写入防火墙规则描述（便于人工识别）
RULES=api.example.com|TCP|443,80|ACCEPT||生产API, \
      vpn.example.com|UDP|1194|ACCEPT|2|VPN接入, \
      dev.example.com|TCP|22|DROP|1,3|禁止SSH

# ─── 全局设置 ───
TAG=auto-dns
INTERVAL=5m
DNS=8.8.8.8:53
LOG_LEVEL=info        # debug / info / warn / error
WEBUI_PORT=9090
```

**设计要点：**

| 变量 | 说明 |
|------|------|
| `TARGETS` | 仅含资源声明（provider / resource_id / region），编号隐式从 1 开始 |
| `TC_ACCESS_ID` / `TC_ACCESS_KEY` | 腾讯云凭据，`tc_lighthouse` 和 `tc_cvm` 共用 |
| `ALI_ACCESS_ID` / `ALI_ACCESS_KEY` | 阿里云凭据，`ali_swas` 和 `ali_ecs` 共用 |
| `RULES` | 第 5 列 `targets` 控制规则应用到哪些目标；省略 = 全部目标；第 6 列 `comment` 可选备注 |
| `TAG` | 规则标签前缀，替代旧的 `RULE_TAG` |
| `INTERVAL` | 同步间隔，替代旧的 `CHECK_INTERVAL` |
| `DNS` | DNS 服务器，替代旧的 `DNS_SERVER` |
| `LOG_LEVEL` | 日志级别，默认 `info` |
| `WEBUI_PORT` | WebUI 端口，默认 9090（WebUI 模式可配置） |

**变量名精简对比：**

| 旧变量 | 新变量 | 变化 |
|--------|--------|------|
| `TENCENTCLOUD_SECRET_ID` | `TC_ACCESS_ID` | 凭据独立，不嵌入 TARGETS |
| `TENCENTCLOUD_SECRET_KEY` | `TC_ACCESS_KEY` | 同上 |
| `LIGHTHOUSE_INSTANCE_ID` | 内嵌于 `TARGETS` | 资源 ID 统一在 TARGETS |
| `LIGHTHOUSE_REGION` | 内嵌于 `TARGETS` | Region 统一在 TARGETS |
| `DOMAIN_RULES` | `RULES` | 变量名精简 + 新增目标绑定列 + 保留 comment 列 |
| `RULE_TAG` | `TAG` | 变量名精简 |
| `CHECK_INTERVAL` | `INTERVAL` | 变量名精简 |
| `DNS_SERVER` | `DNS` | 变量名精简 |

**凭据安全说明：** 旧版将 AK/SK 嵌入 `DOMAIN_RULES` 同级变量（`TENCENTCLOUD_SECRET_ID` / `TENCENTCLOUD_SECRET_KEY`），密钥会暴露在进程列表（`/proc/1/environ`）、`docker inspect` 输出及日志中。新版将凭据改为独立环境变量（`TC_ACCESS_ID` / `TC_ACCESS_KEY` / `ALI_ACCESS_ID` / `ALI_ACCESS_KEY`），TARGETS 中仅含资源标识，不含任何密钥信息。程序启动时根据每个 Target 的 `CloudType` 自动匹配对应凭据。

**反斜杠换行处理：** `.env` 标准不支持多行值。解析器在读取前预处理：将 `\` 续行的多行合并为单行，再按逗号分割。

> **注意**：`.env` 模式仅在用户主动传入 `TARGETS` 环境变量时激活（高级/运维模式）。默认情况下程序以 WebUI 模式启动，配置通过 WebUI 界面传入并持久化到 SQLite。详见第三节「配置模式重设计」。

### 5.11 各 Provider 实现要点

#### 腾讯云 Lighthouse（现有，迁移）

- 将 `firewall/client.go` 和 `firewall/rule.go` 中的 Lighthouse 特定逻辑迁移到 `provider/tc_lighthouse.go`
- 保持 API 调用方式不变
- 规则标识：`FirewallRuleDescription` 以 `[RULE_TAG]` 开头

#### 腾讯云 CVM 安全组

- SDK: `tencentcloud-sdk-go/tencentcloud/vpc/v20170312`
- 关键 API: `DescribeSecurityGroupPolicies` / `CreateSecurityGroupPolicies` / `DeleteSecurityGroupPolicies`
- **操作对象是 SecurityGroupId**（不是 InstanceId），用户需提前将安全组绑定到目标 CVM 实例
- 端口格式：单端口 `"80"` 或范围 `"8000-8010"`（不支持逗号分隔，需在 ConvertPort 中转换）
- 删除规则时需传 `PolicyIndex`（从 Describe 获取）或完整规则匹配
- 规则描述：`PolicyDescription` 字段
- API 频率限制 100次/秒，较宽松

#### 阿里云轻量云 (SWAS-OPEN)

- SDK: `alibaba-cloud-sdk-go/services/swas-open`（V1）
- 关键 API: `ListFirewallRules` / `CreateFirewallRules` / `DeleteFirewallRules`
- 与腾讯云 Lighthouse 非常相似（实例级防火墙）
- 端口格式：`"80"` 或 `"1/200"`（斜杠分隔范围）
- 规则标识：`Remark` 字段
- API 频率限制 100次/60秒（~1.7次/秒），**需要更长的间隔（建议 800ms）**

#### 阿里云 ECS 安全组

- SDK: `alibaba-cloud-sdk-go/services/ecs`
- 关键 API: `DescribeSecurityGroupAttribute` / `AuthorizeSecurityGroup` / `RevokeSecurityGroup`
- **操作对象是 SecurityGroupId**
- 端口格式：必须为 `"80/80"` 或 `"1/200"`（即使单端口也要斜杠格式）
- IPv6 使用独立字段 `Ipv6SourceCidrIp`（与 IPv4 的 `SourceCidrIp` 互斥）
- 规则标识：`Description` 字段（限制 256 字节，足够）
- 规则唯一性由 `协议+端口+源CIDR+策略+优先级` 决定，重复添加不会报错但也不会新增

---

## 六、多实例多地域支持

（已合并到第五节「多云支持架构」中，通过 `TARGETS` 统一配置）

### 架构改动要点

```
Config
  └── Targets []TargetConfig     // 从 TARGETS 解析
        ├── CloudType  CloudType  // provider 字段
        ├── Region     string
        └── ResourceID string

  └── Rules []DomainRule         // 从 RULES 解析
        ├── Host     string
        ├── Protocol string
        ├── Ports    string
        ├── Action   string
        ├── Targets  []int       // 目标编号列表（空 = 全部）
        └── Comment  string      // 可选备注，写入防火墙规则描述

Syncer
  ├── providers []Provider       // 每个 Target 对应一个 Provider 实例
  └── syncAll()
        └── for i, p := range providers
              └── rules := filterRulesForTarget(cfg.Rules, i+1)
              └── for _, rule := range rules
                    └── syncDomain(p, rule)
```

**关键约束：**
- 同一云厂商同一 Region 可复用 SDK 连接
- 不同云厂商 API 频率限制不同，取最严格的间隔（建议统一 800ms）
- 检测到规则已存在时提示并跳过，不报错
- 规则通过 targets 列精确控制应用到哪些云资源，未指定则应用到全部

---

## 七、WebUI 管理面板

### 技术选型

| 层 | 技术 | 理由 |
|---|------|------|
| 后端路由 | Go `net/http` + `encoding/json` | 标准库，无外部依赖 |
| 前端框架 | **Vue 3** (CDN 引入) | 轻量、学习成本低、单文件可用 |
| 静态资源嵌入 | Go `embed` | 编译进二进制，零部署成本 |
| 配置持久化 | **SQLite** (via `modernc.org/sqlite`) | 纯 Go 实现，无需 CGO，单文件数据库 |
| UI 组件 | **Naive UI** 或手写 CSS | 轻量，不引入重型 UI 库 |

### 功能规划

```
WebUI (localhost:9090)
├── 仪表盘
│   ├── 实例状态概览（云厂商、类型、最近同步时间）
│   └── 规则统计（总数、本工具管理数、最近变更）
├── 云资源管理
│   ├── 添加/编辑/删除云资源（选择云厂商 → 产品 → 输入 AK/SK → 输入资源 ID）
│   ├── 连接测试（验证 AK/SK 和资源 ID 是否有效）
│   └── 批量导入
├── 域名规则管理
│   ├── 可视化编辑 RULES（表单替代管道符语法）
│   ├── 规则绑定云资源（多对多）
│   └── 实时 DNS 解析测试
├── 全局设置
│   ├── DNS 服务器配置
│   ├── 检查间隔设置
│   └── RULE_TAG 配置
├── 同步日志
│   └── 最近 N 条同步记录（时间、结果、变更详情）
├── 高级功能
│   ├── 规则试运行（Dry Run：DNS 解析 + Diff，不实际写入）
│   ├── 配置导入/导出（JSON 格式，便于备份迁移）
│   └── 健康检查端点（GET /api/health）
└── 告警配置
    ├── 邮件推送（SMTP）
    └── Webhook 通知
```

### 架构示意

```
main.go
├── 模式判定（determineMode）
├── HTTP Server (:9090)          // WebUI API（仅 WebUI 模式启动）
│   ├── /api/targets             // 云资源 CRUD
│   ├── /api/rules               // 规则 CRUD
│   ├── /api/sync/status         // 同步状态
│   ├── /api/sync/dryrun         // 规则试运行
│   ├── /api/settings            // 全局设置
│   ├── /api/test-connection     // 测试云厂商连接
│   ├── /api/config/export       // 配置导出
│   ├── /api/config/import       // 配置导入
│   ├── /api/health              // 健康检查
│   └── /                        // 嵌入的 Vue SPA
├── Syncer (goroutine)           // 同步主循环
│   └── configCh                 // 接收热更新配置
└── Config Store (SQLite)        // 持久化配置（仅 WebUI 模式）
```

### 配置来源

详见第三节「配置模式重设计」。两种模式互斥：

- **`.env` 模式**：从环境变量解析配置，不启动 WebUI，不读写 SQLite
- **WebUI 模式**：从 SQLite 加载配置，通过 WebUI 界面管理，支持热更新

---

## 八、可扩展性设计

### 事件总线（Event Bus）

```go
type EventType string

const (
    EventSyncStart    EventType = "sync:start"
    EventSyncComplete EventType = "sync:complete"
    EventSyncError    EventType = "sync:error"
    EventRuleChanged  EventType = "rule:changed"
    EventDNSFailed    EventType = "dns:failed"
)

type Event struct {
    Type      EventType
    Timestamp time.Time
    Data      map[string]any
}

type Subscriber interface {
    OnEvent(event Event) error
}

type EventBus struct {
    subscribers map[EventType][]Subscriber
}
```

### 扩展插件

```
notifier/
├── email.go       // SMTP 邮件告警
├── webhook.go     // Webhook 通知（钉钉/飞书/Slack）
├── telegram.go    // Telegram Bot
└── log.go         // 本地日志记录
```

---

## 九、推荐实施路线

```
Phase 1 — Provider 抽象 + 多云基础 + 配置重设计（1-2 周）
├── [ ] 定义 Provider 接口 + 统一规则模型（RuleInfo / RuleAction）
├── [ ] 迁移现有 Lighthouse 代码为 TencentLighthouseProvider
├── [ ] 重构 Syncer 调用 Provider 接口（解耦云厂商特定代码）
├── [ ] 实现 TencentCVMProvider（安全组）
├── [ ] 实现 AliyunSWASProvider（轻量云）
├── [ ] 实现 AliyunECSProvider（安全组）
├── [ ] 重构 Config 支持新 .env 格式（TARGETS / RULES / TAG）
├── [ ] 实现运行模式判定逻辑（env / webui / docker / desktop）
├── [ ] 废弃旧 .env 变量，无需向后兼容
├── [ ] 单元测试：各 Provider 的 ConvertPort + Diff 逻辑
├── [ ] 凭据分离：TC_ACCESS_ID / TC_ACCESS_KEY / ALI_ACCESS_ID / ALI_ACCESS_KEY
├── [ ] CLI 子命令：version / validate / backup / restore
├── [ ] 日志级别配置：LOG_LEVEL 环境变量（debug/info/warn/error）
└── [ ] 版本号注入：version/version.go（-ldflags 注入）

Phase 2 — WebUI + 配置持久化（1-2 周）
├── [ ] 引入 SQLite 配置持久化（config/store.go）
├── [ ] 实现 REST API（云资源/规则/设置 CRUD）
├── [ ] 搭建 Vue 3 前端框架 + embed 嵌入
├── [ ] 实现基础 WebUI（云资源管理 + 规则可视化编辑）
├── [ ] 实现连接测试功能（验证 AK/SK 有效性）
├── [ ] 仪表盘（同步状态、规则统计）
├── [ ] 健康检查端点（/api/health）
└── [ ] 配置热重载（Syncer 接收新配置无需重启）

Phase 3 — 告警 + 高级功能（1-2 周）
├── [ ] 事件总线 EventBus 实现
├── [ ] 邮件告警（SMTP）
├── [ ] Webhook 通知（钉钉/飞书）
├── [ ] 同步日志查看
├── [ ] 规则试运行（Dry Run）
├── [ ] 配置导入/导出（JSON 格式）
├── [ ] Docker HEALTHCHECK 对接 WebUI
└── [ ] 规则模板（预设常用规则集）

Phase 4 — 桌面端 + 打磨（1-2 周）
├── [ ] 系统托盘集成（fyne.io/systray，CGO_ENABLED=1）
├── [ ] Build tags 分离（docker vs desktop，含 CGO 差异）
├── [ ] 开机自启注册（Windows Startup / macOS Login Items）
├── [ ] macOS .app / Windows .exe 打包
├── [ ] WebUI/托盘显示版本号（Phase 1 已实现注入）
├── [ ] 同步日志持久化（SQLite 存储最近 N 条）
└── [ ] 规则变更历史 & 回滚
```

---

## 十、技术决策总结

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 后端语言 | **Go**（维持现状） | 零重写成本、单二进制、跨平台编译 |
| 配置默认模式 | **WebUI + SQLite 优先** | 降低使用门槛，可视化配置，`.env` 仅作为高级/运维模式 |
| .env 模式触发 | **检测 `TARGETS` 环境变量存在** | 自动判定，无需额外参数；可通过 `FWALIZER_MODE` 强制覆盖 |
| 多云抽象 | **Provider 接口 + 工厂注册** | 解耦彻底，新增云厂商仅需实现接口 |
| 统一规则模型 | **RuleInfo / RuleAction** | 屏蔽各云 API 差异，Diff 逻辑通用化 |
| 腾讯云 CVM SDK | `tencentcloud-sdk-go/tencentcloud/vpc` | 安全组属于 VPC 模块 |
| 阿里云 SDK | `alibaba-cloud-sdk-go`（V1） | 短期可用，V2 SDK 待 SWAS-OPEN 独立拆分后迁移 |
| Web 后端 | `net/http` + `encoding/json` | 标准库，无外部依赖 |
| Web 前端 | **Vue 3** | 轻量、易上手、生态成熟 |
| 配置存储 | **SQLite** (`modernc.org/sqlite`) | 纯 Go、单文件、无需安装数据库 |
| 桌面方案 | **系统托盘 + 浏览器 WebUI** | 最符合"后台工具"定位，WebUI 零额外开发成本，体积小 |
| 系统托盘库 | **`fyne.io/systray`** | 跨平台、轻量（~2MB），需 CGO |
| 构建分离 | **Go build tags + CGO 差异** | Docker `CGO_ENABLED=0` 纯静态，Desktop `CGO_ENABLED=1` 含托盘 |
| 配置热重载 | **channel 通知 Syncer** | WebUI 修改即时生效，无需重启进程 |
| 健康检查 | **`/api/health` HTTP 端点** | Docker HEALTHCHECK + 仪表盘共用 |
| 扩展架构 | **Event Bus + Subscriber** | 解耦、易扩展、无外部依赖 |
| WebUI 安全 | **默认绑定 127.0.0.1** | 仅本机访问，Docker 用户通过 `-p` 自行控制暴露范围 |
| WebUI 端口 | **可配置，默认 9090** | 通过 `WEBUI_PORT` 环境变量调整，避免端口冲突 |
| RULES Comment | **6 列格式，保留备注** | `domain\|protocol\|ports\|action\|targets\|comment`，人工识别 |
| SDK Client 复用 | **ClientPool 连接池** | 相同云厂商+Region+凭据的 Target 共享 Client，避免重复创建 |
| 同步并发策略 | **跨厂商并行，同厂商串行** | 不同云厂商 API 配额独立，可并行；同厂商内串行保护配额 |
| 桌面存储路径 | **系统标准路径** | macOS `~/Library/Application Support`，Windows `%APPDATA%`，Linux `~/.config` |
| 版本号注入 | **ldflags `-X main.version`** | 编译时注入，WebUI/托盘显示，便于用户确认版本 |
| 凭据管理 | **按云厂商独立环境变量** | 避免密钥泄露到进程列表 / docker inspect，TARGETS 不含敏感信息 |
| CLI 命令 | **version / validate / backup / restore** | 开箱即用：配置校验、手动备份恢复、版本查询 |
| 日志级别 | **`LOG_LEVEL` 环境变量** | debug/info/warn/error，默认 info，按需调优 |
| DNS 失败处理 | **渐进式熔断** | 连续失败自动告警 + 暂停异常域名同步，不影响其他域名 |

---

## 十一、补充设计要点

### 11.1 DNS 解析超时优化

当前 DNS 超时分两层配置：
- `net.Dialer` 连接超时 **10s**（建立 UDP/TCP 连接）
- `context.WithTimeout` 整体查询超时 **15s**（含连接 + 查询）

多域名场景下，若第一个域名 DNS 超时，会阻塞后续所有域名的同步。

**优化方案：**
- 将 context 超时从 15s 降低到 10s（可通过 `DNS_TIMEOUT` 环境变量配置）
- 各域名 DNS 解析并行执行，提前收集所有解析结果后再进入 Diff 阶段
- 单个域名解析失败不影响其他域名（现有逻辑已满足）

### 11.2 Docker HEALTHCHECK 升级

引入 WebUI 后，Docker HEALTHCHECK 应优先使用 HTTP 健康检查：

```dockerfile
# WebUI 模式启动较慢（HTTP Server + SQLite 初始化），start-period 调整为 10s
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q --spider http://localhost:9090/api/health || exit 1
```

- `.env` 模式（无 WebUI）仍使用 `killall -0` 进程检测
- WebUI 模式使用 `/api/health` HTTP 检测，能更准确地反映服务可用性

### 11.3 同步日志持久化

`.env` 模式仅输出 stdout 日志（`docker logs` 查看），WebUI 模式需将同步记录持久化到 SQLite：

| 日志字段 | 说明 |
|----------|------|
| `timestamp` | 同步时间 |
| `target` | 目标云资源标识 |
| `domain` | 域名 |
| `result` | success / failed / skipped |
| `added` | 新增规则数 |
| `deleted` | 删除规则数 |
| `error` | 错误信息（失败时） |

- SQLite 保留最近 1000 条记录（可配置），超出自动清理
- WebUI 「同步日志」页面展示，支持筛选和搜索

### 11.4 .env 反斜杠续行解析

标准 `.env` 格式不支持多行值。解析器在读取前进行预处理：

```go
// 预处理：将 \ 续行的多行合并为单行
// 例如: TARGETS=xxx, \
//           yyy
// 合并为: TARGETS=xxx, yyy
```

**注意事项：**
- 仅处理以 `\` 结尾的续行，不处理其他转义字符
- 解析失败时给出清晰错误提示（如行号、上下文）
- 合并后按逗号分割时，自动去除每项的前后空白

### 11.5 版本号管理

```go
// version/version.go
package version

var Version = "dev"  // 通过 -ldflags "-X fwalizer/version.Version=v1.0.0" 注入
```

- Docker 构建和桌面构建均通过 `-ldflags` 注入版本号
- WebUI 页面底部显示当前版本号
- 系统托盘右键菜单显示版本号
- GitHub Actions 构建时自动从 git tag 提取版本号

### 11.6 阿里云 SDK V2 迁移评估

阿里云 V1 Go SDK 已终止支持。短期使用 V1，后续关注 V2 SDK（`github.com/alibabacloud-go/`）的 SWAS-OPEN 覆盖进展再迁移。迁移时只需修改 Provider 实现，不影响接口层。

### 11.7 补充设计要点（遗漏项）

#### CLI 命令支持

提供基础 CLI 子命令，提升运维体验：

| 命令 | 功能 | 说明 |
|------|------|------|
| `fwalizer version` | 显示版本号 | 等价于 `fwalizer -v` |
| `fwalizer validate` | 校验 .env 配置 | 不启动同步，仅验证格式和云凭据有效性 |
| `fwalizer backup` | 备份 SQLite 数据库 | 复制 `config.db` → `config.db.bak.{timestamp}` |
| `fwalizer restore [file]` | 从备份恢复 | 恢复指定备份文件到 `config.db` |

#### 配置校验命令

`fwalizer validate` 执行以下检查：
- `.env` 格式校验（TARGETS / RULES 字段数、端口范围、协议类型等）
- 云凭据有效性（调用各云 Describe API 验证 AK/SK）
- 资源 ID 存在性（验证实例/安全组是否可访问）
- DNS 服务器可达性（向指定 DNS 发送测试查询）
- 域名解析预检（对所有配置的域名执行一次 DNS 解析）

#### 日志级别配置

通过 `LOG_LEVEL` 环境变量控制日志输出级别：

| 级别 | 说明 |
|------|------|
| `debug` | 详细输出：DNS 解析结果、API 请求/响应、Diff 详情 |
| `info` | 默认级别：同步状态、规则变更摘要 |
| `warn` | 仅告警：DNS 失败、重试、规则已存在等 |
| `error` | 仅错误：写入失败、连接失败等致命问题 |

#### 连续 DNS 失败告警

单个域名 DNS 解析失败时保留现有规则（现有逻辑），但需对**持续失败**进行告警：

```
连续失败处理策略：
1 次 → WARN 日志，保留现有规则
3 次 → ERROR 日志，输出连续失败告警
5 次 → 标记为 "熔断" 状态，暂停该域名同步（不影响其他域名）
恢复后（DNS 解析成功） → 自动解除熔断
```

- 熔断状态通过 Event Bus 发送 `EventDNSFailed` 事件，可触发邮件/Webhook 告警
- 熔断阈值可通过 `DNS_FAIL_THRESHOLD` 环境变量配置（默认 5）

#### SQLite 备份与恢复

提供 CLI 命令手动管理 SQLite 数据库备份：

```bash
fwalizer backup                    # 备份 config.db → config.db.bak.{timestamp}
fwalizer restore config.db.bak.1   # 从指定备份恢复
fwalizer backup --list             # 列出所有备份文件
```

- 备份文件与 `config.db` 位于同一目录
- 最多保留 5 个备份（自动轮转，超出删除最旧的）
- 备份前自动校验 SQLite 数据库完整性（`PRAGMA integrity_check`）
