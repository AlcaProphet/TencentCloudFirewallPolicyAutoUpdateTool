// API 响应类型定义

export interface TargetConfig {
  id: number
  cloud_type: string
  region: string
  resource_id: string
}

// ZoneRegion 地域及其可用区（GET /api/zones，地域自动补全数据源）
export interface ZoneRegion {
  id: string
  name: string
  zones: string[]
}

// ScannedResource 扫描到的云资源（GET /api/scanned-resources，资源 ID 自动补全数据源）
export interface ScannedResource {
  id: number
  cloud_type: string
  region: string
  resource_id: string
  resource_name: string
}

export interface DomainRule {
  id: number
  host: string
  protocol: string
  ports: string
  action: string
  targets: number[]
  comment: string
  enable_ipv6: boolean
}

export interface SyncStatus {
  running: boolean
  last_sync: string | null
  enabled: boolean // 同步开关（Step 11 起后端必返回）
}

// RuleChange 规则变更摘要（Dry Run 明细化）
export interface RuleChange {
  protocol: string
  port: string
  action: string
  cidr: string // IPv4 或 IPv6 的 CIDR
  desc: string // 规则描述（含 [TAG]）
}

// DryRunResult 试运行结果（to_add/to_delete 为规则明细数组）
export interface DryRunResult {
  provider: string
  domain: string
  to_add: RuleChange[]
  to_delete: RuleChange[]
  error?: string
}

// DryRunResponse Dry Run 响应包装（空状态语义化）
export interface DryRunResponse {
  results: DryRunResult[]
  warnings: string[]
}

// TestConnectionResult 连接测试结果
export interface TestConnectionResult {
  success: boolean
  message?: string
  error?: string
}

export interface SyncLogEntry {
  timestamp: string
  target: string
  domain: string
  result: string
  added: number
  deleted: number
  error?: string // 失败详情（后端已返回，Build4 Step 4 前端消费）
}

export interface SyncEvent {
  type: string
  timestamp: string
  data: Record<string, unknown>
}

export interface AlertEmailConfig {
  enabled: boolean
  host: string
  port: string
  username: string
  password: string
  from_addr: string
  to_addr: string
}

export interface AlertWebhookConfig {
  enabled: boolean
  url: string
  channel?: string
}
