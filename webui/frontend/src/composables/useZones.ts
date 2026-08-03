// useZones 地域数据共享逻辑：GET /api/zones 按 cloud_type 构造自动补全选项
// 消费方：Targets.vue（添加/编辑目标弹窗）
// 模块级单例状态：多组件共享一份数据，避免重复请求
import { ref } from 'vue'
import type { SelectOption } from 'naive-ui'
import { request } from '../api'
import type { ZoneRegion } from '../types'

const zonesMap = ref<Record<string, ZoneRegion[]>>({})
const loaded = ref(false)

// 加载地域数据（幂等：已加载不重复请求；失败静默，下次调用重试）
async function load() {
  if (loaded.value) return
  try {
    zonesMap.value = await request<Record<string, ZoneRegion[]>>('/api/zones')
    loaded.value = true
  } catch { /* 失败静默 */ }
}

// 按 cloud_type 构造 NSelect 选项：label = 中文地域名（地域 ID），value = 地域 ID
// 仅提供预填充建议，不限制输入（NSelect filterable+tag 允许任意值）
function regionOptions(cloudType: string): SelectOption[] {
  return (zonesMap.value[cloudType] || []).map((r) => ({ label: `${r.name}（${r.id}）`, value: r.id }))
}

export function useZones() {
  return { load, regionOptions }
}
