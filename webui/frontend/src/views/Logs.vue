<script setup lang="ts">
import { NDataTable, NTag } from 'naive-ui'
import { ref, onMounted, onUnmounted, h } from 'vue'

const logs = ref<any[]>([])
const events = ref<any[]>([])
let es: EventSource | null = null

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
})

onUnmounted(() => {
  if (es) es.close()
})

const columns = [
  { title: '时间', key: 'timestamp' },
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

const eventColumns = [
  { title: '时间', key: 'timestamp' },
  { title: '类型', key: 'type' },
  { title: '详情', key: 'data', render: (row: any) => JSON.stringify(row.data) },
]
</script>

<template>
  <div>
    <h2>同步日志</h2>

    <h3>实时事件</h3>
    <NDataTable :columns="eventColumns" :data="events" :bordered="true" :max-height="200" size="small" />

    <h3 style="margin-top: 16px">历史记录</h3>
    <NDataTable :columns="columns" :data="logs" :bordered="true" :max-height="400" />
  </div>
</template>
