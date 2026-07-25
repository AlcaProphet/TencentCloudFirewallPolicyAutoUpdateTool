# FWAlizer 技术实现细节（Build1）

> 代码级别的实现参考。设计大纲见 [Design1.md](./Design1.md)，AI 编码指令见 [AGENTS.md](./AGENTS.md)。

---

## 一、重构后目录结构

```
fwalizer/
├── main.go                      # 入口：模式判定 + 启动
├── app/                         # 应用生命周期管理
│   ├── app.go                   # App 结构体，协调各组件
│   ├── mode.go                  # 运行模式检测（env/webui/docker/desktop）
│   ├── systray.go               # 系统托盘（仅桌面端，//go:build desktop）
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
│   ├── common.go                # 通用工具（ownedRules、Diff、端口转换）
│   ├── tc_lighthouse.go         # 腾讯云 Lighthouse 实现
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
├── internal/
│   ├── portconv/                # 端口格式转换
│   └── tag/                     # 规则标签生成/解析
├── version/
│   └── version.go               # 版本信息（ldflags 注入）
└── build/
    └── Dockerfile
```

---

## 二、核心数据模型

### 2.1 CloudType 定义

```go
type CloudType string

const (
    CloudTCLighthouse CloudType = "tc_lighthouse"
    CloudTCCVM       CloudType = "tc_cvm"
    CloudAliSWAS     CloudType = "ali_swas"
    CloudAliECS      CloudType = "ali_ecs"
)
```

### 2.2 RuleInfo（云端查询回来的规则）

```go
type RuleInfo struct {
    Protocol      string // TCP / UDP / ICMP / ALL
    Port          string // 归一化为 "port" 或 "start-end"
    CidrBlock     string // IPv4 CIDR
    Ipv6CidrBlock string // IPv6 CIDR
    Action        string // ACCEPT / DROP
    Description   string // 规则描述/备注
    PolicyIndex   string // 规则索引（安全组类需要）
    RuleID        string // 规则唯一 ID
}
```

### 2.3 RuleAction（要写入云端的规则）

```go
type RuleAction struct {
    Protocol      string
    Port          string // 已转换为对应云厂商的端口格式
    CidrBlock     string
    Ipv6CidrBlock string
    Action        string
    Description   string
}
```

### 2.4 TargetConfig 与 Config

```go
type TargetConfig struct {
    CloudType  CloudType
    Region     string
    ResourceID string // InstanceId 或 SecurityGroupId
}

type Config struct {
    TCAccessID   string        // 腾讯云凭据
    TCAccessKey  string
    AliAccessID  string        // 阿里云凭据
    AliAccessKey string

    Targets     []TargetConfig
    DomainRules []DomainRule
    Tag         string
    Interval    time.Duration
    DNS         string
    DNSTimeout  time.Duration  // DNS 解析超时，默认 10s
    LogLevel    string         // debug / info / warn / error
}
```

### 2.5 DomainRule（RULES 解析结果）

```go
type DomainRule struct {
    Host     string
    Protocol string // TCP / UDP / TCP+UDP / ICMP
    Ports    string // 单端口、逗号分隔、范围（8000-8010）、ALL；ICMP 时固定为 ALL
    Action   string
    Targets  []int  // 目标编号（空或 * = 全部）
    Comment  string // 可选备注
}
```

**Targets 解析规则：**
- 空字符串或 `*` → 应用到所有 Target
- `"1,3"` → 解析为 `[]int{1, 3}`，编号从 1 开始对应 TARGETS 顺序
- 编号超出范围时启动报错（配置校验阶段拦截）

**端口格式说明：**
- 单端口：`443`
- 多端口：`443,80`
- 范围：`8000-8010`（start-end，start ≤ end）
- 混合：`80,443,8000-8010`
- 全端口：`ALL`
- ICMP 协议时端口固定为 `ALL`，用户无需填写

---

## 三、Provider 接口与注册

### 3.1 Provider 接口

```go
type Provider interface {
    Name() string
    CloudType() CloudType
    GetRules() ([]RuleInfo, error)
    CreateRules(rules []RuleAction) error
    DeleteRules(rules []RuleAction) error
    ConvertPort(port string) string
    TargetIndex() int
}
```

**ConvertPort 职责：** 将统一端口格式（如 `80,443,8000-8010`）转换为对应云厂商的端口格式：
- 腾讯云 Lighthouse：`80,443,8000-8010`（保持原样）
- 腾讯云 CVM：`80,443,8000-8010`（保持原样）
- 阿里云轻量云：`80/80,443/443,8000/8010`（斜杠格式）
- 阿里云 ECS：`80/80,443/443,8000/8010`（斜杠格式）

