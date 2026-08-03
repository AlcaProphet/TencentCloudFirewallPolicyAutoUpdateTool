<script setup lang="ts">
// 运行测试页：Dry Run + 连接测试双标签（激活状态与路由 query 同步 ?tab=dryrun|connection）
import { NTabs, NTabPane, NButton, NSelect, NInput, NSpace, useMessage } from 'naive-ui'
import { ref, watch, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import DryRunResults from '../components/DryRunResults.vue'
import { useDryRun } from '../composables/useDryRun'
import { useZones } from '../composables/useZones'
import { useScannedResources } from '../composables/useScannedResources'
import { request } from '../api'
import { cloudOptions, resourceIdHint } from '../constants'
import type { TestConnectionResult } from '../types'

const message = useMessage()
const route = useRoute()
const router = useRouter()

// ─── Tab 1：模拟测试（Build4 Step 12：Dry Run 更名） ───
const { loading, results, warnings, lastRunAt, run } = useDryRun()

async function runDryRun() {
  try {
    await run()
    message.success('模拟测试完成')
  } catch (e: any) {
    message.error(`模拟测试失败: ${e.message}`)
  }
}

// ─── Tab 2：连接测试（表单内嵌，不抽 composable） ───
// resource_id 初始为 null：避免 NSelect tag 模式将空字符串视为已选值导致 placeholder 不显示
const testForm = ref<{ cloud_type: string; region: string; resource_id: string | null }>({ cloud_type: 'tc_lighthouse', region: '', resource_id: null })
const testResult = ref('')
const testLoading = ref(false)

// 地域自动补全：预填建议 + 允许输入任意值（与 Targets.vue 一致）
const { load: loadZones, regionOptions } = useZones()
const regionOpts = computed(() => {
  const opts = regionOptions(testForm.value.cloud_type)
  const cur = testForm.value.region
  if (cur && !opts.some((o) => o.value === cur)) {
    opts.push({ label: cur, value: cur })
  }
  return opts
})

// 资源 ID 自动补全：扫描结果 + 允许输入任意值（与 Targets.vue 一致）
const { load: loadScanned, resourceOptions, resourcesOf } = useScannedResources()
const resourceOpts = computed(() => {
  const opts = resourceOptions(testForm.value.cloud_type)
  const cur = testForm.value.resource_id
  if (cur && !opts.some((o) => o.value === cur)) {
    opts.push({ label: cur, value: cur })
  }
  return opts
})

// 资源-地域联动：选择扫描出的资源时自动填入其所在区域（与 Targets.vue 一致）
watch(() => testForm.value.resource_id, (rid) => {
  if (!rid) return
  const found = resourcesOf(testForm.value.cloud_type).find((r) => r.resource_id === rid)
  if (found && found.region) {
    testForm.value.region = found.region
  }
})
watch(() => testForm.value.cloud_type, (ct) => { loadScanned(ct) })
onMounted(() => {
  loadZones()
  loadScanned(testForm.value.cloud_type)
})

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
      <NTabPane name="dryrun" tab="模拟测试">
        <NSpace vertical>
          <!-- 说明文字（Build4 Step 12：模拟测试语义） -->
          <div style="font-size: 12px; color: #888; margin-bottom: 4px">模拟测试仅生成变更预览，不实际写入云防火墙规则</div>
          <NSpace align="center">
            <NButton type="primary" :loading="loading" @click="runDryRun">执行模拟测试</NButton>
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
          <NSelect
            v-model:value="testForm.resource_id"
            :options="resourceOpts"
            :placeholder="resourceIdHint(testForm.cloud_type)"
            filterable
            tag
            clearable
          />
          <NSelect
            v-model:value="testForm.region"
            :options="regionOpts"
            filterable
            tag
            clearable
            placeholder="选择或输入地域 ID（如 ap-guangzhou）"
          />
          <NButton type="primary" :loading="testLoading" @click="testConnection">测试连接</NButton>
          <p v-if="testResult" style="color: #666">{{ testResult }}</p>
        </NSpace>
      </NTabPane>
    </NTabs>
  </div>
</template>
