<script setup lang="ts">
import { NDataTable, NButton, NModal, NForm, NFormItem, NInput, NSelect, NSpace, useMessage } from 'naive-ui'
import { ref, onMounted, h } from 'vue'

const targets = ref<any[]>([])
const showModal = ref(false)
const editingId = ref<number | null>(null)
const form = ref({ cloud_type: 'tc_lighthouse', region: '', resource_id: '' })
const testResult = ref('')
const message = useMessage()

const cloudOptions = [
  { label: '腾讯云轻量云', value: 'tc_lighthouse' },
  { label: '腾讯云CVM', value: 'tc_cvm' },
  { label: '阿里云轻量云', value: 'ali_swas' },
  { label: '阿里云ECS', value: 'ali_ecs' },
]

async function load() {
  const res = await fetch('/api/targets')
  targets.value = await res.json() || []
}

onMounted(load)

function openAdd() {
  editingId.value = null
  form.value = { cloud_type: 'tc_lighthouse', region: '', resource_id: '' }
  showModal.value = true
}

function openEdit(row: any) {
  editingId.value = row.id
  form.value = { cloud_type: row.cloud_type, region: row.region, resource_id: row.resource_id }
  showModal.value = true
}

async function saveTarget() {
  const method = editingId.value ? 'PUT' : 'POST'
  const url = editingId.value ? `/api/targets/${editingId.value}` : '/api/targets'
  await fetch(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ cloud_type: form.value.cloud_type, region: form.value.region, resource_id: form.value.resource_id }),
  })
  showModal.value = false
  message.success(editingId.value ? '更新成功' : '添加成功')
  load()
}

async function deleteTarget(row: any) {
  await fetch(`/api/targets/${row.id}`, { method: 'DELETE' })
  message.success('删除成功')
  load()
}

async function testConnection() {
  testResult.value = '测试中...'
  const res = await fetch('/api/test-connection', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ cloud_type: form.value.cloud_type, region: form.value.region, resource_id: form.value.resource_id }),
  })
  const data = await res.json()
  testResult.value = data.success ? data.message : `失败: ${data.error}`
}

const cloudLabelMap: Record<string, string> = Object.fromEntries(
  cloudOptions.map(o => [o.value, o.label])
)

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
        <NFormItem label="云产品">
          <NSelect v-model:value="form.cloud_type" :options="cloudOptions" />
        </NFormItem>
        <NFormItem label="资源ID">
          <NInput v-model:value="form.resource_id" placeholder="lhins-xxx / sg-xxx" />
        </NFormItem>
        <NFormItem label="地域">
          <NInput v-model:value="form.region" placeholder="ap-guangzhou" />
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
