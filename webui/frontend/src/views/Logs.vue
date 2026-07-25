<script setup lang="ts">
import { NDataTable, NTag } from 'naive-ui'
import { ref, onMounted, h } from 'vue'

const logs = ref<any[]>([])

onMounted(async () => {
  const res = await fetch('/api/sync/logs')
  logs.value = await res.json() || []
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
</script>

<template>
  <div>
    <h2>同步日志</h2>
    <NDataTable :columns="columns" :data="logs" :bordered="true" :max-height="500" />
  </div>
</template>
