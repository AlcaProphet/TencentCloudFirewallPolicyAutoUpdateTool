<script setup lang="ts">
import { NDataTable, NTag, NCollapse, NCollapseItem } from 'naive-ui'
import { ref, onMounted, onUnmounted, h } from 'vue'

const logs = ref<any[]>([])
const events = ref<any[]>([])
const logLines = ref<string[]>([])
let es: EventSource | null = null
let logEs: EventSource | null = null

// ─── 时间格式化（UTC） ───
function formatTime(ts: string): string {
  if (!ts) return '-'
  const d = new Date(ts)
  if (isNaN(d.getTime())) return ts
  return d.toISOString().replace('T', ' ').substring(0, 19) + ' UTC'
}

// ─── 事件类型映射 ───
const eventTypeLabels: Record<string, string> = {
  'sync:start': '同步开始',
  'sync:complete': '同步完成',
  'sync:error': '同步失败',
  'dns:failed': 'DNS解析失败',
  'rule:changed': '规则变更',
}

function eventTagType(type: string): string {
  if (type === 'sync:error' || type === 'dns:failed') return 'error'
  if (type === 'sync:complete') return 'success'
  return 'info'
}

function formatEventData(row: any): string {
  const d = row.data || {}
  switch (row.type) {
    case 'sync:start':
      return `${d.targets ?? 0} 个目标，${d.rules ?? 0} 条规则`
    case 'sync:complete':
      return d.domain ? `${d.provider} / ${d.domain}` : `耗时 ${d.duration ?? '-'}`
    case 'sync:error':
      return `${d.provider ?? ''} / ${d.domain ?? ''}：${d.error ?? '未知错误'}`
    case 'dns:failed':
      return `${d.domain ?? ''}：${d.error ?? '解析超时'}`
    default:
      return JSON.stringify(d)
  }
}

// ─── 生命周期 ───
onMounted(async () => {
  const res = await fetch('/api/sync/logs')
  logs.value = await res.json() || []

  // SSE 实时事件推送
  es = new EventSource('/api/sync/events')
  es.onmessage = (e) => {
    try {
      const ev = JSON.parse(e.data)
      events.value.unshift(ev)
      if (events.value.length > 50) events.value.pop()
    } catch { /* ignore */ }
  }

  // SSE 实时日志流
  logEs = new EventSource('/api/logs/stream')
  logEs.onmessage = (e) => {
    logLines.value.push(e.data)
    if (logLines.value.length > 200) logLines.value.shift()
  }
})

onUnmounted(() => {
  if (es) es.close()
  if (logEs) logEs.close()
})

// ─── 历史记录列 ───
const columns = [
  { title: '时间 (UTC)', key: 'timestamp', render: (row: any) => formatTime(row.timestamp) },
  { title: '目标', key: 'target' },
  { title: '域名', key: 'domain' },
  {
    title: '结果',
    key: 'result',
    render(row: any) {
      const type = row.result === 'success' ? 'success' : row.result === 'failed' ? 'error' : 'warning'
      return h(NTag, { type, size: 'small' }, { default: () => row.result })
    }
  },
  { title: '新增', key: 'added' },
  { title: '删除', key: 'deleted' },
]

// ─── 实时事件列 ───
const eventColumns = [
  { title: '时间 (UTC)', key: 'timestamp', render: (row: any) => formatTime(row.timestamp) },
  {
    title: '事件', key: 'type',
    render(row: any) {
      return h(NTag, { size: 'small', type: eventTagType(row.type) as any }, { default: () => eventTypeLabels[row.type] || row.type })
    }
  },
  { title: '详情', key: 'data', render: (row: any) => formatEventData(row) },
]
</script>

<template>
  <div>
    <h2>同步日志</h2>

    <h3>实时事件</h3>
    <NDataTable :columns="eventColumns" :data="events" :bordered="true" :max-height="200" size="small" />

    <h3 style="margin-top: 16px">历史记录</h3>
    <NDataTable :columns="columns" :data="logs" :bordered="true" :max-height="400" />

    <NCollapse style="margin-top: 16px">
      <NCollapseItem title="运行日志（实时）" name="logs">
        <pre style="max-height: 300px; overflow-y: auto; background: #1e1e1e; color: #d4d4d4; padding: 12px; border-radius: 6px; font-size: 12px; line-height: 1.6; white-space: pre-wrap; word-break: break-all;">{{ logLines.join('\n') || '等待日志输出...' }}</pre>
      </NCollapseItem>
    </NCollapse>
  </div>
</template>
