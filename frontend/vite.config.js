import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 4002,
    host: true,
    proxy: {
      '/recipes': 'http://localhost:4003',
      '/parse-recipe': 'http://localhost:4003',
      '/meal-plans': 'http://localhost:4003',
      '/grocery-list': 'http://localhost:4003',
      '/grocery-lists': 'http://localhost:4003',
    },
  },
  build: {
    // Generate source maps for better debugging
    sourcemap: true,
    // Optimize chunk splitting
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          router: ['react-router'],
          query: ['@tanstack/react-query'],
        },
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './tests/setup.js',
  },
})
