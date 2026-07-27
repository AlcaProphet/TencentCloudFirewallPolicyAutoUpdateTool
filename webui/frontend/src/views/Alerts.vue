<script setup lang="ts">
import { NCard, NForm, NFormItem, NInput, NSwitch, NButton, useMessage } from 'naive-ui'
import { ref, onMounted } from 'vue'

const message = useMessage()

const email = ref({
  enabled: false,
  host: '',
  port: '587',
  username: '',
  password: '',
  from_addr: '',
  to_addr: '',
})

const webhook = ref({
  enabled: false,
  url: '',
  channel: 'dingtalk',
})

const saving = ref(false)

async function load() {
  const res = await fetch('/api/alerts')
  const data = await res.json()
  if (data.email) email.value = data.email
  if (data.webhook) webhook.value = data.webhook
}

onMounted(load)

async function save() {
  saving.value = true
  try {
    const res = await fetch('/api/alerts', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: email.value, webhook: webhook.value }),
    })
    if (res.ok) {
      message.success('保存成功')
    } else {
      const data = await res.json()
      message.error(`保存失败: ${data.error}`)
    }
  } catch (e: any) {
    message.error(`保存失败: ${e.message}`)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <h2>告警配置</h2>

    <NCard title="邮件告警" size="small" style="margin-bottom: 16px">
      <template #header-extra>
        <NSwitch v-model:value="email.enabled" />
      </template>
      <NForm :model="email" label-placement="left" label-width="100">
        <NFormItem label="SMTP 主机">
          <NInput v-model:value="email.host" placeholder="smtp.example.com" :disabled="!email.enabled" />
        </NFormItem>
        <NFormItem label="端口">
          <NInput v-model:value="email.port" placeholder="587" :disabled="!email.enabled" />
        </NFormItem>
        <NFormItem label="用户名">
          <NInput v-model:value="email.username" placeholder="user@example.com" :disabled="!email.enabled" />
        </NFormItem>
        <NFormItem label="密码">
          <NInput v-model:value="email.password" type="password" placeholder="SMTP 密码" :disabled="!email.enabled" />
        </NFormItem>
        <NFormItem label="发件人">
          <NInput v-model:value="email.from_addr" placeholder="noreply@example.com" :disabled="!email.enabled" />
        </NFormItem>
        <NFormItem label="收件人">
          <NInput v-model:value="email.to_addr" placeholder="admin@example.com（多人逗号分隔）" :disabled="!email.enabled" />
        </NFormItem>
      </NForm>
    </NCard>

    <NCard title="Webhook 告警" size="small" style="margin-bottom: 16px">
      <template #header-extra>
        <NSwitch v-model:value="webhook.enabled" />
      </template>
      <NForm :model="webhook" label-placement="left" label-width="100">
        <NFormItem label="Webhook URL">
          <NInput v-model:value="webhook.url" placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx" :disabled="!webhook.enabled" />
        </NFormItem>
        <NFormItem label="通知渠道">
          <NSelect v-model:value="webhook.channel" :options="[
            { label: '钉钉', value: 'dingtalk' },
            { label: '飞书', value: 'feishu' },
            { label: 'Slack', value: 'slack' },
          ]" :disabled="!webhook.enabled" />
        </NFormItem>
      </NForm>
    </NCard>

    <NButton type="primary" :loading="saving" @click="save">保存配置</NButton>
  </div>
</template>
