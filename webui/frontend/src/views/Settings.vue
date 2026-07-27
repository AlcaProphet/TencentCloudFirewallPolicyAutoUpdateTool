<script setup lang="ts">
import { NForm, NFormItem, NInput, NSelect, NButton, NSpace, useMessage } from 'naive-ui'
import { ref, onMounted } from 'vue'

const settings = ref<Record<string, string>>({})
const message = useMessage()

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
  message.success('保存成功')
}

function exportConfig() {
  window.open('/api/config/export', '_blank')
}

async function importConfig(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const text = await file.text()
  try {
    const data = JSON.parse(text)
    const res = await fetch('/api/config/import', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    })
    const result = await res.json()
    if (res.ok) {
      message.success('导入成功')
      const s = await fetch('/api/settings')
      settings.value = await s.json() || {}
    } else {
      message.error(result.error || '导入失败')
    }
  } catch {
    message.error('JSON 格式错误')
  }
  input.value = ''
}
</script>

<template>
  <div>
    <h2>全局设置</h2>
    <NForm :model="settings" label-placement="left" label-width="160">
      <h3 style="margin: 0 0 12px">云厂商凭据</h3>
      <NFormItem label="腾讯云 SecretId">
        <NInput v-model:value="settings.tc_access_id" type="password" show-password-on="click" placeholder="AKIDxxx" />
      </NFormItem>
      <NFormItem label="腾讯云 SecretKey">
        <NInput v-model:value="settings.tc_access_key" type="password" show-password-on="click" placeholder="SecretKey" />
      </NFormItem>
      <NFormItem label="阿里云 AccessKeyId">
        <NInput v-model:value="settings.ali_access_id" type="password" show-password-on="click" placeholder="LTAIxxx" />
      </NFormItem>
      <NFormItem label="阿里云 AccessKeySecret">
        <NInput v-model:value="settings.ali_access_key" type="password" show-password-on="click" placeholder="AccessKeySecret" />
      </NFormItem>

      <h3 style="margin: 16px 0 12px">全局设置</h3>
      <NFormItem label="TAG">
        <NInput v-model:value="settings.tag" />
      </NFormItem>
      <NFormItem label="同步间隔">
        <NInput v-model:value="settings.interval" />
      </NFormItem>
      <NFormItem label="DNS 服务器">
        <NInput v-model:value="settings.dns" />
      </NFormItem>
      <NFormItem label="DNS 超时">
        <NInput v-model:value="settings.dns_timeout" placeholder="10s" />
      </NFormItem>
      <NFormItem label="日志级别">
        <NSelect v-model:value="settings.log_level" :options="[
          { label: 'Debug', value: 'debug' },
          { label: 'Info', value: 'info' },
          { label: 'Warn', value: 'warn' },
          { label: 'Error', value: 'error' },
        ]" />
      </NFormItem>
      <NFormItem>
        <NSpace>
          <NButton type="primary" @click="save">保存</NButton>
          <NButton @click="exportConfig">导出配置</NButton>
          <label>
            <NButton tag="span">导入配置</NButton>
            <input type="file" accept=".json" style="display: none" @change="importConfig" />
          </label>
        </NSpace>
      </NFormItem>
    </NForm>
  </div>
</template>
