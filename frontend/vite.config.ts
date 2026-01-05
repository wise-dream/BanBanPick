import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  // Загружаем переменные окружения из корневой директории проекта
  envDir: '..',
  server: {
    hmr: {
      protocol: 'ws',
      host: 'localhost',
      port: 5173,
    },
  },
})
