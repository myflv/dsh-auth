import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  // 相对路径构建：登录页挂在 /<hash>/ 下时，资源也能正确解析到 /<hash>/assets/
  base: './',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
