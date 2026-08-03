// 共享常量（消除 Targets.vue 与 Advanced.vue 的重复定义）
import type { SelectOption } from 'naive-ui'

// 云产品选项（与后端 config.CloudType 一致）
export const cloudOptions: SelectOption[] = [
  { label: '腾讯云轻量云', value: 'tc_lighthouse' },
  { label: '腾讯云CVM', value: 'tc_cvm' },
  { label: '阿里云轻量云', value: 'ali_swas' },
  { label: '阿里云ECS', value: 'ali_ecs' },
]

// 云类型 → 中文名映射
// 手写循环避免 Object.fromEntries 推导出 SelectOption 联合类型（与 Record<string, string> 不兼容）
export const cloudLabelMap: Record<string, string> = {}
for (const o of cloudOptions) {
  cloudLabelMap[String(o.value)] = String(o.label)
}

// 云类型 → 资源 ID 输入提示（placeholder 文案）
// 轻量云类填写实例 ID，CVM/ECS 类填写安全组 ID
// 供 Targets.vue 与 RunTest.vue 共用
const resourceIdHints: Record<string, string> = {
  tc_lighthouse: '实例ID（lhins- 开头）',
  tc_cvm: '安全组ID（sg- 开头）',
  ali_swas: '实例ID（UUID 格式）',
  ali_ecs: '安全组ID（sg- 开头）',
}

// 按云类型取资源 ID 提示文案（未知类型回退通用文案）
export function resourceIdHint(cloudType: string): string {
  return resourceIdHints[cloudType] || '实例ID / 安全组ID'
}
