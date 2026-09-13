import { defineConfig, type UserConfig, type ConfigEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { viteDevMock } from './src/mock/viteDevMock'

export default defineConfig(({ command }: ConfigEnv): UserConfig => {
  return {
    plugins: [
      vue(),
      ...(command === 'serve' ? [viteDevMock()] : [])
    ],
    base: './',
    build: {
      rollupOptions: {
        output: {
          manualChunks: {
            'vendor-vue': ['vue', 'pinia'],
            'vendor-naive': ['naive-ui', '@vicons/tabler'],
            'vendor-highlight': ['highlight.js/lib/core']
          }
        }
      }
    },
    server: {
      port: 3000
    }
  }
})
