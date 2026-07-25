<script setup lang="ts">
import { NCard, NStatistic, NGrid, NGi, NButton, NTag, NSpace, NModal, NDataTable } from 'naive-ui'
import { ref, onMounted, onUnmounted } from 'vue'

const status = ref<any>({ running: false })
const dryrunResults = ref<any[]>([])
const showDryrun = ref(false)
let timer: ReturnType<typeof setInterval> | null = null

async function fetchStatus() {
  try {
    const res = await fetch('/api/sync/status')
    status.value = await res.json()
  } catch { /* ignore */ }
}

async function triggerSync() {
  await fetch('/api/sync/trigger', { method: 'POST' })
  setTimeout(fetchStatus, 1000)
}

async function dryRun() {
  const res = await fetch('/api/sync/dryrun', { method: 'POST' })
  dryrunResults.value = await res.json() || []
  showDryrun.value = true
}

const dryrunColumns = [
  { title: 'Provider', key: 'provider' },
  { title: '域名', key: 'domain' },
  { title: '待添加', key: 'to_add' },
  { title: '待删除', key: 'to_delete' },
  { title: '错误', key: 'error' },
]

onMounted(() => {
  fetchStatus()
  timer = setInterval(fetchStatus, 5000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div>
    <h2>仪表盘</h2>
    <NGrid :cols="3" :x-gap="16">
      <NGi>
        <NCard>
          <NStatistic label="同步引擎">
            <NTag :type="status.running ? 'success' : 'default'" size="small">
              {{ status.running ? '运行中' : '已停止' }}
            </NTag>
          </NStatistic>
        </NCard>
      </NGi>
      <NGi>
        <NCard>
          <NStatistic label="上次同步" :value="status.last_sync ? new Date(status.last_sync).toLocaleString() : '无'" />
        </NCard>
      </NGi>
      <NGi>
        <NCard>
          <NSpace>
            <NButton type="primary" size="small" @click="triggerSync">立即同步</NButton>
            <NButton size="small" @click="dryRun">试运行</NButton>
          </NSpace>
        </NCard>
      </NGi>
    </NGrid>

    <NModal v-model:show="showDryrun" title="试运行结果" preset="card" style="width: 700px">
      <NDataTable :columns="dryrunColumns" :data="dryrunResults" :bordered="true" size="small" />
    </NModal>
  </div>
</template>
