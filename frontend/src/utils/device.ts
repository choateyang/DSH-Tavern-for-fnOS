import { ref, onMounted, onUnmounted } from 'vue'

/**
 * 响应式触屏与移动端设备检测
 * 用于在移动端/触屏设备上自动禁用会阻碍点击的 Tooltip 气泡
 */
export function useIsTouchDevice() {
  const isTouch = ref(false)

  let mql: MediaQueryList | null = null
  const update = () => {
    if (typeof window === 'undefined') return
    isTouch.value = window.matchMedia('(hover: none), (pointer: coarse)').matches || window.innerWidth <= 640
  }

  onMounted(() => {
    update()
    mql = window.matchMedia('(hover: none), (pointer: coarse)')
    mql.addEventListener?.('change', update)
    window.addEventListener('resize', update, { passive: true })
  })

  onUnmounted(() => {
    mql?.removeEventListener?.('change', update)
    window.removeEventListener('resize', update)
  })

  return isTouch
}
