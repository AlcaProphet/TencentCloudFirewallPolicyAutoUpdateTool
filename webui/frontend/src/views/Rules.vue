<script setup lang="ts">
import { NDataTable, NButton, NModal, NForm, NFormItem, NInput, NSelect, NSpace, NSwitch, NTag, useMessage } from 'naive-ui'
import { ref, onMounted, h } from 'vue'

const rules = ref<any[]>([])
const showModal = ref(false)
const editingId = ref<number | null>(null)
const form = ref({ host: '', protocol: 'TCP', ports: '', action: 'ACCEPT', comment: '', enable_ipv6: false })
const message = useMessage()

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

async function load() {
  const res = await fetch('/api/rules')
  rules.value = await res.json() || []
}

onMounted(load)

function openAdd() {
  editingId.value = null
  form.value = { host: '', protocol: 'TCP', ports: '', action: 'ACCEPT', comment: '', enable_ipv6: false }
  showModal.value = true
}

function openEdit(row: any, index: number) {
  editingId.value = index + 1
  form.value = { host: row.host, protocol: row.protocol, ports: row.ports, action: row.action, comment: row.comment || '', enable_ipv6: !!row.enable_ipv6 }
  showModal.value = true
}

async function saveRule() {
  const method = editingId.value ? 'PUT' : 'POST'
  const url = editingId.value ? `/api/rules/${editingId.value}` : '/api/rules'
  await fetch(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(form.value),
  })
  showModal.value = false
  message.success(editingId.value ? '更新成功' : '添加成功')
  load()
}

async function deleteRule(index: number) {
  await fetch(`/api/rules/${index + 1}`, { method: 'DELETE' })
  message.success('删除成功')
  load()
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
    title: '操作', key: 'actions',
    render(row: any, index: number) {
      return h(NSpace, { size: 'small' }, {
        default: () => [
          h(NButton, { size: 'tiny', onClick: () => openEdit(row, index) }, { default: () => '编辑' }),
          h(NButton, { size: 'tiny', type: 'error', onClick: () => deleteRule(index) }, { default: () => '删除' }),
        ]
      })
    }
  },
]
</script>

<template>
  <div>
    <NSpace justify="space-between" align="center">
      <h2>域名规则</h2>
      <NButton type="primary" size="small" @click="openAdd">添加规则</NButton>
    </NSpace>
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
          <NInput v-model:value="form.ports" placeholder="443,80 / 8000-8010 / ALL" />
        </NFormItem>
        <NFormItem label="动作">
          <NSelect v-model:value="form.action" :options="actionOptions" />
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
