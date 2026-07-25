# FWAlizer 技术实现细节（Build1）

> 代码级别的实现参考。设计大纲见 [Design1.md](./Design1.md)，AI 编码指令见 [AGENTS.md](./AGENTS.md)。

---

## 构建进度追踪

| Step | 内容 | 状态 |
|------|------|------|
| 1 | 项目骨架 + 基础工具 | ☐ 未开始 |
| 2 | .env 解析器 + 配置校验 | ☐ 未开始 |
| 3 | DNS 解析器 | ☐ 未开始 |
| 4 | Provider 抽象层 | ☐ 未开始 |
| 5 | 腾讯云 Lighthouse Provider | ☐ 未开始 |
| 6 | Syncer 同步引擎 | ☐ 未开始 |
| 7 | App 生命周期 + CLI + main.go | ☐ 未开始 |
| 8 | 腾讯云 CVM Provider | ☐ 未开始 |
| 9 | 阿里云 SWAS Provider | ☐ 未开始 |
| 10 | 阿里云 ECS Provider | ☐ 未开始 |
| 11 | DNS 熔断 + 同步日志 | ☐ 未开始 |
| 12 | Docker 构建 + Makefile | ☐ 未开始 |
| 13 | WebUI 后端 | ☐ 未开始 |
| 14 | WebUI 前端 | ☐ 未开始 |
| 15 | 告警 + 高级功能 | ☐ 未开始 |
| 16 | 桌面端 | ☐ 未开始 |

> 状态标记：☐ 未开始 / ◧ 进行中 / ☑ 已完成

---

## 分步构建计划

> 原则：每一步完成后均可编译、可测试。不跳步，不并行多步。
> AI 执行指令：每次仅执行一个 Step，完成后运行验收命令，确认通过后再进入下一步。

---

### Step 1：项目骨架 + 基础工具

**目标：** 建立新目录结构，完成无外部依赖的基础模块。

**前置条件：** 无

**产出文件：**

#### 1.1 `go.mod`

```go
module github.com/alcaprophet/fwalizer

go 1.25
```

#### 1.2 `version/version.go`

```go
package version

// Version 由 ldflags 注入，开发时保持 "dev"
var Version = "dev"
```

#### 1.3 `internal/portconv/portconv.go`

```go
package portconv

import "strings"

// Parse 将统一端口格式拆分为独立端口/范围列表
// 输入: "80,443,8000-8010" → ["80", "443", "8000-8010"]
// 输入: "ALL" → ["ALL"]
func Parse(ports string) []string {
    ports = strings.TrimSpace(ports)
    if ports == "" || strings.EqualFold(ports, "ALL") {
        return []string{"ALL"}
    }
    parts := strings.Split(ports, ",")
    result := make([]string, 0, len(parts))
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p != "" {
            result = append(result, p)
        }
    }
    return result
}

// ToSlash 将单个端口/范围转为阿里云斜杠格式
// "80" → "80/80"，"8000-8010" → "8000/8010"，"ALL" → "-1/-1"
func ToSlash(port string) string {
    if strings.EqualFold(port, "ALL") {
        return "-1/-1"
    }
    if strings.Contains(port, "-") {
        parts := strings.SplitN(port, "-", 2)
        return parts[0] + "/" + parts[1]
    }
    return port + "/" + port
}
```

#### 1.4 `internal/tag/tag.go`

```go
package tag

import (
    "fmt"
    "strings"
)

// Format 生成规则描述："[TAG] comment"
func Format(tag, comment string) string {
    if comment == "" {
        return fmt.Sprintf("[%s]", tag)
    }
    return fmt.Sprintf("[%s] %s", tag, comment)
}

// HasPrefix 判断规则描述是否属于本工具管理
func HasPrefix(description, tag string) bool {
    return strings.HasPrefix(description, "["+tag+"]")
}

// Parse 从描述中提取 comment 部分，非本工具规则返回 ok=false
func Parse(description, tag string) (comment string, ok bool) {
    prefix := "[" + tag + "]"
    if !strings.HasPrefix(description, prefix) {
        return "", false
    }
    return strings.TrimSpace(description[len(prefix):]), true
}
```

#### 1.5 `config/config.go`

```go
package config

import "time"

type CloudType string

const (
    CloudTCLighthouse CloudType = "tc_lighthouse"
    CloudTCCVM       CloudType = "tc_cvm"
    CloudAliSWAS     CloudType = "ali_swas"
    CloudAliECS      CloudType = "ali_ecs"
)

type RuleInfo struct {
    Protocol      string // TCP / UDP / TCP+UDP / ICMP / ICMPv6 / ALL
    Port          string // 归一化为 "port" 或 "start-end" 或 "ALL"
    CidrBlock     string // IPv4 CIDR，如 "1.2.3.4/32"
    Ipv6CidrBlock string // IPv6 CIDR，如 "2001:db8::1/128"
    Action        string // ACCEPT / DROP
    Description   string // 规则描述/备注
    PolicyIndex   string // CVM 安全组删除时需要
    RuleID        string // 阿里云 SWAS/ECS 删除时需要
}

type RuleAction struct {
    Protocol      string
    Port          string // 已转换为对应云厂商的端口格式
    CidrBlock     string
    Ipv6CidrBlock string
    Action        string
    Description   string
}

type TargetConfig struct {
    CloudType  CloudType
    Region     string
    ResourceID string // InstanceId 或 SecurityGroupId
}

type DomainRule struct {
    Host     string
    Protocol string // TCP / UDP / TCP+UDP / ICMP
    Ports    string // 单端口、逗号分隔、范围、ALL
    Action   string // ACCEPT / DROP
    Targets  []int  // 目标编号（空 = 全部）
    Comment  string
}

type Config struct {
    TCAccessID       string
    TCAccessKey      string
    AliAccessID      string
    AliAccessKey     string
    Targets          []TargetConfig
    DomainRules      []DomainRule
    Tag              string
    Interval         time.Duration
    DNS              string
    DNSTimeout       time.Duration // 默认 10s
    DNSFailThreshold int           // 默认 5
    LogLevel         string        // debug / info / warn / error
    WebUIPort        int           // 默认 9090
    Mode             string        // env / webui / 空=自动
}
```

**约束：**
- 本步不引入任何第三方依赖
- `internal/` 下的包不依赖 `config/`（避免循环依赖）
- 所有注释使用中文

**验收：**
```bash
go build ./...
go test ./internal/... -v
```

---

### Step 2：.env 解析器 + 配置校验

**目标：** 能正确解析新格式 .env 并校验有效性。

**前置条件：** Step 1 完成

**产出文件：**

#### 2.1 `config/env.go`