### 3.2 工厂注册

```go
type Factory func(cfg TargetConfig) (Provider, error)

var registry = map[CloudType]Factory{}

func Register(cloudType CloudType, factory Factory) {
    registry[cloudType] = factory
}

func NewProvider(cfg TargetConfig) (Provider, error) {
    factory, ok := registry[cfg.CloudType]
    if !ok {
        return nil, fmt.Errorf("不支持的云产品类型: %s", cfg.CloudType)
    }
    return factory(cfg)
}
```

各 Provider 在 `init()` 中自注册。

### 3.3 通用 Diff 逻辑

```go
type ruleKey struct {
    protocol      string
    port          string
    cidrBlock     string
    ipv6CidrBlock string
    action        string
}

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

func Diff(
    resolved []dns.ResolvedIP,
    rule DomainRule,
    desc string,
    existing []RuleInfo,
    portConverter func(string) string,
) (toAdd []RuleAction, toDelete []RuleAction) {
    // 通用 diff 逻辑，与云厂商无关
}
```

---

## 四、Syncer 同步引擎

### 4.1 核心结构

```go
type Syncer struct {
    cfg        *config.Config
    providers  []provider.Provider
    resolver   *dns.Resolver
    configCh   chan *config.Config  // 热更新
    stopCh     chan struct{}
    wg         sync.WaitGroup
    clientPool *ClientPool
}

// ClientPool SDK Client 复用池
// 相同 cloudType + region + accessID 的 Target 共享同一个 SDK Client，
// 避免重复创建连接。Provider 创建时从 Pool 获取/复用 Client。
type ClientPool struct {
    mu      sync.Mutex
    clients map[string]any  // key: cloudType|region|accessID
}
```

**ClientPool 与 Provider 的关系：**
- Syncer 持有 ClientPool，负责生命周期管理
- Provider 工厂函数接收 ClientPool 参数，创建时从池中获取或新建 SDK Client
- 同一 cloudType + region + accessID 的多个 Target 复用同一个 Client

### 4.2 同步主循环

```go
func (s *Syncer) syncAll() {
    groups := s.groupByCloud()           // 按云厂商分组
    var groupWg sync.WaitGroup
    for cloudType, providers := range groups {
        groupWg.Add(1)
        go func(ct CloudType, ps []provider.Provider) {
            defer groupWg.Done()
            for _, p := range ps {
                rules := filterRulesForTarget(s.cfg.DomainRules, p.TargetIndex())
                for _, rule := range rules {
                    s.syncDomain(p, rule)
                    time.Sleep(s.rateLimitInterval(ct))
                }
            }
        }(cloudType, providers)
    }
    groupWg.Wait()
}
```

### 4.3 频率限制

```go
func (s *Syncer) rateLimitInterval(cloudType CloudType) time.Duration {
    switch cloudType {
    case CloudAliSWAS:
        return 800 * time.Millisecond  // 100次/60秒
    default:
        return 500 * time.Millisecond  // 腾讯云、阿里云 ECS
    }
}
```

### 4.4 乐观锁重试

- 写入前重新 Describe → 重新 diff → 重新 Create/Delete
- 最多 3 次，指数退避
- 不传入 FirewallVersion（由云 API 自行管理）

---

## 五、各云厂商 API 对比

| 操作 | 腾讯云 Lighthouse | 腾讯云 CVM | 阿里云轻量云 | 阿里云 ECS |
|------|-------------------|-----------|-------------|-----------|
| **查询** | `DescribeFirewallRules` | `DescribeSecurityGroupPolicies` | `ListFirewallRules` | `DescribeSecurityGroupAttribute` |
| **添加** | `CreateFirewallRules` | `CreateSecurityGroupPolicies` | `CreateFirewallRules` | `AuthorizeSecurityGroup` |
| **删除** | `DeleteFirewallRules` | `DeleteSecurityGroupPolicies` | `DeleteFirewallRules` | `RevokeSecurityGroup` |
| **Endpoint** | `lighthouse.tencentcloudapi.com` | `vpc.tencentcloudapi.com` | `swas.{region}.aliyuncs.com` | `ecs.{region}.aliyuncs.com` |
| **频率限制** | 10次/秒 | 100次/秒 | 100次/60秒 | 不限 |
| **规则标识字段** | `FirewallRuleDescription` | `PolicyDescription` | `Remark` | `Description` |
| **端口格式** | `80` 或 `443,80` 或 `ALL` | `80` 或 `8000-8010` | `80` 或 `1/200` | `80/80` 或 `1/200` |
| **IPv6 字段** | `Ipv6CidrBlock` | `Ipv6CidrBlock` | `SourceCidrIp`（待确认） | `Ipv6SourceCidrIp` |
| **操作对象** | `InstanceId` | `SecurityGroupId` | `InstanceId` | `SecurityGroupId` |

