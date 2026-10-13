import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Wails 嵌入产物需使用相对路径引用资源
export default defineConfig({
  plugins: [vue()],
  base: './',
})