```go
package config

import (
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"
)

// LoadEnv 从 .env 文件加载配置
func LoadEnv(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("读取 .env 失败: %w", err)
    }
    return ParseEnv(string(data))
}

// ParseEnv 解析 .env 内容
func ParseEnv(content string) (*Config, error) {
    // 1. 反斜杠续行合并
    content = mergeContinuation(content)
    // 2. 解析键值对
    kv := parseKeyValue(content)
    // 3. 构建 Config
    cfg := &Config{
        Tag:              getOr(kv, "TAG", "auto-dns"),
        DNS:              getOr(kv, "DNS", "8.8.8.8:53"),
        LogLevel:         getOr(kv, "LOG_LEVEL", "info"),
        DNSTimeout:       10 * time.Second,
        DNSFailThreshold: 5,
        WebUIPort:        9090,
        TCAccessID:       kv["TC_ACCESS_ID"],
        TCAccessKey:      kv["TC_ACCESS_KEY"],
        AliAccessID:      kv["ALI_ACCESS_ID"],
        AliAccessKey:     kv["ALI_ACCESS_KEY"],
        Mode:             kv["FWALIZER_MODE"],
    }
    // 解析可选值
    if v := kv["DNS_TIMEOUT"]; v != "" {
        d, err := time.ParseDuration(v)
        if err != nil { return nil, fmt.Errorf("DNS_TIMEOUT 格式错误: %w", err) }
        cfg.DNSTimeout = d
    }
    if v := kv["DNS_FAIL_THRESHOLD"]; v != "" {
        n, err := strconv.Atoi(v)
        if err != nil { return nil, fmt.Errorf("DNS_FAIL_THRESHOLD 必须为整数: %w", err) }
        cfg.DNSFailThreshold = n
    }
    if v := kv["WEBUI_PORT"]; v != "" {
        n, err := strconv.Atoi(v)
        if err != nil { return nil, fmt.Errorf("WEBUI_PORT 必须为整数: %w", err) }
        cfg.WebUIPort = n
    }
    if v := kv["INTERVAL"]; v != "" {
        d, err := time.ParseDuration(v)
        if err != nil { return nil, fmt.Errorf("INTERVAL 格式错误: %w", err) }
        cfg.Interval = d
    } else {
        cfg.Interval = 5 * time.Minute
    }
    // 解析 TARGETS
    if v := kv["TARGETS"]; v != "" {
        targets, err := parseTargets(v)
        if err != nil { return nil, err }
        cfg.Targets = targets
    }
    // 解析 RULES
    if v := kv["RULES"]; v != "" {
        rules, err := parseRules(v, len(cfg.Targets))
        if err != nil { return nil, err }
        cfg.DomainRules = rules
    }
    return cfg, nil
}

// mergeContinuation 将 `\` 续行合并为单行
func mergeContinuation(content string) string {
    // 将 "\<\n>" 及其后的前导空格替换为单空格
    lines := strings.Split(content, "\n")
    var merged []string
    for i := 0; i < len(lines); i++ {
        line := lines[i]
        for strings.HasSuffix(strings.TrimSpace(line), "\\") {
            line = strings.TrimSuffix(strings.TrimSpace(line), "\\")
            i++
            if i < len(lines) {
                line += " " + strings.TrimSpace(lines[i])
            }
        }
        merged = append(merged, line)
    }
    return strings.Join(merged, "\n")
}

// parseKeyValue 解析 KEY=VALUE，忽略注释和空行
func parseKeyValue(content string) map[string]string {
    kv := make(map[string]string)
    for _, line := range strings.Split(content, "\n") {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        idx := strings.Index(line, "=")
        if idx < 0 { continue }
        key := strings.TrimSpace(line[:idx])
        val := strings.TrimSpace(line[idx+1:])
        kv[key] = val
    }
    return kv
}

// parseTargets 解析 "provider|resource_id|region, ..."
func parseTargets(s string) ([]TargetConfig, error) {
    entries := splitEntries(s)
    var targets []TargetConfig
    for i, entry := range entries {
        parts := strings.Split(entry, "|")
        if len(parts) != 3 {
            return nil, fmt.Errorf("TARGETS 第 %d 项格式错误，应为 provider|resource_id|region", i+1)
        }
        ct := CloudType(strings.TrimSpace(parts[0]))
        switch ct {
        case CloudTCLighthouse, CloudTCCVM, CloudAliSWAS, CloudAliECS:
        default:
            return nil, fmt.Errorf("TARGETS 第 %d 项 provider 不合法: %s", i+1, parts[0])
        }
        targets = append(targets, TargetConfig{
            CloudType:  ct,
            ResourceID: strings.TrimSpace(parts[1]),
            Region:     strings.TrimSpace(parts[2]),
        })
    }
    return targets, nil
}

// parseRules 解析 "host|protocol|ports|action|targets|comment, ..."
func parseRules(s string, targetCount int) ([]DomainRule, error) {
    entries := splitEntries(s)
    var rules []DomainRule
    for i, entry := range entries {
        parts := strings.Split(entry, "|")
        if len(parts) < 4 || len(parts) > 6 {
            return nil, fmt.Errorf("RULES 第 %d 项格式错误，应为 host|protocol|ports|action[|targets[|comment]]", i+1)
        }
        rule := DomainRule{
            Host:     strings.TrimSpace(parts[0]),
            Protocol: strings.ToUpper(strings.TrimSpace(parts[1])),
            Ports:    strings.TrimSpace(parts[2]),
            Action:   strings.ToUpper(strings.TrimSpace(parts[3])),
        }
        // 协议校验
        switch rule.Protocol {
        case "TCP", "UDP", "TCP+UDP", "ICMP":
        default:
            return nil, fmt.Errorf("RULES 第 %d 项协议不合法: %s", i+1, rule.Protocol)
        }
        // ICMP 强制端口为 ALL
        if rule.Protocol == "ICMP" {
            rule.Ports = "ALL"
        }
        // Action 校验
        if rule.Action != "ACCEPT" && rule.Action != "DROP" {
            return nil, fmt.Errorf("RULES 第 %d 项 action 不合法: %s", i+1, rule.Action)
        }
        // targets 解析
        if len(parts) >= 5 {
            t := strings.TrimSpace(parts[4])
            if t != "" && t != "*" {
                nums, err := parseTargetNums(t, targetCount)
                if err != nil {
                    return nil, fmt.Errorf("RULES 第 %d 项: %w", i+1, err)
                }
                rule.Targets = nums
            }
        }
        // comment
        if len(parts) >= 6 {
            rule.Comment = strings.TrimSpace(parts[5])
        }
        rules = append(rules, rule)
    }
    return rules, nil
}

// splitEntries 按逗号拆分条目（忽略尾部空格）
func splitEntries(s string) []string {
    parts := strings.Split(s, ",")
    var result []string
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p != "" {
            result = append(result, p)
        }
    }
    return result
}

// parseTargetNums 解析 "1,3" → []int{1,3}，并校验范围
func parseTargetNums(s string, max int) ([]int, error) {
    parts := strings.Split(s, ",")
    var nums []int
    for _, p := range parts {
        n, err := strconv.Atoi(strings.TrimSpace(p))
        if err != nil {
            return nil, fmt.Errorf("targets 编号不合法: %s", p)
        }
        if n < 1 || n > max {
            return nil, fmt.Errorf("targets 编号 %d 超出范围 [1,%d]", n, max)
        }
        nums = append(nums, n)
    }
    return nums, nil
}

func getOr(kv map[string]string, key, def string) string {
    if v := kv[key]; v != "" { return v }
    return def
}
```

#### 2.2 `config/validate.go`

```go
package config

import "fmt"

// Validate 校验配置完整性
func (c *Config) Validate() error {
    if len(c.Targets) == 0 {
        return fmt.Errorf("TARGETS 不能为空")
    }
    if len(c.DomainRules) == 0 {
        return fmt.Errorf("RULES 不能为空")
    }
    // 检查每个 Target 是否有对应凭据
    for i, t := range c.Targets {
        switch t.CloudType {
        case CloudTCLighthouse, CloudTCCVM:
            if c.TCAccessID == "" || c.TCAccessKey == "" {
                return fmt.Errorf("TARGETS[%d] 为腾讯云，但 TC_ACCESS_ID/TC_ACCESS_KEY 未设置", i+1)
            }
        case CloudAliSWAS, CloudAliECS:
            if c.AliAccessID == "" || c.AliAccessKey == "" {
                return fmt.Errorf("TARGETS[%d] 为阿里云，但 ALI_ACCESS_ID/ALI_ACCESS_KEY 未设置", i+1)
            }
        }
        if t.ResourceID == "" {
            return fmt.Errorf("TARGETS[%d] resource_id 不能为空", i+1)
        }
        if t.Region == "" {
            return fmt.Errorf("TARGETS[%d] region 不能为空", i+1)
        }
    }
    return nil
}
```

**约束：**
- 解析器不依赖网络，纯字符串处理
- 错误信息必须包含行号/索引，便于用户定位
- 注释和空行必须忽略
- `#` 开头的行是注释，但 `KEY=value # comment` 不做行内注释截断

**验收：**
```bash
go test ./config/... -v
```
测试用例必须覆盖：正常解析、续行合并、缺少字段、协议不合法、编号越界、空文件。

---

### Step 3：DNS 解析器

**目标：** 支持自定义 DNS + A/AAAA 解析 + 超时控制。

**前置条件：** Step 1 完成（使用 config.Config 中的 DNS/DNSTimeout）

**产出文件：**

#### 3.1 `dns/resolver.go`