---

## 六、各 Provider 实现要点

### 6.1 腾讯云 Lighthouse

- 迁移现有 `firewall/client.go` 和 `firewall/rule.go`
- 保持 API 调用方式不变
- 规则标识：`FirewallRuleDescription` 以 `[TAG]` 开头
- **绝不**使用 `ModifyFirewallRules`

### 6.2 腾讯云 CVM 安全组

- SDK: `tencentcloud-sdk-go/tencentcloud/vpc/v20170312`
- 操作对象是 `SecurityGroupId`（非 InstanceId）
- 端口格式：单端口或范围，需在 ConvertPort 中转换逗号分隔
- 删除需传 `PolicyIndex`
- 规则描述：`PolicyDescription`

### 6.3 阿里云轻量云 (SWAS-OPEN)

- SDK: `alibaba-cloud-sdk-go/services/swas-open`（V1）
- 端口格式：`"80"` 或 `"1/200"`（斜杠分隔）
- 规则标识：`Remark`
- 频率限制严格：100次/60秒，建议 800ms 间隔

### 6.4 阿里云 ECS 安全组

- SDK: `alibaba-cloud-sdk-go/services/ecs`
- 操作对象是 `SecurityGroupId`
- 端口格式必须用斜杠：`"80/80"` 或 `"1/200"`
- IPv6 用独立字段 `Ipv6SourceCidrIp`
- 规则标识：`Description`（≤ 256 字节）

---

## 七、.env 配置格式（重构后）

```env
# 云资源目标（provider|resource_id|region）
TARGETS=tc_lighthouse|lhins-abc|ap-guangzhou, \
        tc_cvm|sg-def|ap-shanghai, \
        ali_swas|ace0706b|cn-hangzhou, \
        ali_ecs|sg-ghi|cn-shenzhen

# 云厂商凭据（按厂商分离，避免密钥泄露）
TC_ACCESS_ID=
TC_ACCESS_KEY=
ALI_ACCESS_ID=
ALI_ACCESS_KEY=

# 域名规则（host|protocol|ports|action|targets|comment）
RULES=api.example.com|TCP|443,80|ACCEPT||生产API, \
      vpn.example.com|UDP|1194|ACCEPT|2|VPN接入, \
      game.example.com|TCP|8000-8010|ACCEPT|1,3|游戏端口, \
      ping.example.com|ICMP|ALL|ACCEPT||允许Ping

# 全局设置
TAG=auto-dns
INTERVAL=5m
DNS=8.8.8.8:53
LOG_LEVEL=info
WEBUI_PORT=9090
```

### 变量名变更对照

| 旧变量 | 新变量 | 变化 |
|--------|--------|------|
| `TENCENTCLOUD_SECRET_ID` | `TC_ACCESS_ID` | 独立凭据 |
| `TENCENTCLOUD_SECRET_KEY` | `TC_ACCESS_KEY` | 独立凭据 |
| `LIGHTHOUSE_INSTANCE_ID` | 内嵌 `TARGETS` | 统一资源声明 |
| `LIGHTHOUSE_REGION` | 内嵌 `TARGETS` | 统一资源声明 |
| `DOMAIN_RULES` | `RULES` | 新增 targets + comment 列 |
| `RULE_TAG` | `TAG` | 精简 |
| `CHECK_INTERVAL` | `INTERVAL` | 精简 |
| `DNS_SERVER` | `DNS` | 精简 |

### 反斜杠续行

`.env` 不支持多行，解析器预处理：
```go
// 将 \ 续行的多行合并为单行
// TARGETS=xxx, \
//          yyy
// → TARGETS=xxx, yyy
```

TARGETS 和 RULES 均使用**逗号 `,`** 作为条目分隔符，反斜杠 `\` 用于视觉换行。解析顺序：先合并续行 → 再按逗号拆分条目 → 最后按 `|` 拆分字段。

---

## 八、构建体系

### 8.1 Build Tags 分离

```go
// systray.go
//go:build desktop

