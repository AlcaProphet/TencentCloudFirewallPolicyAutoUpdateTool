<script setup lang="ts">
// 应用外壳：侧边栏导航 + 全局主题（light/dark，DB 持久化，Build4 Step 5）
import { NLayout, NLayoutSider, NLayoutContent, NMenu, NConfigProvider, NMessageProvider, NSwitch } from 'naive-ui'
import { darkTheme } from 'naive-ui'
import { useRouter } from 'vue-router'
import { ref, onMounted } from 'vue'
import { useSettings } from './composables/useSettings'

const router = useRouter()
const activeKey = ref('/')
const { theme, applyTheme, setTheme } = useSettings()

const menuOptions = [
  { label: '仪表盘', key: '/' },
  { label: '云资源管理', key: '/targets' },
  { label: '域名规则', key: '/rules' },
  { label: '全局设置', key: '/settings' },
  { label: '同步日志', key: '/logs' },
  { label: '运行测试', key: '/run-test' },
  { label: '告警配置', key: '/alerts' },
]

function handleMenuUpdate(key: string) {
  activeKey.value = key
  router.push(key)
}

onMounted(applyTheme)
</script>

<template>
  <NConfigProvider :theme="theme === 'dark' ? darkTheme : null">
    <NMessageProvider>
      <NLayout has-sider style="height: 100vh">
        <NLayoutSider bordered :width="200">
          <div style="padding: 16px 16px 0; font-weight: bold; font-size: 18px; display: flex; justify-content: space-between; align-items: center">
            <span>FWAlizer</span>
            <!-- 全局主题切换（Build4 Step 5：改进 1） -->
            <NSwitch size="small" :value="theme === 'dark'" @update:value="(v: boolean) => setTheme(v ? 'dark' : 'light')" />
          </div>
          <NMenu
            style="margin-top: 12px"
            :value="activeKey"
            :options="menuOptions"
            @update:value="handleMenuUpdate"
          />
        </NLayoutSider>
        <NLayoutContent content-style="padding: 24px;">
          <router-view />
        </NLayoutContent>
      </NLayout>
    </NMessageProvider>
  </NConfigProvider>
</template>