```go
package dns

import (
    "context"
    "fmt"
    "net"
    "time"
)

// ResolvedIP 解析结果
type ResolvedIP struct {
    IP     net.IP
    IsIPv6 bool
}

// CIDR 返回带掩码的 CIDR 格式
func (r ResolvedIP) CIDR() string {
    if r.IsIPv6 {
        return r.IP.String() + "/128"
    }
    return r.IP.String() + "/32"
}

// Resolver 自定义 DNS 解析器
type Resolver struct {
    resolver *net.Resolver
    timeout  time.Duration
}

// NewResolver 创建解析器
// dnsAddr 格式: "8.8.8.8:53" 或 "8.8.8.8"（默认补 :53）
func NewResolver(dnsAddr string, timeout time.Duration) *Resolver {
    if !hasPort(dnsAddr) {
        dnsAddr = dnsAddr + ":53"
    }
    r := &net.Resolver{
        PreferGo: true,
        Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
            d := net.Dialer{Timeout: timeout}
            return d.DialContext(ctx, "udp", dnsAddr)
        },
    }
    return &Resolver{resolver: r, timeout: timeout}
}

// Resolve 同时解析 A 和 AAAA 记录
func (r *Resolver) Resolve(ctx context.Context, host string) ([]ResolvedIP, error) {
    ctx, cancel := context.WithTimeout(ctx, r.timeout)
    defer cancel()

    var results []ResolvedIP

    // A 记录
    addrs, err := r.resolver.LookupIPAddr(ctx, host)
    if err != nil {
        return nil, fmt.Errorf("DNS 解析失败 %s: %w", host, err)
    }
    for _, addr := range addrs {
        isV6 := addr.IP.To4() == nil
        results = append(results, ResolvedIP{IP: addr.IP, IsIPv6: isV6})
    }

    if len(results) == 0 {
        return nil, fmt.Errorf("DNS 解析无结果: %s", host)
    }
    return results, nil
}

func hasPort(addr string) bool {
    _, _, err := net.SplitHostPort(addr)
    return err == nil
}
```

**约束：**
- 仅支持 UDP 协议查询 DNS（简单场景足够）
- 超时同时控制连接和整体（用同一个 timeout）
- 返回结果不去重（由调用方处理）
- 本步不包含熔断逻辑（Step 11 再加）

**验收：**
```bash
go test ./dns/... -v
```
测试用例：解析 `localhost`、解析不存在的域名、超时场景（用无效 DNS 地址）。

---

### Step 4：Provider 抽象层

**目标：** 定义接口、工厂、通用 Diff 逻辑、ClientPool。

**前置条件：** Step 1 完成

**产出文件：**

#### 4.1 `provider/provider.go`

```go
package provider

import (
    "github.com/alcaprophet/fwalizer/config"
    "github.com/alcaprophet/fwalizer/dns"
)

// Provider 多云抽象接口
type Provider interface {
    // Name 返回可读名称，如 "tc_lighthouse(lhins-abc)"
    Name() string
    // CloudType 返回云产品类型
    CloudType() config.CloudType
    // GetRules 查询当前所有规则
    GetRules() ([]config.RuleInfo, error)
    // CreateRules 增量添加规则
    CreateRules(rules []config.RuleAction) error
    // DeleteRules 精确删除规则（传入 RuleInfo 因需要 RuleID/PolicyIndex）
    DeleteRules(rules []config.RuleInfo) error
    // ConvertPorts 统一端口 → 云厂商格式列表
    ConvertPorts(port string) []string
    // TargetIndex 返回在 TARGETS 中的索引（0-based）
    TargetIndex() int
}

// DiffResult Diff 计算结果
type DiffResult struct {
    ToAdd    []config.RuleAction
    ToDelete []config.RuleInfo
}

// SyncDomain 单个域名的同步结果
type SyncDomainResult struct {
    Domain  string
    Target  string
    Added   int
    Deleted int
    Error   error
}

// ResolvedIPs 便捷别名
type ResolvedIPs = []dns.ResolvedIP
```

#### 4.2 `provider/registry.go`

```go
package provider

import (
    "fmt"
    "sync"

    "github.com/alcaprophet/fwalizer/config"
)

// Factory 创建 Provider 的工厂函数
type Factory func(cfg config.TargetConfig, index int, pool *ClientPool) (Provider, error)

var (
    mu       sync.RWMutex
    registry = map[config.CloudType]Factory{}
)

// Register 注册 Provider 工厂（在各 Provider 的 init() 中调用）
func Register(ct config.CloudType, f Factory) {
    mu.Lock()
    defer mu.Unlock()
    registry[ct] = f
}

// NewProvider 创建 Provider 实例
func NewProvider(cfg config.TargetConfig, index int, pool *ClientPool) (Provider, error) {
    mu.RLock()
    f, ok := registry[cfg.CloudType]
    mu.RUnlock()
    if !ok {
        return nil, fmt.Errorf("不支持的云产品类型: %s", cfg.CloudType)
    }
    return f(cfg, index, pool)
}
```

#### 4.3 `provider/common.go`

```go
package provider

import (
    "strings"
    "sync"

    "github.com/alcaprophet/fwalizer/config"
    "github.com/alcaprophet/fwalizer/dns"
    "github.com/alcaprophet/fwalizer/internal/tag"
)

// OwnedRules 筛选本工具管理的规则（描述以 [TAG] 开头）
func OwnedRules(allRules []config.RuleInfo, tagStr string) []config.RuleInfo {
    var owned []config.RuleInfo
    for _, r := range allRules {
        if tag.HasPrefix(r.Description, tagStr) {
            owned = append(owned, r)
        }
    }
    return owned
}

// ruleKey 用于比较规则是否相同
type ruleKey struct {
    protocol      string
    port          string
    cidrBlock     string
    ipv6CidrBlock string
    action        string
}

func keyOf(r config.RuleInfo) ruleKey {
    return ruleKey{
        protocol:      strings.ToUpper(r.Protocol),
        port:          strings.ToUpper(r.Port),
        cidrBlock:     r.CidrBlock,
        ipv6CidrBlock: r.Ipv6CidrBlock,
        action:        strings.ToUpper(r.Action),
    }
}

func keyOfAction(r config.RuleAction) ruleKey {
    return ruleKey{
        protocol:      strings.ToUpper(r.Protocol),
        port:          strings.ToUpper(r.Port),
        cidrBlock:     r.CidrBlock,
        ipv6CidrBlock: r.Ipv6CidrBlock,
        action:        strings.ToUpper(r.Action),
    }
}

// Diff 计算需要添加和删除的规则
// resolved: DNS 解析结果
// rule: 域名规则配置
// desc: 规则描述（已包含 [TAG]）
// existing: 当前云端属于本工具的规则
// p: Provider（用于 ConvertPorts）
func Diff(
    resolved []dns.ResolvedIP,
    rule config.DomainRule,
    desc string,
    existing []config.RuleInfo,
    p Provider,
) DiffResult {
    // 1. 构建期望规则集
    desired := buildDesired(resolved, rule, desc, p)

    // 2. 构建现有规则索引
    existingKeys := make(map[ruleKey]config.RuleInfo)
    for _, r := range existing {
        existingKeys[keyOf(r)] = r
    }

    // 3. 计算 toAdd：期望中有、现有中无
    var toAdd []config.RuleAction
    desiredKeys := make(map[ruleKey]bool)
    for _, d := range desired {
        k := keyOfAction(d)
        desiredKeys[k] = true
        if _, exists := existingKeys[k]; !exists {
            toAdd = append(toAdd, d)
        }
    }

    // 4. 计算 toDelete：现有中有、期望中无
    var toDelete []config.RuleInfo
    for _, r := range existing {
        k := keyOf(r)
        if !desiredKeys[k] {
            toDelete = append(toDelete, r)
        }
    }

    return DiffResult{ToAdd: toAdd, ToDelete: toDelete}
}

// buildDesired 根据 DNS 结果和规则配置构建期望规则列表
func buildDesired(
    resolved []dns.ResolvedIP,
    rule config.DomainRule,
    desc string,
    p Provider,
) []config.RuleAction {
    var actions []config.RuleAction
    ports := p.ConvertPorts(rule.Ports)

    for _, ip := range resolved {
        for _, port := range ports {
            action := config.RuleAction{
                Protocol:    rule.Protocol,
                Port:        port,
                Action:      rule.Action,
                Description: desc,
            }
            if ip.IsIPv6 {
                action.Ipv6CidrBlock = ip.CIDR()
            } else {
                action.CidrBlock = ip.CIDR()
            }
            actions = append(actions, action)
        }
    }
    return actions
}

// ClientPool SDK Client 复用池
type ClientPool struct {
    mu      sync.Mutex
    clients map[string]any // key: cloudType|region|accessID
}

func NewClientPool() *ClientPool {
    return &ClientPool{clients: make(map[string]any)}
}

// GetOrCreate 获取或创建 Client
func (p *ClientPool) GetOrCreate(key string, create func() (any, error)) (any, error) {
    p.mu.Lock()
    defer p.mu.Unlock()
    if c, ok := p.clients[key]; ok {
        return c, nil
    }
    c, err := create()
    if err != nil {
        return nil, err
    }
    p.clients[key] = c
    return c, nil
}
```

