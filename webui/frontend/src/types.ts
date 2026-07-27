// API 响应类型定义

export interface TargetConfig {
  id: number
  cloud_type: string
  region: string
  resource_id: string
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
}

export interface DryRunResult {
  provider: string
  domain: string
  to_add: number
  to_delete: number
  error: string
}

export interface SyncLogEntry {
  timestamp: string
  target: string
  domain: string
  result: string
  added: number
  deleted: number
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
