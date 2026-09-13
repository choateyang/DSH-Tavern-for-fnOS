/**
 * 通用防抖与节流工具
 */

export function debounce<T extends (...args: any[]) => any>(
  fn: T,
  wait = 500
): (...args: Parameters<T>) => void {
  let timer: ReturnType<typeof setTimeout> | null = null
  return function (this: any, ...args: Parameters<T>) {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => {
      fn.apply(this, args)
      timer = null
    }, wait)
  }
}

export function throttle<T extends (...args: any[]) => any>(
  fn: T,
  wait = 500
): (...args: Parameters<T>) => void {
  let lastTime = 0
  return function (this: any, ...args: Parameters<T>) {
    const now = Date.now()
    if (now - lastTime >= wait) {
      lastTime = now
      fn.apply(this, args)
    }
  }
}

/**
 * 异步防并发锁定包装器：在 Promise 未完成前自动忽略后续重复点击
 */
export function withAsyncLock<T extends (...args: any[]) => Promise<any>>(
  fn: T
): (...args: Parameters<T>) => Promise<ReturnType<T> | void> {
  let isRunning = false
  return async function (this: any, ...args: Parameters<T>) {
    if (isRunning) return
    isRunning = true
    try {
      return await fn.apply(this, args)
    } finally {
      isRunning = false
    }
  }
}