**约束：**
- Diff 逻辑与云厂商无关，纯内存计算
- ruleKey 比较时 Protocol/Port/Action 统一转大写
- ClientPool 的 key 格式：`tc_lighthouse|ap-guangzhou|AKIDxxxx`

**验收：**
```bash
go test ./provider/... -v
```
测试用例：Diff 计算（新增、删除、无变化）、OwnedRules 筛选、ClientPool 复用。

---

### Step 5：腾讯云 Lighthouse Provider

**目标：** 第一个完整可用的 Provider。

**前置条件：** Step 4 完成

**产出文件：** `provider/tc_lighthouse.go`

**依赖添加：**
```bash
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse
```

**实现要点：**

```go
package provider

import (
    "github.com/alcaprophet/fwalizer/config"
    lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
    "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
    "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

func init() {
    Register(config.CloudTCLighthouse, newTCLighthouse)
}

type TCLighthouse struct {
    client      *lighthouse.Client
    instanceID  string
    targetIndex int
}

func newTCLighthouse(cfg config.TargetConfig, index int, pool *ClientPool) (Provider, error) {
    // 从 pool 获取或创建 client
    // key = "tc_lighthouse|" + cfg.Region + "|" + accessID
    // ...
}
```

**各方法实现要点：**

| 方法 | API | 关键细节 |
|------|-----|----------|
| `GetRules()` | `DescribeFirewallRules` | 分页查询（Limit=100），映射 FirewallRuleDescription→Description |
| `CreateRules()` | `CreateFirewallRules` | IPv6 规则用 Ipv6CidrBlock 字段，与 CidrBlock 分两条；ICMP+IPv6 用 ICMPv6 |
| `DeleteRules()` | `DeleteFirewallRules` | 捕获 `ResourceNotFound.FirewallRulesNotFound` 视为成功 |
| `ConvertPorts()` | 无 | 直接返回 `[port]`（Lighthouse 支持逗号分隔） |

**描述字段限制：** `FirewallRuleDescription` ≤ 64 字符，生成时需截断。

**验收：**
```bash
go build ./...
# 可选：配合真实凭据测试
./fwalizer validate
```

---

### Step 6：Syncer 同步引擎

**目标：** 完成同步主循环，能端到端跑通单 Provider。

**前置条件：** Step 3 + Step 5 完成

**产出文件：**

#### 6.1 `syncer/syncer.go`

```go
package syncer

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"

    "github.com/alcaprophet/fwalizer/config"
    "github.com/alcaprophet/fwalizer/dns"
    "github.com/alcaprophet/fwalizer/internal/tag"
    "github.com/alcaprophet/fwalizer/provider"
)

type Syncer struct {
    cfg        *config.Config
    providers  []provider.Provider
    resolver   *dns.Resolver
    configCh   chan *config.Config
    stopCh     chan struct{}
    wg         sync.WaitGroup
}

func New(cfg *config.Config, providers []provider.Provider, resolver *dns.Resolver) *Syncer {
    return &Syncer{
        cfg:       cfg,
        providers: providers,
        resolver:  resolver,
        configCh:  make(chan *config.Config, 1),
        stopCh:    make(chan struct{}),
    }
}

// Run 启动同步主循环（阻塞，直到收到停止信号）
func (s *Syncer) Run() {
    ticker := time.NewTicker(s.cfg.Interval)
    defer ticker.Stop()

    // 启动时立即执行一次
    s.syncAll()

    for {
        select {
        case <-ticker.C:
            s.syncAll()
        case newCfg := <-s.configCh:
            slog.Info("配置热重载")
            s.cfg = newCfg
            ticker.Reset(newCfg.Interval)
        case <-s.stopCh:
            slog.Info("同步引擎停止")
            return
        }
    }
}

// Stop 优雅停止
func (s *Syncer) Stop() {
    close(s.stopCh)
}

// Reload 热重载配置
func (s *Syncer) Reload(cfg *config.Config) {
    s.configCh <- cfg
}

// syncAll 执行一轮完整同步
func (s *Syncer) syncAll() {
    slog.Info("开始同步", "targets", len(s.providers), "rules", len(s.cfg.DomainRules))
    start := time.Now()

    // 按云厂商分组，跨云并行
    groups := s.groupByCloud()
    var wg sync.WaitGroup
    for ct, providers := range groups {
        wg.Add(1)
        go func(ct config.CloudType, ps []provider.Provider) {
            defer wg.Done()
            for _, p := range ps {
                rules := filterRulesForTarget(s.cfg.DomainRules, p.TargetIndex())
                for _, rule := range rules {
                    s.syncDomain(p, rule)
                    time.Sleep(rateLimitInterval(ct))
                }
            }
        }(ct, providers)
    }
    wg.Wait()

    slog.Info("同步完成", "耗时", time.Since(start).Round(time.Millisecond))
}

// syncDomain 同步单个域名到单个 Provider
func (s *Syncer) syncDomain(p provider.Provider, rule config.DomainRule) {
    // 1. DNS 解析
    resolved, err := s.resolver.Resolve(context.Background(), rule.Host)
    if err != nil {
        slog.Warn("DNS 解析失败，保留现有规则", "domain", rule.Host, "error", err)
        return
    }

    // 2. 获取当前规则
    allRules, err := p.GetRules()
    if err != nil {
        slog.Error("获取规则失败", "provider", p.Name(), "error", err)
        return
    }

    // 3. 筛选本工具的规则
    owned := provider.OwnedRules(allRules, s.cfg.Tag)

    // 4. 计算 Diff
    desc := tag.Format(s.cfg.Tag, rule.Comment)
    diff := provider.Diff(resolved, rule, desc, owned, p)

    // 5. 执行删除
    if len(diff.ToDelete) > 0 {
        if err := s.retryDelete(p, diff.ToDelete); err != nil {
            slog.Error("删除规则失败", "provider", p.Name(), "domain", rule.Host, "error", err)
            return
        }
        slog.Info("删除规则", "provider", p.Name(), "domain", rule.Host, "count", len(diff.ToDelete))
    }

    // 6. 执行添加
    if len(diff.ToAdd) > 0 {
        if err := s.retryCreate(p, diff.ToAdd); err != nil {
            slog.Error("添加规则失败", "provider", p.Name(), "domain", rule.Host, "error", err)
            return
        }
        slog.Info("添加规则", "provider", p.Name(), "domain", rule.Host, "count", len(diff.ToAdd))
    }
}

func (s *Syncer) groupByCloud() map[config.CloudType][]provider.Provider {
    groups := make(map[config.CloudType][]provider.Provider)
    for _, p := range s.providers {
        ct := p.CloudType()
        groups[ct] = append(groups[ct], p)
    }
    return groups
}

func filterRulesForTarget(rules []config.DomainRule, targetIndex int) []config.DomainRule {
    var filtered []config.DomainRule
    for _, r := range rules {
        if len(r.Targets) == 0 {
            filtered = append(filtered, r) // 空 = 所有目标
            continue
        }
        for _, t := range r.Targets {
            if t == targetIndex+1 { // Targets 是 1-based
                filtered = append(filtered, r)
                break
            }
        }
    }
    return filtered
}

// WaitForSignal 等待停止信号
func WaitForSignal(s *Syncer) {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
    <-sigCh
    slog.Info("收到停止信号，等待当前轮次完成...")
    s.Stop()
}
```

#### 6.2 `syncer/retry.go`

