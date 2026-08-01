<script setup lang="ts">
// 运行测试页：Dry Run + 连接测试双标签（激活状态与路由 query 同步 ?tab=dryrun|connection）
import { NTabs, NTabPane, NButton, NSelect, NInput, NSpace, useMessage } from 'naive-ui'
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import DryRunResults from '../components/DryRunResults.vue'
import { useDryRun } from '../composables/useDryRun'
import { request } from '../api'
import { cloudOptions } from '../constants'
import type { TestConnectionResult } from '../types'

const message = useMessage()
const route = useRoute()
const router = useRouter()

// ─── Tab 1：Dry Run ───
const { loading, results, warnings, lastRunAt, run } = useDryRun()

async function runDryRun() {
  try {
    await run()
    message.success('Dry Run 完成')
  } catch (e: any) {
    message.error(`Dry Run 失败: ${e.message}`)
  }
}

// ─── Tab 2：连接测试（表单内嵌，不抽 composable） ───
const testForm = ref({ cloud_type: 'tc_lighthouse', region: '', resource_id: '' })
const testResult = ref('')
const testLoading = ref(false)

async function testConnection() {
  testLoading.value = true
  testResult.value = '测试中...'
  try {
    const data = await request<TestConnectionResult>(
      '/api/test-connection',
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(testForm.value),
      },
      15000 // 15s 超时：云 API 卡住时及时提示
    )
    testResult.value = data.success ? (data.message || '连接成功') : `失败: ${data.error || '未知错误'}`
  } catch (e: any) {
    testResult.value = e?.message === '请求超时'
      ? '连接超时（15 秒），请检查网络或云 API 状态'
      : `请求失败: ${e.message}`
  } finally {
    testLoading.value = false
  }
}

// ─── Tab 与路由 query 同步 ───
const activeTab = ref(String(route.query.tab || 'dryrun'))
watch(activeTab, (v) => {
  router.replace({ query: { ...route.query, tab: v } })
})
</script>

<template>
  <div>
    <h2>运行测试</h2>
    <NTabs v-model:value="activeTab" type="line">
      <NTabPane name="dryrun" tab="Dry Run">
        <NSpace vertical>
          <NSpace align="center">
            <NButton type="primary" :loading="loading" @click="runDryRun">执行 Dry Run</NButton>
            <span v-if="lastRunAt" style="font-size: 12px; color: #999">
              上次执行：{{ lastRunAt.toLocaleTimeString() }}
            </span>
          </NSpace>
          <DryRunResults :results="results" :warnings="warnings" :has-run="lastRunAt !== null" />
        </NSpace>
      </NTabPane>

      <NTabPane name="connection" tab="连接测试">
        <NSpace vertical style="max-width: 400px">
          <NSelect v-model:value="testForm.cloud_type" :options="cloudOptions" placeholder="选择云产品" />
          <NInput v-model:value="testForm.resource_id" placeholder="资源ID（lhins-xxx / sg-xxx）" />
          <NInput v-model:value="testForm.region" placeholder="地域（ap-guangzhou）" />
          <NButton type="primary" :loading="testLoading" @click="testConnection">测试连接</NButton>
          <p v-if="testResult" style="color: #666">{{ testResult }}</p>
        </NSpace>
      </NTabPane>
    </NTabs>
  </div>
</template>
