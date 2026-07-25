<script setup lang="ts">
import { NDataTable, NButton, NModal, NForm, NFormItem, NInput, NSelect, NSpace, useMessage } from 'naive-ui'
import { ref, onMounted, h } from 'vue'

const rules = ref<any[]>([])
const showModal = ref(false)
const editingId = ref<number | null>(null)
const form = ref({ Host: '', Protocol: 'TCP', Ports: '', Action: 'ACCEPT', Comment: '' })
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
  form.value = { Host: '', Protocol: 'TCP', Ports: '', Action: 'ACCEPT', Comment: '' }
  showModal.value = true
}

function openEdit(row: any, index: number) {
  editingId.value = index + 1
  form.value = { Host: row.Host, Protocol: row.Protocol, Ports: row.Ports, Action: row.Action, Comment: row.Comment || '' }
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
  { title: '域名', key: 'Host' },
  { title: '协议', key: 'Protocol' },
  { title: '端口', key: 'Ports' },
  { title: '动作', key: 'Action' },
  { title: '备注', key: 'Comment' },
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
          <NInput v-model:value="form.Host" placeholder="api.example.com" />
        </NFormItem>
        <NFormItem label="协议">
          <NSelect v-model:value="form.Protocol" :options="protocolOptions" />
        </NFormItem>
        <NFormItem label="端口">
          <NInput v-model:value="form.Ports" placeholder="443,80 / 8000-8010 / ALL" />
        </NFormItem>
        <NFormItem label="动作">
          <NSelect v-model:value="form.Action" :options="actionOptions" />
        </NFormItem>
        <NFormItem label="备注">
          <NInput v-model:value="form.Comment" placeholder="可选" />
        </NFormItem>
        <NButton type="primary" @click="saveRule">保存</NButton>
      </NForm>
    </NModal>
  </div>
</template>