```go
package syncer

import (
    "log/slog"
    "time"

    "github.com/alcaprophet/fwalizer/config"
    "github.com/alcaprophet/fwalizer/provider"
)

const maxRetries = 3

// retryCreate 带重试的创建
func (s *Syncer) retryCreate(p provider.Provider, rules []config.RuleAction) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = p.CreateRules(rules)
        if err == nil {
            return nil
        }
        if isIdempotentCreate(err) {
            slog.Warn("规则已存在，跳过", "provider", p.Name())
            return nil
        }
        if !isRetryable(err) {
            return err
        }
        backoff := time.Duration(1<<uint(i)) * time.Second
        slog.Warn("创建失败，重试", "attempt", i+1, "backoff", backoff, "error", err)
        time.Sleep(backoff)
    }
    return err
}

// retryDelete 带重试的删除
func (s *Syncer) retryDelete(p provider.Provider, rules []config.RuleInfo) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = p.DeleteRules(rules)
        if err == nil {
            return nil
        }
        if isIdempotentDelete(err) {
            slog.Warn("规则已不存在，视为成功", "provider", p.Name())
            return nil
        }
        if !isRetryable(err) {
            return err
        }
        backoff := time.Duration(1<<uint(i)) * time.Second
        slog.Warn("删除失败，重试", "attempt", i+1, "backoff", backoff, "error", err)
        time.Sleep(backoff)
    }
    return err
}

// isRetryable 判断是否可重试
func isRetryable(err error) bool {
    // 网络超时、频率限制、服务端错误、防火墙忙 → 可重试
    // 参数错误、权限错误 → 不可重试
    msg := err.Error()
    retryable := []string{
        "RequestLimitExceeded",
        "InternalError",
        "FirewallBusy",
        "timeout",
        "connection refused",
    }
    for _, r := range retryable {
        if contains(msg, r) {
            return true
        }
    }
    return false
}

// isIdempotentCreate 判断“规则已存在”
func isIdempotentCreate(err error) bool {
    msg := err.Error()
    return contains(msg, "FirewallRulesExist") ||
        contains(msg, "FirewallRuleAlreadyExist")
}

// isIdempotentDelete 判断“规则已不存在”
func isIdempotentDelete(err error) bool {
    msg := err.Error()
    return contains(msg, "FirewallRulesNotFound") ||
        contains(msg, "SecurityGroupRuleId.NotFound") ||
        contains(msg, "SecurityGroupRule.RuleNotExist") ||
        contains(msg, "InvalidInstanceId.NotFound")
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, sub string) bool {
    for i := 0; i <= len(s)-len(sub); i++ {
        if s[i:i+len(sub)] == sub {
            return true
        }
    }
    return false
}
```

#### 6.3 `syncer/ratelimit.go`

```go
package syncer

import (
    "time"

    "github.com/alcaprophet/fwalizer/config"
)

// rateLimitInterval 根据云厂商返回请求间隔
func rateLimitInterval(ct config.CloudType) time.Duration {
    switch ct {
    case config.CloudAliSWAS:
        return 800 * time.Millisecond // 100次/60秒
    case config.CloudTCLighthouse:
        return 200 * time.Millisecond // 10次/秒
    default:
        return 100 * time.Millisecond // CVM 50次/秒、ECS 无限制
    }
}
```

**约束：**
- syncDomain 中 DNS 失败不删除现有规则（仅 WARN 日志）
- 重试时不重新 Describe（当前设计），仅在下一轮 syncAll 时重新获取
- 优雅退出：收到信号后等待当前 syncAll 完成

**验收：**
```bash
go build ./...
# 配合 .env + Lighthouse 凭据实际运行
./fwalizer
```

---

### Step 7：App 生命周期 + CLI + main.go

**目标：** 可执行的完整二进制（.env 模式）。

**前置条件：** Step 6 完成

**产出文件：**

#### 7.1 `app/mode.go`

```go
package app

import "os"

// Mode 运行模式
type Mode string

const (
    ModeEnv   Mode = "env"   // .env 文件驱动，无 WebUI
    ModeWebUI Mode = "webui" // SQLite + WebUI
)

// DetectMode 检测运行模式
func DetectMode(forced string) Mode {
    if forced == "env" || forced == "webui" {
        return Mode(forced)
    }
    // 自动检测：TARGETS 环境变量存在 → env 模式
    if os.Getenv("TARGETS") != "" {
        return ModeEnv
    }
    return ModeWebUI
}
```

#### 7.2 `app/app.go`

```go
package app

import (
    "fmt"
    "log/slog"

    "github.com/alcaprophet/fwalizer/config"
    "github.com/alcaprophet/fwalizer/dns"
    "github.com/alcaprophet/fwalizer/provider"
    "github.com/alcaprophet/fwalizer/syncer"
)

// Run 应用主入口
func Run(cfg *config.Config, mode Mode) error {
    // 1. 初始化日志
    initLogger(cfg.LogLevel)

    // 2. 校验配置
    if err := cfg.Validate(); err != nil {
        return fmt.Errorf("配置校验失败: %w", err)
    }

    // 3. 创建 ClientPool
    pool := provider.NewClientPool()

    // 4. 创建 Providers
    var providers []provider.Provider
    for i, t := range cfg.Targets {
        p, err := provider.NewProvider(t, i, pool)
        if err != nil {
            return fmt.Errorf("创建 Provider 失败 [%s]: %w", t.ResourceID, err)
        }
        providers = append(providers, p)
    }

    // 5. 创建 DNS Resolver
    resolver := dns.NewResolver(cfg.DNS, cfg.DNSTimeout)

    // 6. 创建 Syncer 并启动
    s := syncer.New(cfg, providers, resolver)
    go s.Run()

    // 7. 等待停止信号
    syncer.WaitForSignal(s)
    return nil
}

func initLogger(level string) {
    var lvl slog.Level
    switch level {
    case "debug": lvl = slog.LevelDebug
    case "warn":  lvl = slog.LevelWarn
    case "error": lvl = slog.LevelError
    default:      lvl = slog.LevelInfo
    }
    slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})))
}
```

#### 7.3 `app/cli.go`

```go
package app

import (
    "fmt"
    "os"

    "github.com/alcaprophet/fwalizer/config"
    "github.com/alcaprophet/fwalizer/version"
)

// RunCLI 处理子命令，返回 true 表示已处理（不需进入主流程）
func RunCLI(args []string) bool {
    if len(args) < 2 {
        return false
    }
    switch args[1] {
    case "version":
        fmt.Printf("fwalizer %s\n", version.Version)
        return true
    case "validate":
        path := ".env"
        if len(args) >= 3 {
            path = args[2]
        }
        cfg, err := config.LoadEnv(path)
        if err != nil {
            fmt.Fprintf(os.Stderr, "解析失败: %v\n", err)
            os.Exit(1)
        }
        if err := cfg.Validate(); err != nil {
            fmt.Fprintf(os.Stderr, "校验失败: %v\n", err)
            os.Exit(1)
        }
        fmt.Printf("配置有效: %d 个目标, %d 条规则\n", len(cfg.Targets), len(cfg.DomainRules))
        return true
    }
    return false
}
```

#### 7.4 `main.go`

```go
package main

import (
    "fmt"
    "os"

    "github.com/alcaprophet/fwalizer/app"
    "github.com/alcaprophet/fwalizer/config"
)

func main() {
    // CLI 子命令优先
    if app.RunCLI(os.Args) {
        return
    }

    // 检测模式
    cfg := &config.Config{}
    mode := app.DetectMode(os.Getenv("FWALIZER_MODE"))

    switch mode {
    case app.ModeEnv:
        // 从 .env 加载
        var err error
        cfg, err = config.LoadEnv(".env")
        if err != nil {
            fmt.Fprintf(os.Stderr, "加载 .env 失败: %v\n", err)
            os.Exit(1)
        }
    case app.ModeWebUI:
        // Phase 2 实现，当前占位
        fmt.Fprintln(os.Stderr, "WebUI 模式尚未实现，请使用 .env 模式")
        os.Exit(1)
    }

    if err := app.Run(cfg, mode); err != nil {
        fmt.Fprintf(os.Stderr, "运行失败: %v\n", err)
        os.Exit(1)
    }
}
```

**约束：**
- main.go 仅做分发，不包含业务逻辑
- .env 模式下不启动 WebUI、不写 SQLite
- WebUI 模式在本步仅占位，Step 13 实现

**验收：**
```bash
go build -o fwalizer .
./fwalizer version          # 输出 "fwalizer dev"
./fwalizer validate .env    # 校验配置
./fwalizer                  # 启动同步（需要真实凭据）
```

---

### Step 8：腾讯云 CVM Provider

**目标：** 支持腾讯云 CVM 安全组。

**前置条件：** Step 7 完成

**产出文件：** `provider/tc_cvm.go`

**依赖添加：**
```bash
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc
```

**实现要点：**