package app
// 系统托盘代码仅桌面端编译
```

### 8.2 构建命令

```bash
# Docker 构建（不含系统托盘，纯静态）
CGO_ENABLED=0 go build -tags docker -ldflags="-s -w -X main.version=$VERSION" .

# 桌面构建（含系统托盘，需要 CGO）
CGO_ENABLED=1 go build -tags desktop -ldflags="-s -w -X main.version=$VERSION" .
```

### 8.3 构建差异

| 维度 | Docker | 桌面 |
|------|--------|------|
| CGO | `CGO_ENABLED=0` | `CGO_ENABLED=1` |
| Build Tags | `-tags docker` | `-tags desktop` |
| 包含模块 | core + webui + provider | core + webui + provider + systray |
| 体积 | ~14MB（镜像内） | ~12-14MB |
| 交叉编译 | 仅 `linux/amd64` | 需各平台工具链 |

---

## 九、WebUI 技术实现

### 9.1 架构

```
main.go
├── 模式判定（determineMode）
├── HTTP Server (:9090)
│   ├── /api/targets             （云资源 CRUD）
│   ├── /api/rules               （规则 CRUD）
│   ├── /api/sync/status         （同步状态）
│   ├── /api/sync/dryrun         （试运行）
│   ├── /api/settings            （全局设置）
│   ├── /api/test-connection     （测试连接）
│   ├── /api/config/export       （配置导出）
│   ├── /api/config/import       （配置导入）
│   ├── /api/health              （健康检查）
│   └── /                        （Vue SPA，embed）
├── Syncer (goroutine)           （同步主循环）
│   └── configCh                 （热更新）
└── Config Store (SQLite)        （配置持久化）
```

### 9.2 配置热重载

```go
func (s *Syncer) Reload(newCfg *config.Config) {
    s.configCh <- newCfg
}
```

WebUI 修改配置 → 写入 SQLite → 发送新 Config 到 channel → Syncer 接收并重载。

### 9.3 首次运行体验（桌面端）

- 启动后自动打开 `http://localhost:{PORT}`
- 托盘气泡提示
- SQLite 存储路径：
  - macOS: `~/Library/Application Support/fwalizer/config.db`
  - Windows: `%APPDATA%/fwalizer/config.db`
  - Linux: `~/.config/fwalizer/config.db`

---

## 十、SDK 依赖

```
// go.mod
require (
    // 腾讯云
    github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.3.137
    github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse v1.3.108
    github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc v1.3.xxx    // 新增

    // 阿里云（V1，已终止支持，后续关注 V2 迁移）
    github.com/aliyun/alibaba-cloud-sdk-go v1.63.xxx                         // 新增

    // 系统托盘（仅桌面端）
    fyne.io/systray v1.11.0                                                  // 新增

    // WebUI 存储
    modernc.org/sqlite v1.xxx                                                // 新增
)
```

---

## 十一、事件总线

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

**投递语义：**
- 异步投递：事件通过 goroutine 分发，不阻塞同步主流程
- Subscriber 返回的 error 仅记录日志，不影响其他 Subscriber 或同步引擎
- 无重试机制（告警失败不应影响核心功能）

---

## 十二、补充技术细节

### 12.1 DNS 超时优化

- `net.Dialer` 连接超时：**10s**
- `context.WithTimeout` 整体超时：**10s**（通过 `DNS_TIMEOUT` 环境变量配置，默认 `10s`）
- 各域名 DNS 解析并行执行，提前收集结果后再 Diff

### 12.2 Docker HEALTHCHECK 升级

```dockerfile
# WebUI 模式
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q --spider http://localhost:9090/api/health || exit 1
```

- WebUI 模式：HTTP 端点检测（`/api/health`）
- `.env` 模式：`killall -0` 进程检测

### 12.3 同步日志持久化（SQLite）

| 字段 | 说明 |
|------|------|
| `timestamp` | 同步时间 |
| `target` | 目标云资源标识 |
| `domain` | 域名 |
| `result` | success / failed / skipped |
| `added` | 新增规则数 |
| `deleted` | 删除规则数 |
| `error` | 错误信息 |

保留最近 1000 条（可配置），超出自动清理。

### 12.4 版本号管理

```go
// version/version.go
package version

var Version = "dev"  // -ldflags "-X github.com/alcaprophet/fwalizer/version.Version=v1.0.0" 注入
```

### 12.5 CLI 子命令

