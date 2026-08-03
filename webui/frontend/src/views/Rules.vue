<script setup lang="ts">
import { NDataTable, NButton, NModal, NForm, NFormItem, NInput, NSelect, NSpace, NSwitch, NTag, useMessage } from 'naive-ui'
import { ref, onMounted, h, watch } from 'vue'
import { request } from '../api'
import { cloudLabelMap } from '../constants'
import type { DomainRule } from '../types'

const rules = ref<DomainRule[]>([])
const showModal = ref(false)
const editingId = ref<number | null>(null)
const form = ref({ host: '', protocol: 'TCP', ports: '', action: 'ACCEPT', comment: '', enable_ipv6: false, targets: [] as number[] })
const message = useMessage()
const targetOptions = ref<{ label: string; value: number }[]>([])

const protocolOptions = [
  { label: 'TCP', value: 'TCP' },
  { label: 'UDP', value: 'UDP' },
  { label: 'TCP+UDP', value: 'TCP+UDP' },
  { label: 'ICMP', value: 'ICMP' },
]

const actionOptions = [
  { label: 'ACCEPT', value: 'ACCEPT' },
  { label: 'DROP', value: 'DROP' },
]

// ICMP 协议时自动将端口设为 ALL
watch(() => form.value.protocol, (newProto) => {
  if (newProto === 'ICMP') {
    form.value.ports = 'ALL'
  }
})

async function load() {
  try {
    rules.value = await request<DomainRule[]>('/api/rules')
  } catch (e: any) {
    message.error(`加载规则失败: ${e.message}`)
  }
}

async function loadTargets() {
  try {
    const data = await request<any[]>('/api/targets')
    targetOptions.value = data.map((t: any) => ({
      label: `${cloudLabelMap[t.cloud_type] || t.cloud_type} / ${t.resource_id}`,
      value: t.id,
    }))
  } catch (e: any) {
    message.error(`加载目标失败: ${e.message}`)
  }
}

onMounted(async () => {
  await load()
  await loadTargets()
})

function openAdd() {
  editingId.value = null
  form.value = { host: '', protocol: 'TCP', ports: '', action: 'ACCEPT', comment: '', enable_ipv6: false, targets: [] }
  showModal.value = true
}

function openEdit(row: any) {
  editingId.value = row.id
  form.value = { host: row.host, protocol: row.protocol, ports: row.ports, action: row.action, comment: row.comment || '', enable_ipv6: !!row.enable_ipv6, targets: Array.isArray(row.targets) ? [...row.targets] : [] }
  showModal.value = true
}

async function saveRule() {
  const method = editingId.value ? 'PUT' : 'POST'
  const url = editingId.value ? `/api/rules/${editingId.value}` : '/api/rules'
  try {
    await request(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(form.value),
    })
    showModal.value = false
    message.success(editingId.value ? '更新成功' : '添加成功')
    load()
  } catch (e: any) {
    message.error(`保存失败: ${e.message}`) // 修复：非 2xx 不再误报成功
  }
}

async function deleteRule(row: any) {
  try {
    await request(`/api/rules/${row.id}`, { method: 'DELETE' })
    message.success('删除成功')
    load()
  } catch (e: any) {
    message.error(`删除失败: ${e.message}`) // 修复：非 2xx 不再误报成功
  }
}

const columns = [
  { title: '#', key: 'index', render: (_: any, i: number) => i + 1 },
  { title: '域名', key: 'host' },
  { title: '协议', key: 'protocol' },
  { title: '端口', key: 'ports' },
  { title: '动作', key: 'action' },
  {
    title: 'IP版本', key: 'enable_ipv6',
    render(row: any) {
      return h(NTag, { size: 'small', type: row.enable_ipv6 ? 'info' : 'default' }, { default: () => row.enable_ipv6 ? 'IPv4+6' : '仅IPv4' })
    }
  },
  { title: '备注', key: 'comment' },
  {
    title: '适用目标', key: 'targets',
    render(row: any) {
      if (!row.targets || row.targets.length === 0) return '全部'
      return row.targets.map((id: number) => {
        const opt = targetOptions.value.find(o => o.value === id)
        return opt ? opt.label : `#${id}`
      }).join(', ')
    }
  },
  {
    title: '操作', key: 'actions',
    render(row: any) {
      return h(NSpace, { size: 'small' }, {
        default: () => [
          h(NButton, { size: 'tiny', onClick: () => openEdit(row) }, { default: () => '编辑' }),
          h(NButton, { size: 'tiny', type: 'error', onClick: () => deleteRule(row) }, { default: () => '删除' }),
        ]
      })
    }
  },
]
</script>

<template>
  <div>
    <h2>域名规则</h2>
    <NButton type="primary" size="large" style="margin: 8px 0 12px" @click="openAdd">添加规则</NButton>
    <NDataTable :columns="columns" :data="rules" :bordered="true" />

    <NModal v-model:show="showModal" :title="editingId ? '编辑规则' : '添加规则'" preset="card" style="width: 500px">
      <NForm :model="form" label-placement="left" label-width="60">
        <NFormItem label="域名">
          <NInput v-model:value="form.host" placeholder="api.example.com" />
        </NFormItem>
        <NFormItem label="协议">
          <NSelect v-model:value="form.protocol" :options="protocolOptions" />
        </NFormItem>
        <NFormItem label="端口">
          <NInput v-model:value="form.ports" :placeholder="form.protocol === 'ICMP' ? 'ICMP 协议固定为 ALL' : '443,80 / 8000-8010 / ALL'" :disabled="form.protocol === 'ICMP'" />
        </NFormItem>
        <NFormItem label="动作">
          <NSelect v-model:value="form.action" :options="actionOptions" />
        </NFormItem>
        <NFormItem label="适用目标">
          <NSelect v-model:value="form.targets" :options="targetOptions" multiple placeholder="留空 = 全部目标" clearable />
        </NFormItem>
        <NFormItem label="备注">
          <NInput v-model:value="form.comment" placeholder="可选" />
        </NFormItem>
        <NFormItem label="解析IPv6">
          <NSwitch v-model:value="form.enable_ipv6" />
          <span style="margin-left: 8px; font-size: 12px; color: #999">{{ form.enable_ipv6 ? '同时使用 A + AAAA 记录' : '仅使用 A 记录（IPv4）' }}</span>
        </NFormItem>
        <NButton type="primary" @click="saveRule">保存</NButton>
      </NForm>
    </NModal>
  </div>
</template>