| 方法 | API | 关键细节 |
|------|-----|----------|
| `GetRules()` | `DescribeSecurityGroupPolicies` | 只取 Ingress 部分；映射 PolicyDescription→Description、PolicyIndex |
| `CreateRules()` | `CreateSecurityGroupPolicies` | 只写 Ingress；每条规则单独一个 Port；检查规则总数≤100 |
| `DeleteRules()` | `DeleteSecurityGroupPolicies` | 用 PolicyIndex 删除；只删 Ingress |
| `ConvertPorts()` | 无 | `portconv.Parse(port)` 拆分为多条 |

**特殊约束：**
- 一次请求只能操作单方向（Ingress）
- 规则总数上限 100 条（入站+出站），CreateRules 前检查当前 Ingress 数量
- 删除时传 `PolicyIndex` 字段（从 GetRules 查询结果中获取）
- 不传 Version 参数

**验收：**
```bash
go build ./...
go vet ./...
```

---

### Step 9：阿里云 SWAS Provider

**目标：** 支持阿里云轻量应用服务器。

**前置条件：** Step 7 完成

**产出文件：** `provider/ali_swas.go`

**依赖添加：**
```bash
go get github.com/alibabacloud-go/swas-open-20200601/v3
go get github.com/alibabacloud-go/darabonba-openapi/v2
go get github.com/aliyun/credentials-go
```

**实现要点：**

| 方法 | API | 关键细节 |
|------|-----|----------|
| `GetRules()` | `ListFirewallRules` | 分页（PageSize=100）；映射 Remark→Description、RuleId→RuleID、RuleProtocol→Protocol |
| `CreateRules()` | `CreateFirewallRules` | 批量创建；SourceCidrIp 填 IP；Remark 填描述 |
| `DeleteRules()` | `DeleteFirewallRules` | 传入 RuleIds 列表（从 RuleInfo.RuleID 获取） |
| `ConvertPorts()` | 无 | `portconv.ToSlash(port)` 转斜杠格式 |

**特殊约束：**
- 仅支持 IPv4，IPv6 解析结果跳过（记录 WARN 日志）
- ICMP 端口用 `-1/-1`
- 协议字段名 `RuleProtocol`，取值 TCP/UDP/TCP+UDP/ICMP
- 频率限制严格，Syncer 已配置 800ms 间隔
- Endpoint: `swas.{region}.aliyuncs.com`

**验收：**
```bash
go build ./...
go vet ./...
```

---

### Step 10：阿里云 ECS Provider

**目标：** 支持阿里云 ECS 安全组。

**前置条件：** Step 7 完成

**产出文件：** `provider/ali_ecs.go`

**依赖添加：**
```bash
go get github.com/alibabacloud-go/ecs-20140526/v7
```

**实现要点：**

| 方法 | API | 关键细节 |
|------|-----|----------|
| `GetRules()` | `DescribeSecurityGroupAttribute` | Direction=ingress；映射 Description、SecurityGroupRuleId→RuleID、PortRange→Port |
| `CreateRules()` | `AuthorizeSecurityGroup` | Permissions 数组；IPv4 用 SourceCidrIp，IPv6 用 Ipv6SourceCidrIp（互斥）；Priority=1 |
| `DeleteRules()` | `RevokeSecurityGroup` | 用 SecurityGroupRuleId 数组删除（推荐方式） |
| `ConvertPorts()` | 无 | `portconv.ToSlash(port)` 转斜杠格式 |

**特殊约束：**
- IPv6 和 IPv4 不可同时设置，必须分两条规则
- ICMP 端口用 `-1/-1`
- Description 字段 1~512 字符
- 规则已存在时 API 调用成功但不重复添加（无需特殊处理）
- 删除不存在的规则会返回错误码，已在 retry.go 中处理
- Endpoint: `ecs.{region}.aliyuncs.com`

**验收：**
```bash
go build ./...
go vet ./...
go test ./... 
```
四个 Provider 全部编译通过，无警告。

---

### Step 11：DNS 熔断 + EventBus

**目标：** 增强容错能力。

**前置条件：** Step 6 完成

**产出文件：**

#### 11.1 `dns/circuitbreaker.go`

```go
package dns

import (
    "log/slog"
    "sync"
)

// CircuitBreaker 每个域名独立的熔断器
type CircuitBreaker struct {
    mu            sync.Mutex
    failCount     map[string]int  // 域名 → 连续失败次数
    threshold     int             // 熔断阈值
}

func NewCircuitBreaker(threshold int) *CircuitBreaker {
    return &CircuitBreaker{
        failCount: make(map[string]int),
        threshold: threshold,
    }
}

// IsOpen 判断域名是否已熔断
func (cb *CircuitBreaker) IsOpen(domain string) bool {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    return cb.failCount[domain] >= cb.threshold
}

// RecordSuccess 记录成功，解除熔断
func (cb *CircuitBreaker) RecordSuccess(domain string) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    if cb.failCount[domain] > 0 {
        slog.Info("DNS 熔断解除", "domain", domain)
    }
    cb.failCount[domain] = 0
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure(domain string) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.failCount[domain]++
    count := cb.failCount[domain]
    if count == cb.threshold {
        slog.Error("DNS 熔断触发", "domain", domain, "连续失败", count)
    }
}
```

#### 11.2 `notifier/bus.go`

```go
package notifier

import (
    "log/slog"
    "sync"
    "time"
)

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
    mu          sync.RWMutex
    subscribers map[EventType][]Subscriber
}

func NewEventBus() *EventBus {
    return &EventBus{subscribers: make(map[EventType][]Subscriber)}
}

func (b *EventBus) Subscribe(eventType EventType, sub Subscriber) {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.subscribers[eventType] = append(b.subscribers[eventType], sub)
}

// Publish 异步投递，不阻塞调用方
func (b *EventBus) Publish(event Event) {
    b.mu.RLock()
    subs := b.subscribers[event.Type]
    b.mu.RUnlock()

    for _, sub := range subs {
        go func(s Subscriber) {
            if err := s.OnEvent(event); err != nil {
                slog.Warn("事件处理失败", "type", event.Type, "error", err)
            }
        }(sub)
    }
}
```

**验收：**
```bash
go build ./...
go test ./dns/... ./notifier/... -v
```
测试：熔断触发/解除、EventBus 异步投递。

---

### Step 12：Docker 构建 + Makefile

**目标：** 容器化部署可用。

**前置条件：** Step 7 完成

**产出文件：**

#### 12.1 `build/Dockerfile`

```dockerfile
# 编译阶段
FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -tags docker \
    -ldflags="-s -w -X github.com/alcaprophet/fwalizer/version.Version=${VERSION}" \
    -o /fwalizer .

# 运行阶段
FROM alpine:3.20
RUN adduser -D appuser
COPY --from=builder /fwalizer /usr/local/bin/fwalizer
USER appuser
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD pgrep fwalizer || exit 1
ENTRYPOINT ["fwalizer"]
```

#### 12.2 `Makefile` 更新要点

```makefile
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/alcaprophet/fwalizer/version.Version=$(VERSION)

build:
	go build -ldflags="$(LDFLAGS)" -o fwalizer .

docker-build:
	docker build -f build/Dockerfile --build-arg VERSION=$(VERSION) -t fwalizer .
```

#### 12.3 `.dockerignore`

```
Documents/
*.md
.env
.git/
```

**验收：**
```bash
make build && ./fwalizer version
make docker-build
docker run --rm fwalizer version
```

---

### Step 13：WebUI 后端（Phase 2）

**目标：** WebUI 模式下 API 可用，配置持久化到 SQLite。

**前置条件：** Step 12 完成

**产出文件及要点：**

| 文件 | 要点 |
|------|------|
| `config/store.go` | SQLite 初始化（WAL + busy_timeout=5000）；表：targets/rules/settings/sync_logs |
| `webui/server.go` | `net/http` + `http.ServeMux`；绑定 `127.0.0.1:{port}`；静态文件 serve |
| `webui/api/targets.go` | GET/POST/PUT/DELETE /api/targets；POST /api/test-connection |
| `webui/api/rules.go` | GET/POST/PUT/DELETE /api/rules |
| `webui/api/sync.go` | GET /api/sync/status；POST /api/sync/trigger；POST /api/sync/dryrun |
| `webui/api/settings.go` | GET/PUT /api/settings；GET /api/config/export；POST /api/config/import |
| `app/app.go` 补充 | WebUI 模式启动 HTTP Server + Syncer；pidfile 防多实例 |

**验收：**
```bash
./fwalizer  # 无 TARGETS 环境变量 → 进入 WebUI 模式
curl http://127.0.0.1:9090/api/health  # 返回 {"status":"ok"}
```

