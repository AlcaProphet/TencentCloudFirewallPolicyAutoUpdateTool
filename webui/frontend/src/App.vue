<script setup lang="ts">
import { NLayout, NLayoutSider, NLayoutContent, NMenu, NConfigProvider } from 'naive-ui'
import { useRouter } from 'vue-router'
import { ref } from 'vue'

const router = useRouter()
const activeKey = ref('/')

const menuOptions = [
  { label: '仪表盘', key: '/' },
  { label: '云资源管理', key: '/targets' },
  { label: '域名规则', key: '/rules' },
  { label: '全局设置', key: '/settings' },
  { label: '同步日志', key: '/logs' },
]

function handleMenuUpdate(key: string) {
  activeKey.value = key
  router.push(key)
}
</script>

<template>
  <NConfigProvider>
    <NLayout has-sider style="height: 100vh">
      <NLayoutSider bordered :width="200">
        <div style="padding: 16px; font-weight: bold; font-size: 18px">FWAlizer</div>
        <NMenu
          :value="activeKey"
          :options="menuOptions"
          @update:value="handleMenuUpdate"
        />
      </NLayoutSider>
      <NLayoutContent content-style="padding: 24px;">
        <router-view />
      </NLayoutContent>
    </NLayout>
  </NConfigProvider>
</template>
