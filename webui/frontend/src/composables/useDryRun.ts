// Dry Run 共享组合逻辑：loading / results / warnings / error / lastRunAt / run()
import { ref } from 'vue'
import { request } from '../api'
import type { DryRunResponse } from '../types'

export function useDryRun() {
  const loading = ref(false)
  const results = ref<DryRunResponse['results']>([])
  const warnings = ref<string[]>([])
  const error = ref('')
  const lastRunAt = ref<Date | null>(null)

  // 执行 Dry Run；失败时抛错由页面 message.error 展示
  async function run() {
    loading.value = true
    error.value = ''
    results.value = []
    warnings.value = []
    try {
      const data = await request<DryRunResponse>('/api/sync/dryrun', { method: 'POST' })
      results.value = data.results || []
      warnings.value = data.warnings || []
      lastRunAt.value = new Date()
    } catch (e: any) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  return { loading, results, warnings, error, lastRunAt, run }
}