---

### Step 14：WebUI 前端（Phase 2）

**目标：** 浏览器可操作全部功能。

**前置条件：** Step 13 完成

**产出文件：** `webui/frontend/` 目录

**技术约束：**
- Vue 3 通过 CDN 引入（`<script src="https://unpkg.com/vue@3">`）
- 无构建工具（无 npm/webpack）
- 单 HTML 文件 + 内联 JS，或简单多文件
- 通过 `//go:embed frontend/dist` 编译进二进制

**页面：** 仪表盘 / 云资源管理 / 域名规则 / 全局设置 / 同步日志

**验收：** 浏览器访问 `http://127.0.0.1:9090` 可操作全部功能。

---

### Step 15：告警 + 高级功能（Phase 3）

**目标：** 同步异常时可通知用户。

**前置条件：** Step 11 + Step 13 完成

**产出文件：**

| 文件 | 要点 |
|------|------|
| `notifier/email.go` | SMTP 发送；配置：SMTP_HOST/PORT/USER/PASS/FROM/TO |
| `notifier/webhook.go` | POST JSON 到指定 URL；支持钉钉/飞书/Slack 格式 |
| `app/cli.go` 补充 | `fwalizer backup` / `fwalizer restore [file]` |

**验收：** 模拟同步失败 → 收到邮件/Webhook 通知。

---

### Step 16：桌面端（Phase 4）

**目标：** 系统托盘常驻 + 开机自启。

**前置条件：** Step 14 完成

**产出文件：** `app/systray.go`（`//go:build desktop`）

**依赖添加：**
```bash
go get fyne.io/systray
```

**实现要点：**
- 托盘菜单：状态指示 / 打开配置面板 / 立即同步 / 开机自启[开关] / 退出
- 开机自启：Windows 写注册表，macOS 生成 plist
- 启动后自动打开浏览器（`open` / `xdg-open` / `rundll32`）

**验收：**
```bash
CGO_ENABLED=1 go build -tags desktop -o fwalizer .
./fwalizer  # 托盘出现，浏览器自动打开
```

---

### 构建顺序依赖图

```
Step 1 (骨架)
  └─ Step 2 (.env)
       └─ Step 3 (DNS)
            └─ Step 4 (Provider 抽象)
                 └─ Step 5 (Lighthouse)
                      └─ Step 6 (Syncer)
                           └─ Step 7 (App + main)  ← MVP
                                ├─ Step 8 (CVM)
                                ├─ Step 9 (SWAS)
                                ├─ Step 10 (ECS)
                                └─ Step 11 (熔断 + EventBus)
                                     └─ Step 12 (Docker)
                                          ├─ Step 13 (WebUI 后端)
                                          │    └─ Step 14 (WebUI 前端)
                                          ├─ Step 15 (告警)
                                          └─ Step 16 (桌面端)
```

**关键约束：**
- Step 1–7 为最小可用版本（MVP），仅支持 Lighthouse + .env 模式
- Step 8–10 可任意顺序，但每个完成后必须全量编译 + `go vet`
- Step 13–16 属于后续 Phase，可根据优先级调整顺序
- 每个 Step 完成后必须：`go build ./... && go vet ./... && go test ./...`

---

# 技术实现细节

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
│   └── resolver.go              # DNS 解析
├── provider/                    # 多云抽象层
│   ├── provider.go              # Provider 接口定义
│   ├── registry.go              # Provider 注册表（工厂模式）
│   ├── common.go                # 通用工具（ownedRules、Diff、端口转换）
│   ├── tc_lighthouse.go         # 腾讯云 Lighthouse 实现
│   ├── tc_cvm.go                # 腾讯云 CVM 安全组实现
│   ├── ali_swas.go              # 阿里云轻量云实现
│   └── ali_ecs.go               # 阿里云 ECS 安全组实现
├── syncer/                      # 同步引擎
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
    Protocol      string // TCP / UDP / TCP+UDP / ICMP / ALL
    Port          string // 归一化为 "port" 或 "start-end"
    CidrBlock     string // IPv4 CIDR
    Ipv6CidrBlock string // IPv6 CIDR
    Action        string // ACCEPT / DROP
    Description   string // 规则描述/备注
    PolicyIndex   string // 规则索引（CVM 安全组删除时需要）
    RuleID        string // 规则唯一 ID（阿里云 SWAS/ECS 删除时需要）
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
    DNSFailThreshold int      // DNS 连续失败熔断阈值，默认 5
    LogLevel    string         // debug / info / warn / error
    WebUIPort   int            // WebUI 端口，默认 9090
    Mode        string         // 运行模式：env / webui（空 = 自动检测）
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
    DeleteRules(rules []RuleInfo) error  // 传入 RuleInfo（含 RuleID/PolicyIndex）
    ConvertPorts(port string) []string   // 统一端口 → 云厂商格式（可能一对多）
    TargetIndex() int
}
```

**ConvertPorts 职责：** 将统一端口格式（如 `80,443,8000-8010`）转换为对应云厂商的端口格式列表：
- 腾讯云 Lighthouse：`["80,443,8000-8010"]`（保持原样，单条规则）
- 腾讯云 CVM：`["80", "443", "8000-8010"]`（不支持逗号分隔，拆分为多条规则）
- 阿里云轻量云：`["80/80", "443/443", "8000/8010"]`（斜杠格式，每条规则一个端口/范围）
- 阿里云 ECS：`["80/80", "443/443", "8000/8010"]`（斜杠格式，每条规则一个端口/范围）
- ICMP 协议：腾讯云用 `ALL`，阿里云用 `-1/-1`（不经过 ConvertPorts，由 Provider 内部处理）

**DeleteRules 说明：** 传入 `[]RuleInfo`（而非 RuleAction），因为删除需要 RuleID（阿里云）或 PolicyIndex（CVM）等查询时才有的字段。

### 3.2 工厂注册

```go
type Factory func(cfg TargetConfig, pool *ClientPool) (Provider, error)

var registry = map[CloudType]Factory{}

func Register(cloudType CloudType, factory Factory) {
    registry[cloudType] = factory
}