| 命令 | 功能 |
|------|------|
| `fwalizer version` | 显示版本号 |
| `fwalizer validate` | 校验 .env 配置（格式、凭据有效性、资源可达性、DNS 服务器） |
| `fwalizer backup` | 备份 SQLite → `config.db.bak.{timestamp}` |
| `fwalizer restore [file]` | 从备份恢复 |

### 12.6 日志级别

| 级别 | 说明 |
|------|------|
| `debug` | DNS 结果、API 请求/响应、Diff 详情 |
| `info` | 同步状态、规则变更摘要（默认） |
| `warn` | DNS 失败、重试、规则已存在 |
| `error` | 写入失败、连接失败 |

### 12.7 DNS 渐进式熔断

```
1 次失败 → WARN 日志，保留现有规则
3 次失败 → ERROR 日志，连续失败告警
5 次失败 → 熔断（暂停该域名同步，不影响其他域名）
熔断后　 → 半开状态：每轮仍尝试一次解析
解析成功 → 解除熔断，恢复正常同步
```

- 熔断阈值可通过 `DNS_FAIL_THRESHOLD` 配置（默认 5）
- 熔断通过 EventBus 发送 `EventDNSFailed` 事件
- 半开探测失败不计入失败次数，维持熔断状态

### 12.8 SQLite 备份

```bash
fwalizer backup                    # 备份
fwalizer restore config.db.bak.1   # 恢复
fwalizer backup --list             # 列出备份
```

- 最多保留 5 个备份（自动轮转）
- 备份前 `PRAGMA integrity_check` 校验

### 12.9 Dry Run 机制

Dry Run 执行完整的同步流程直到 Diff 计算完成，但**不实际调用 CreateRules / DeleteRules**：

```
DNS 解析 → GetRules → OwnedRules → Diff → 返回结果（不写入）
```

- WebUI 端点：`POST /api/sync/dryrun`
- 返回 JSON：每个域名/目标的 `toAdd` 和 `toDelete` 规则列表
- 复用 Syncer 的 Diff 逻辑，仅跳过写入步骤
- 不触发 EventBus 事件

### 12.10 配置导入/导出

格式：**JSON**（与 WebUI API 一致）

```json
{
  "version": 1,
  "targets": [...],
  "rules": [...],
  "settings": {
    "tag": "auto-dns",
    "interval": "5m",
    "dns": "8.8.8.8:53",
    "log_level": "info"
  }
}
```

- 导出：`GET /api/config/export` → 下载 JSON 文件
- 导入：`POST /api/config/import` → 上传 JSON，校验后写入 SQLite
- 导入时校验格式完整性，凭据字段不导出（安全考虑）
- 导入成功后触发配置热重载

### 12.11 SQLite 并发策略

- 启用 **WAL 模式**（`PRAGMA journal_mode=WAL`），支持读写并发
- 写操作使用短事务，避免长时间持锁
- WebUI 写配置、Syncer 读配置、日志写入可并行
- 初始化时执行 `PRAGMA busy_timeout=5000` 避免锁等待超时

### 12.12 进程锁（防多实例）

- **WebUI 模式**：启动时创建 pidfile（与 SQLite 同目录），写入当前 PID
  - macOS/Linux：`~/.config/fwalizer/fwalizer.pid` 或对应标准路径
  - Windows：`%APPDATA%/fwalizer/fwalizer.pid`
  - 启动时检测 pidfile 是否存在且进程存活，是则拒绝启动
  - 正常退出时删除 pidfile
- **`.env` 模式**：无 pidfile（Docker 环境由容器编排保证单实例）

### 12.13 桌面端开机自启

仅 Windows 和 macOS 支持，作为 WebUI「全局设置」中的开关项，**默认关闭**。

**Windows：**
```go
// 写入注册表
key := `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
// 值: "FWAlizer" = "C:\path\to\fwalizer.exe"
// 关闭时删除该注册表项
```

**macOS：**
```xml
<!-- ~/Library/LaunchAgents/com.fwalizer.agent.plist -->
<plist>
  <dict>
    <key>Label</key><string>com.fwalizer.agent</string>
    <key>ProgramArguments</key>
    <array><string>/path/to/fwalizer</string></array>
    <key>RunAtLoad</key><true/>
  </dict>
</plist>
<!-- 启用: launchctl load plist -->
<!-- 禁用: launchctl unload plist + 删除文件 -->
```

**Linux：** 不提供内置支持，用户自行配置 systemd user unit 或其他方式。
