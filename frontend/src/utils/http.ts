import type { ApiResponse, RequestResult } from '../types/api'

export interface RequestOptions extends RequestInit {
  timeout?: number
}

class HttpClient {
  private readonly baseUrl: string

  constructor(baseUrl = 'api') {
    this.baseUrl = baseUrl.replace(/\/+$/, '')
  }

  private resolveUrl(path: string): string {
    const cleanPath = path.replace(/^\/+/, '')
    return `${this.baseUrl}/${cleanPath}`
  }

  private async request<T>(path: string, options: RequestOptions = {}): Promise<RequestResult<T>> {
    const { timeout = 30000, ...fetchOptions } = options
    const url = this.resolveUrl(path)

    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), timeout)

    try {
      const response = await fetch(url, {
        ...fetchOptions,
        signal: controller.signal
      })

      clearTimeout(timer)

      const isJson = response.headers.get('content-type')?.includes('application/json')
      let resBody: ApiResponse<T> | Record<string, unknown> | null = null

      if (isJson) {
        try {
          resBody = await response.json()
        } catch {
          resBody = null
        }
      }

      if (response.ok) {
        if (resBody && typeof resBody === 'object' && 'code' in resBody) {
          const apiResp = resBody as ApiResponse<T>
          if (apiResp.code === 0) {
            return {
              success: true,
              data: apiResp.data,
              message: apiResp.message || 'success',
              timestamp: apiResp.timestamp
            }
          }
          return {
            success: false,
            code: apiResp.code,
            message: apiResp.message || `请求错误 (${apiResp.code})`
          }
        }
        return {
          success: true,
          data: resBody as unknown as T,
          message: 'success'
        }
      }

      const errMsg =
        (resBody as { message?: string } | null)?.message ||
        `请求失败 (HTTP ${response.status})`
      return {
        success: false,
        code: response.status,
        message: errMsg
      }
    } catch (err: unknown) {
      clearTimeout(timer)
      if (err instanceof DOMException && err.name === 'AbortError') {
        return {
          success: false,
          code: 408,
          message: '网络请求超时，请检查服务状态'
        }
      }
      return {
        success: false,
        code: -1,
        message: (err as Error)?.message || '网络连接异常，请重试'
      }
    }
  }

  get<T>(path: string, options?: RequestOptions): Promise<RequestResult<T>> {
    return this.request<T>(path, { ...options, method: 'GET' })
  }

  post<T>(path: string, data?: unknown, options?: RequestOptions): Promise<RequestResult<T>> {
    return this.request<T>(path, {
      ...options,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers
      },
      body: data !== undefined ? JSON.stringify(data) : undefined
    })
  }

  delete<T>(path: string, options?: RequestOptions): Promise<RequestResult<T>> {
    return this.request<T>(path, { ...options, method: 'DELETE' })
  }

  upload<T>(path: string, file: File, options?: RequestOptions): Promise<RequestResult<T>> {
    const formData = new FormData()
    formData.append('file', file)
    return this.request<T>(path, {
      ...options,
      method: 'POST',
      body: formData
    })
  }
}

export const http = new HttpClient()
