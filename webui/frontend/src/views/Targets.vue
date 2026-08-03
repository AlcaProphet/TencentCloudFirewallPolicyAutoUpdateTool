<script setup lang="ts">
import { NDataTable, NButton, NModal, NForm, NFormItem, NInput, NSelect, NSpace, NAlert, useMessage } from 'naive-ui'
import { ref, onMounted, h, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { request } from '../api'
import { cloudOptions, cloudLabelMap, resourceIdHint } from '../constants'
import { useSettings } from '../composables/useSettings'
import { useZones } from '../composables/useZones'
import { useScannedResources } from '../composables/useScannedResources'
import type { TargetConfig, TestConnectionResult } from '../types'

const targets = ref<TargetConfig[]>([])
const showModal = ref(false)
const editingId = ref<number | null>(null)
// resource_id 初始为 null：避免 NSelect tag 模式将空字符串视为已选值导致 placeholder 不显示
const form = ref<{ cloud_type: string; region: string; resource_id: string | null }>({ cloud_type: 'tc_lighthouse', region: '', resource_id: null })
const testResult = ref('')
const message = useMessage()

// ─── Keys 缺失提示（改进 12：仅提示，不阻止保存） ───
const router = useRouter()
const { refresh: refreshSettings, tcReady, aliReady } = useSettings()
const credWarning = ref('')

// 依据当前表单云类型刷新凭据提示
function updateCredWarning() {
  const ct = form.value.cloud_type
  if (ct.startsWith('tc_')) {
    credWarning.value = tcReady.value ? '' : '腾讯云凭据未配置，请先在「全局设置」中填写 SecretId/SecretKey，否则同步将失败'
  } else if (ct.startsWith('ali_')) {
    credWarning.value = aliReady.value ? '' : '阿里云凭据未配置，请先在「全局设置」中填写 AccessKeyId/AccessKeySecret，否则同步将失败'
  } else {
    credWarning.value = ''
  }
}

// 云类型变化时刷新提示
watch(() => form.value.cloud_type, updateCredWarning)
watch(() => form.value.cloud_type, (ct) => { loadScanned(ct) })

// ─── 地域自动补全（改进：GET /api/zones 预填建议，允许输入任意值） ───
const { load: loadZones, regionOptions } = useZones()

// ─── 资源 ID 自动补全（扫描结果 + 手动输入） ───
const { load: loadScanned, resourceOptions, resourcesOf } = useScannedResources()

// 资源-地域联动：选择扫描出的资源时自动填入其所在区域（手动输入未匹配不联动）
watch(() => form.value.resource_id, (rid) => {
  if (!rid) return
  const found = resourcesOf(form.value.cloud_type).find((r) => r.resource_id === rid)
  if (found && found.region) {
    form.value.region = found.region
  }
})

// 地域选项：当前云类型预填列表 + 当前已填值（列表外值时保证编辑回显）
const regionOpts = computed(() => {
  const opts = regionOptions(form.value.cloud_type)
  const cur = form.value.region
  if (cur && !opts.some((o) => o.value === cur)) {
    opts.push({ label: cur, value: cur })
  }
  return opts
})

// 资源 ID 选项：扫描结果 + 当前已填值（保证编辑回显）
const resourceOpts = computed(() => {
  const opts = resourceOptions(form.value.cloud_type)
  const cur = form.value.resource_id
  if (cur && !opts.some((o) => o.value === cur)) {
    opts.push({ label: cur, value: cur })
  }
  return opts
})

// 资源 ID 提示由 placeholder 承载（移除 NAlert 说明块，保证表单三项等高对齐）

async function load() {
  try {
    targets.value = await request<TargetConfig[]>('/api/targets')
  } catch (e: any) {
    message.error(`加载目标失败: ${e.message}`)
  }
}

// 挂载：加载目标 + 凭据状态（refreshSettings 强制拉取保证提示新鲜）
onMounted(async () => {
  await Promise.all([load(), refreshSettings(), loadZones(), loadScanned(form.value.cloud_type)])
  updateCredWarning()
})

function openAdd() {
  editingId.value = null
  form.value = { cloud_type: 'tc_lighthouse', region: '', resource_id: null }
  showModal.value = true
  updateCredWarning()
}

function openEdit(row: TargetConfig) {
  editingId.value = row.id
  form.value = { cloud_type: row.cloud_type, region: row.region, resource_id: row.resource_id }
  showModal.value = true
  updateCredWarning()
}

async function saveTarget() {
  const method = editingId.value ? 'PUT' : 'POST'
  const url = editingId.value ? `/api/targets/${editingId.value}` : '/api/targets'
  try {
    await request(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ cloud_type: form.value.cloud_type, region: form.value.region, resource_id: form.value.resource_id }),
    })
    showModal.value = false
    message.success(editingId.value ? '更新成功' : '添加成功')
    load()
  } catch (e: any) {
    message.error(`保存失败: ${e.message}`) // 修复：非 2xx 不再误报成功
  }
}

