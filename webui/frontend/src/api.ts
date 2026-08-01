// 统一请求封装：res.ok 检查 + error 提取 + 可选超时（AbortController）
// 全站经本模块发起请求，任何非 2xx 响应统一抛 RequestError，杜绝"失败误报成功"

// RequestError 请求失败错误（非 2xx 响应）
export class RequestError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.name = 'RequestError'
    this.status = status
  }
}

// request 统一请求封装
// - 非 2xx 响应抛 RequestError（error 字段取自后端 {"error": ...}）
// - timeoutMs 可选：连接测试传 15000，超时抛 "请求超时"
// - 未知 API 路径返回 HTML 时 res.json() 解析失败进 catch
export async function request<T>(url: string, opts: RequestInit = {}, timeoutMs?: number): Promise<T> {
  const controller = new AbortController()
  const timer = timeoutMs ? setTimeout(() => controller.abort(), timeoutMs) : null
  try {
    const res = await fetch(url, { ...opts, signal: controller.signal })
    let data: any = null
    try {
      data = await res.json()
    } catch {
      /* 非 JSON 响应（如 SPA 兜底 HTML） */
    }
    if (!res.ok) {
      throw new RequestError(res.status, data?.error || `请求失败 (${res.status})`)
    }
    return data as T
  } catch (e: any) {
    if (e instanceof RequestError) throw e
    if (e?.name === 'AbortError') throw new Error('请求超时')
    throw e
  } finally {
    if (timer) clearTimeout(timer)
  }
}
