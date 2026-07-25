<script setup lang="ts">
import { NForm, NFormItem, NInput, NButton } from 'naive-ui'
import { ref, onMounted } from 'vue'

const settings = ref<Record<string, string>>({})

onMounted(async () => {
  const res = await fetch('/api/settings')
  settings.value = await res.json() || {}
})

async function save() {
  await fetch('/api/settings', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(settings.value),
  })
}
</script>

<template>
  <div>
    <h2>全局设置</h2>
    <NForm :model="settings" label-placement="left" label-width="120">
      <NFormItem label="TAG">
        <NInput v-model:value="settings.tag" placeholder="auto-dns" />
      </NFormItem>
      <NFormItem label="同步间隔">
        <NInput v-model:value="settings.interval" placeholder="5m" />
      </NFormItem>
      <NFormItem label="DNS 服务器">
        <NInput v-model:value="settings.dns" placeholder="8.8.8.8:53" />
      </NFormItem>
      <NFormItem>
        <NButton type="primary" @click="save">保存</NButton>
      </NFormItem>
    </NForm>
  </div>
</template>