async function deleteTarget(row: TargetConfig) {
  try {
    await request(`/api/targets/${row.id}`, { method: 'DELETE' })
    message.success('删除成功')
    load()
  } catch (e: any) {
    message.error(`删除失败: ${e.message}`) // 修复：非 2xx 不再误报成功
  }
}

// 弹窗内表单级「测试连接」：用未保存的表单值验证（保留），15s 超时
async function testConnection() {
  testResult.value = '测试中...'
  try {
    const data = await request<TestConnectionResult>(
      '/api/test-connection',
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ cloud_type: form.value.cloud_type, region: form.value.region, resource_id: form.value.resource_id }),
      },
      15000
    )
    testResult.value = data.success ? (data.message || '连接成功') : `失败: ${data.error || '未知错误'}`
  } catch (e: any) {
    testResult.value = e?.message === '请求超时'
      ? '连接超时（15 秒），请检查网络或云 API 状态'
      : `请求失败: ${e.message}`
  }
}

const columns = [
  { title: '#', key: 'index', render: (_: any, i: number) => i + 1 },
  { title: '云产品', key: 'cloud_type', render: (row: any) => cloudLabelMap[row.cloud_type] || row.cloud_type },
  { title: '资源ID', key: 'resource_id' },
  { title: '地域', key: 'region' },
  {
    title: '操作', key: 'actions',
    render(row: any) {
      return h(NSpace, { size: 'small' }, {
        default: () => [
          h(NButton, { size: 'tiny', onClick: () => openEdit(row) }, { default: () => '编辑' }),
          h(NButton, { size: 'tiny', type: 'error', onClick: () => deleteTarget(row) }, { default: () => '删除' }),
        ]
      })
    }
  },
]
</script>

<template>
  <div>
    <NSpace justify="space-between" align="center">
      <h2>云资源管理</h2>
      <NButton type="primary" size="small" @click="openAdd">添加目标</NButton>
    </NSpace>
    <NDataTable :columns="columns" :data="targets" :bordered="true" />

    <NModal v-model:show="showModal" :title="editingId ? '编辑目标' : '添加目标'" preset="card" style="width: 500px">
      <NForm :model="form" label-placement="left" label-width="80">
        <!-- Keys 缺失提示（改进 12） -->
        <NAlert v-if="credWarning" type="warning" style="margin-bottom: 12px">
          {{ credWarning }}
          <NButton text type="primary" size="small" style="margin-left: 8px" @click="router.push('/settings')">去设置</NButton>
        </NAlert>
        <NFormItem label="云产品">
          <NSelect v-model:value="form.cloud_type" :options="cloudOptions" />
        </NFormItem>
        <NFormItem label="资源ID">
          <NSelect
            v-model:value="form.resource_id"
            :options="resourceOpts"
            :placeholder="resourceIdHint(form.cloud_type)"
            filterable
            tag
            clearable
          />
        </NFormItem>
        <NFormItem label="地域">
          <NSelect
            v-model:value="form.region"
            :options="regionOpts"
            filterable
            tag
            clearable
            placeholder="选择或输入地域 ID（如 ap-guangzhou）"
          />
        </NFormItem>
        <NSpace>
          <NButton type="primary" @click="saveTarget">保存</NButton>
          <NButton @click="testConnection">测试连接</NButton>
        </NSpace>
        <p v-if="testResult" style="margin-top: 8px; color: #666">{{ testResult }}</p>
      </NForm>
    </NModal>
  </div>
</template>
