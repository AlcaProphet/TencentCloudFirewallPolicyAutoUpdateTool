<script setup lang="ts">
import { NTabs, NTabPane, NButton, NDataTable, NSelect, NInput, NSpace, NUpload, useMessage } from 'naive-ui'
import { ref } from 'vue'

const message = useMessage()

// Dry Run
const dryRunLoading = ref(false)
const dryRunResults = ref<any[]>([])
const dryRunColumns = [
  { title: 'Provider', key: 'provider' },
  { title: '域名', key: 'domain' },
  { title: '待添加', key: 'to_add' },
  { title: '待删除', key: 'to_delete' },
  { title: '错误', key: 'error' },
]

async function runDryRun() {
  dryRunLoading.value = true
  dryRunResults.value = []
  try {
    const res = await fetch('/api/sync/dryrun', { method: 'POST' })
    const data = await res.json()
    dryRunResults.value = data || []
    message.success('Dry Run 完成')
  } catch (e: any) {
    message.error(`Dry Run 失败: ${e.message}`)
  } finally {
    dryRunLoading.value = false
  }
}

// 配置导入/导出
async function exportConfig() {
  window.open('/api/config/export', '_blank')
}

function handleImport({ file }: any) {
  const reader = new FileReader()
  reader.onload = async () => {
    try {
      const res = await fetch('/api/config/import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: reader.result as string,
      })
      const data = await res.json()
      if (res.ok) {
        message.success('导入成功')
      } else {
        message.error(`导入失败: ${data.error}`)
      }
    } catch (e: any) {
      message.error(`导入失败: ${e.message}`)
    }
  }
  reader.readAsText(file.file)
  return false
}

// 连接测试
const testForm = ref({ cloud_type: 'tc_lighthouse', region: '', resource_id: '' })
const testResult = ref('')
const testLoading = ref(false)
const cloudOptions = [
  { label: '腾讯云轻量云', value: 'tc_lighthouse' },
  { label: '腾讯云CVM', value: 'tc_cvm' },
  { label: '阿里云轻量云', value: 'ali_swas' },
  { label: '阿里云ECS', value: 'ali_ecs' },
]

async function testConnection() {
  testLoading.value = true
  testResult.value = '测试中...'
  try {
    const res = await fetch('/api/test-connection', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(testForm.value),
    })
    const data = await res.json()
    testResult.value = data.success ? data.message : `失败: ${data.error}`
  } catch (e: any) {
    testResult.value = `请求失败: ${e.message}`
  } finally {
    testLoading.value = false
  }
}
</script>

<template>
  <div>
    <h2>高级功能</h2>
    <NTabs type="line">
      <NTabPane name="dryrun" tab="Dry Run">
        <NSpace vertical>
          <NButton type="primary" :loading="dryRunLoading" @click="runDryRun">执行 Dry Run</NButton>
          <NDataTable :columns="dryRunColumns" :data="dryRunResults" :bordered="true" />
        </NSpace>
      </NTabPane>

      <NTabPane name="config" tab="配置导入/导出">
        <NSpace vertical>
          <NButton @click="exportConfig">导出配置（JSON）</NButton>
          <NUpload :show-file-list="false" accept=".json" :on-change="handleImport">
            <NButton>导入配置</NButton>
          </NUpload>
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
