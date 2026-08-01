<script setup lang="ts">
import { NForm, NFormItem, NInput, NSelect, NButton, NSpace, useMessage } from 'naive-ui'
import { ref, onMounted } from 'vue'
import { request } from '../api'

const settings = ref<Record<string, string>>({})
const message = useMessage()

// 间隔格式校验（如 30s、5m、1h、500ms），保存前拦截非法值
const intervalPattern = /^\d+(ms|s|m|h)$/

onMounted(async () => {
  try {
    settings.value = await request<Record<string, string>>('/api/settings')
  } catch (e: any) {
    message.error(`加载设置失败: ${e.message}`)
  }
})

async function save() {
  if (!intervalPattern.test(String(settings.value.interval || ''))) {
    message.error('同步间隔格式无效，示例：30s / 5m / 1h')
    return
  }
  try {
    await request('/api/settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(settings.value),
    })
    message.success('保存成功')
  } catch (e: any) {
    message.error(`保存失败: ${e.message}`) // 修复：非 2xx 不再误报成功
  }
}

function exportConfig() {
  window.open('/api/config/export', '_blank')
}

async function importConfig(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  // 先解析 JSON（独立 try/catch：格式错误与请求错误提示分离）
  let data: any
  try {
    data = JSON.parse(await file.text())
  } catch {
    message.error('JSON 格式错误')
    input.value = ''
    return
  }
  try {
    await request('/api/config/import', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    })
    message.success('导入成功')
  } catch (err: any) {
    message.error(`导入失败: ${err.message}`) // RequestError 携带后端 error 信息
    input.value = ''
    return
  }
  // 导入成功后刷新设置表单（失败不影响导入结果）
  try {
    settings.value = await request<Record<string, string>>('/api/settings')
  } catch (err: any) {
    message.error(`导入成功，但刷新设置失败: ${err.message}`)
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
        <NInput v-model:value="settings.interval" placeholder="5m（30s / 5m / 1h）" />
      </NFormItem>
      <NFormItem label="DNS 服务器">
        <NInput v-model:value="settings.dns" />
      </NFormItem>
      <NFormItem label="DNS 超时">
        <NInput v-model:value="settings.dns_timeout" placeholder="10s" />
      </NFormItem>
      <NFormItem label="DNS 失败阈值">
        <NInput v-model:value="settings.dns_fail_threshold" placeholder="5" />
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