func NewProvider(cfg TargetConfig, pool *ClientPool) (Provider, error) {
    factory, ok := registry[cfg.CloudType]
    if !ok {
        return nil, fmt.Errorf("不支持的云产品类型: %s", cfg.CloudType)
    }
    return factory(cfg, pool)
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
    portConverter func(string) []string,
) (toAdd []RuleAction, toDelete []RuleInfo) {
    // 通用 diff 逻辑，与云厂商无关
    // toDelete 返回 RuleInfo（保留 RuleID/PolicyIndex 供删除使用）
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
        return 800 * time.Millisecond  // 100次/60秒，留余量
    case CloudTCLighthouse:
        return 200 * time.Millisecond  // 10次/秒，留余量
    default:
        return 100 * time.Millisecond  // CVM 50次/秒、阿里云 ECS 无限制
    }
}
```

### 4.4 乐观锁重试

- 写入前重新 Describe → 重新 diff → 重新 Create/Delete
- 最多 3 次，指数退避
- 不传入 FirewallVersion / Version 参数（由云 API 自行管理版本号）

### 4.5 错误处理策略

**“规则已不存在”视为成功：**

删除时如果云 API 返回“规则不存在”类错误，说明规则已被其他途径删除，应视为成功（幂等）：

| 云厂商 | 错误码 | 处理 |
|--------|--------|------|
| 腾讯云 Lighthouse | `ResourceNotFound.FirewallRulesNotFound` | 视为成功，跳过 |
| 腾讯云 CVM | 删除不匹配的规则时静默成功 | 无需处理 |
| 阿里云 SWAS | `InvalidInstanceId.NotFound` / 规则 ID 无效 | 视为成功，跳过 |
| 阿里云 ECS | `InvalidSecurityGroupRuleId.NotFound` / `InvalidSecurityGroupRule.RuleNotExist` | 视为成功，跳过 |

**“规则已存在”视为成功：**

添加时如果云 API 返回“规则已存在”，说明无需重复添加：

| 云厂商 | 错误码 | 处理 |
|--------|--------|------|
| 腾讯云 Lighthouse | `InvalidParameter.FirewallRulesExist` | WARN 日志，跳过 |
| 阿里云 SWAS | `FirewallRuleAlreadyExist` | WARN 日志，跳过 |
| 阿里云 ECS | 调用成功但不重复添加 | 无需处理 |

**真正错误（需重试）：**
- 网络超时、连接失败
- 频率限制（`RequestLimitExceeded`）
- 服务端内部错误（`InternalError`）
- 防火墙忙（`UnsupportedOperation.FirewallBusy`）

### 4.6 规则数量限制

- 腾讯云 CVM 安全组：单安全组规则上限 **100 条**（入站 + 出站合计）
- 其他云厂商：暂无硬性限制（Lighthouse 无明确上限，阿里云较宽松）
- CreateRules 前应检查当前规则数，接近上限时记录 ERROR 日志并跳过新增

---

## 五、各云厂商 API 对比

| 操作 | 腾讯云 Lighthouse | 腾讯云 CVM | 阿里云轻量云 | 阿里云 ECS |
|------|-------------------|-----------|-------------|-----------|
| **查询** | `DescribeFirewallRules` | `DescribeSecurityGroupPolicies` | `ListFirewallRules` | `DescribeSecurityGroupAttribute` |
| **添加** | `CreateFirewallRules` | `CreateSecurityGroupPolicies` | `CreateFirewallRules` | `AuthorizeSecurityGroup` |
| **删除** | `DeleteFirewallRules` | `DeleteSecurityGroupPolicies` | `DeleteFirewallRules` | `RevokeSecurityGroup` |
| **Endpoint** | `lighthouse.tencentcloudapi.com` | `vpc.tencentcloudapi.com` | `swas.{region}.aliyuncs.com` | `ecs.{region}.aliyuncs.com` |
| **频率限制** | 10次/秒 | 查询/删除100次/秒，创建50次/秒 | 100次/60秒 | 不限 |
| **规则标识字段** | `FirewallRuleDescription` | `PolicyDescription` | `Remark` | `Description` |
| **端口格式** | `80` 或 `443,80` 或 `ALL` | `80` 或 `8000-8010`（不支持逗号分隔） | `80/80` 或 `1/200`，ICMP 用 `-1/-1` | `80/80` 或 `1/200`，ICMP 用 `-1/-1` |
| **IPv6 字段** | `Ipv6CidrBlock` | `Ipv6CidrBlock` | `SourceCidrIp`（仅 IPv4，不支持 IPv6） | `Ipv6SourceCidrIp` |
| **操作对象** | `InstanceId` | `SecurityGroupId` | `InstanceId` | `SecurityGroupId` |

---

## 六、各 Provider 实现要点

### 6.1 腾讯云 Lighthouse

- 规则标识：`FirewallRuleDescription` 以 `[TAG]` 开头（≤ 64 字符）
- IPv6：使用 `Ipv6CidrBlock` 字段（与 `CidrBlock` 互斥，分两条规则）
- IPv6 + ICMP 时协议需用 `ICMPv6`（API 支持 TCP/UDP/ICMP/ICMPv6/ALL）
- **绝不**使用 `ModifyFirewallRules`

### 6.2 腾讯云 CVM 安全组

- SDK: `tencentcloud-sdk-go/tencentcloud/vpc/v20170312`
- 操作对象是 `SecurityGroupId`（非 InstanceId）
- 端口格式：仅支持单端口（`80`）或范围（`8000-8010`），**不支持逗号分隔**，需在 ConvertPorts 中拆分为多条规则
- 本工具只操作 **Ingress**（入站）规则，不触碰 Egress
- 删除支持两种方式：指定 `PolicyIndex` 或规则匹配（Action + Protocol + CidrBlock + Port）
- 规则描述：`PolicyDescription`
- 一次请求只能创建/删除单个方向的规则（Ingress 或 Egress）
- 创建频率限制 50次/秒，查询/删除 100次/秒

### 6.3 阿里云轻量云 (SWAS-OPEN)

- SDK: `github.com/alibabacloud-go/swas-open-20200601/v3`（V2 SDK）
- 端口格式：`"80/80"` 或 `"1/200"`（斜杠分隔），ICMP 用 `"-1/-1"`
- 规则标识：`Remark`
- 协议字段名：`RuleProtocol`（取值：TCP / UDP / TCP+UDP / ICMP）
- 删除接口需要 `RuleIds`（规则 ID 列表），需先通过 ListFirewallRules 获取 RuleId
- 频率限制严格：100次/60秒，建议 800ms 间隔
- 仅支持 IPv4（`SourceCidrIp`），不支持 IPv6

### 6.4 阿里云 ECS 安全组

- SDK: `github.com/alibabacloud-go/ecs-20140526/v7`（V2 SDK）
- 操作对象是 `SecurityGroupId`
- 端口格式必须用斜杠：`"80/80"` 或 `"1/200"`，ICMP 用 `"-1/-1"`
- IPv6 用独立字段 `Ipv6SourceCidrIp`（与 `SourceCidrIp` 不可同时设置）
- 规则标识：`Description`（1~512 个字符）
- 规则已存在时调用 AuthorizeSecurityGroup 成功但不重复添加
- 删除支持两种方式：指定 `SecurityGroupRuleId`（推荐）或 Permissions 匹配
- 本工具只操作 **Ingress**（入方向）规则，使用 `AuthorizeSecurityGroup` / `RevokeSecurityGroup`
- 优先级 `Priority`：默认 1，范围 1~100

---

## 七、.env 配置格式

```env
# ═══ 云资源目标（provider|resource_id|region） ═══
TARGETS=tc_lighthouse|lhins-abc|ap-guangzhou, \
        tc_cvm|sg-def|ap-shanghai, \
        ali_swas|ace0706b|cn-hangzhou, \
        ali_ecs|sg-ghi|cn-shenzhen

# ═══ 云厂商凭据（按厂商分离） ═══
TC_ACCESS_ID=
TC_ACCESS_KEY=
ALI_ACCESS_ID=
ALI_ACCESS_KEY=

# ═══ 域名规则（host|protocol|ports|action|targets|comment） ═══
RULES=api.example.com|TCP|443,80|ACCEPT||生产API, \
      vpn.example.com|UDP|1194|ACCEPT|2|VPN接入, \
      game.example.com|TCP|8000-8010|ACCEPT|1,3|游戏端口, \
      ping.example.com|ICMP|ALL|ACCEPT||允许Ping

# ═══ 全局设置 ═══
TAG=auto-dns
INTERVAL=5m
DNS=8.8.8.8:53
LOG_LEVEL=info

# ═══ 可选配置（有合理默认值，不填则自动生效） ═══
# WEBUI_PORT=9090            # WebUI 端口（默认 9090）
# DNS_TIMEOUT=10s            # DNS 解析超时（默认 10s）
# DNS_FAIL_THRESHOLD=5       # DNS 连续失败熔断阈值（默认 5）
# FWALIZER_MODE=env|webui    # 强制运行模式（默认自动检测）
```

### 字段说明

**TARGETS 格式：** `provider|resource_id|region`
- provider：`tc_lighthouse` / `tc_cvm` / `ali_swas` / `ali_ecs`
- resource_id：InstanceId 或 SecurityGroupId
- region：云厂商地域 ID

**RULES 格式：** `host|protocol|ports|action|targets|comment`
- protocol：`TCP` / `UDP` / `TCP+UDP` / `ICMP`
- ports：单端口、逗号分隔、范围（`8000-8010`）、`ALL`；ICMP 时固定为 `ALL`
- action：`ACCEPT` / `DROP`
- targets：空或 `*` = 所有 Target；`1,3` = 指定编号（从 1 开始）
- comment：可选备注，会拼接到规则描述中

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
CGO_ENABLED=0 go build -tags docker -ldflags="-s -w -X github.com/alcaprophet/fwalizer/version.Version=$VERSION" .

# 桌面构建（含系统托盘，需要 CGO）
CGO_ENABLED=1 go build -tags desktop -ldflags="-s -w -X github.com/alcaprophet/fwalizer/version.Version=$VERSION" .
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

    // 阿里云 V2 SDK
    github.com/alibabacloud-go/swas-open-20200601/v3 v3.x.x                  // 新增（轻量云）
    github.com/alibabacloud-go/ecs-20140526/v7 v7.x.x                        // 新增（ECS）
    github.com/alibabacloud-go/darabonba-openapi/v2 v2.x.x                   // 新增（SDK 核心库）
    github.com/aliyun/credentials-go v1.x.x                                  // 新增（凭据管理）

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
- `.env` 模式：`pgrep fwalizer` 进程检测（Alpine 无 `killall -0`）

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
