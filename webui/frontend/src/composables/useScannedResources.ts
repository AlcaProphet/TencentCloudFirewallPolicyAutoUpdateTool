// useScannedResources 扫描资源共享逻辑：GET/POST/DELETE /api/scanned-resources
// 消费方：Settings.vue（卡片内扫描与结果列表）、Targets.vue / RunTest.vue（资源 ID 自动补全）
// 模块级单例缓存：按 cloud_type 缓存扫描结果，避免重复请求
import { ref } from 'vue'
import type { SelectOption } from 'naive-ui'
import { request } from '../api'
import type { ScannedResource } from '../types'

// 各云厂商的扫描结果缓存（key = cloud_type）
const cache = ref<Record<string, ScannedResource[]>>({})

// 拉取某云厂商的扫描结果（失败静默保留旧数据）
async function load(cloudType: string) {
  try {
    cache.value[cloudType] = await request<ScannedResource[]>(`/api/scanned-resources?cloud_type=${cloudType}`)
  } catch { /* 失败静默 */ }
}

// 扫描指定云厂商+地域并刷新缓存；返回错误信息（success=false 时返回 error 文案）
async function scan(cloudType: string, region: string): Promise<string | null> {
  try {
    const data = await request<{ success: boolean; error?: string }>('/api/scan-resources', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ cloud_type: cloudType, region }),
    })
    if (!data.success) return data.error || '扫描失败'
    await load(cloudType)
    return null
  } catch (e: any) {
    return `扫描请求失败: ${e.message}`
  }
}

// 清理某云厂商的扫描结果
async function clear(cloudType: string) {
  try {
    await request(`/api/scanned-resources?cloud_type=${cloudType}`, { method: 'DELETE' })
    cache.value[cloudType] = []
  } catch { /* 失败静默 */ }
}

// 清空全部缓存（「清空所有数据」后调用，避免残留旧数据）
function clearAllCache() {
  cache.value = {}
}

// 构造资源 ID 自动补全选项：label = 名称（资源ID·地域），value = 资源 ID
function resourceOptions(cloudType: string): SelectOption[] {
  return (cache.value[cloudType] || []).map((r) => ({
    label: r.resource_name ? `${r.resource_name}（${r.resource_id}·${r.region}）` : `${r.resource_id}（${r.region}）`,
    value: r.resource_id,
  }))
}

// 获取某云厂商的扫描结果列表（设置页展示用）
function resourcesOf(cloudType: string): ScannedResource[] {
  return cache.value[cloudType] || []
}

export function useScannedResources() {
  return { load, scan, clear, clearAllCache, resourceOptions, resourcesOf }
}
