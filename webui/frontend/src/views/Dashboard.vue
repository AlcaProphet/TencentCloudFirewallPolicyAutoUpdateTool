<script setup lang="ts">
// 仪表盘：同步引擎状态（三态标签）+ 操作（暂停/开启开关、立即同步、运行测试入口）
// enabled=false 时「立即同步」置灰；暂停/开启走 POST /api/sync/pause|resume（先写 DB 后通知 Syncer）
import { NCard, NStatistic, NGrid, NGi, NButton, NTag, NSpace, NTooltip, useMessage } from 'naive-ui'
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { request } from '../api'
import type { SyncStatus } from '../types'

const status = ref<SyncStatus>({ running: false, last_sync: null, enabled: true })
const switching = ref(false) // 开关请求 loading
const router = useRouter()
const message = useMessage()
let timer: ReturnType<typeof setInterval> | null = null

// 三态标签：enabled=true → 运行中（绿）；enabled=false → 已暂停（橙，引擎仍存活于暂停子循环）；running=false → 已停止
const engineTag = computed(() => {
  if (!status.value.running) return { type: 'default' as const, text: '已停止' }
  return status.value.enabled
    ? { type: 'success' as const, text: '运行中' }
    : { type: 'warning' as const, text: '已暂停' }
})

async function fetchStatus() {
  try {
    const s = await request<SyncStatus>('/api/sync/status')
    status.value = s
  } catch { /* 轮询失败忽略 */ }
}

async function triggerSync() {
  try {
    await request('/api/sync/trigger', { method: 'POST' })
    message.success('同步已触发')
  } catch (e: any) {
    message.error(`触发失败: ${e.message}`) // 暂停时后端返回 409
  }
  setTimeout(fetchStatus, 1000)
}

async function toggleSync() {
  switching.value = true
  try {
    await request(`/api/sync/${status.value.enabled ? 'pause' : 'resume'}`, { method: 'POST' })
    status.value.enabled = !status.value.enabled
    message.success(status.value.enabled ? '同步已开启' : '同步已暂停')
  } catch (e: any) {
    message.error(`操作失败: ${e.message}`)
  } finally {
    switching.value = false
    fetchStatus()
  }
}

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
            <NTag :type="engineTag.type" size="small">
              {{ engineTag.text }}
            </NTag>
          </NStatistic>
          <!-- 「运行测试」入口（Dry Run 与连接测试均不受暂停影响） -->
          <NButton text type="primary" size="small" style="margin-top: 8px" @click="router.push('/run-test')">
            运行测试 →
          </NButton>
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
            <!-- 暂停时「立即同步」置灰 + hover 提示 -->
            <NTooltip v-if="!status.enabled">
              <template #trigger>
                <span>
                  <NButton type="primary" size="small" disabled>立即同步</NButton>
                </span>
              </template>
              请先开启同步引擎
            </NTooltip>
            <NButton v-else type="primary" size="small" @click="triggerSync">立即同步</NButton>
            <!-- 暂停/开启开关 -->
            <NButton :type="status.enabled ? 'warning' : 'success'" size="small" :loading="switching" @click="toggleSync">
              {{ status.enabled ? '暂停' : '开启' }}
            </NButton>
          </NSpace>
        </NCard>
      </NGi>
    </NGrid>
  </div>
</template>
