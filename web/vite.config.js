import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// Backend address for `npm run dev`. The browser talks to the Vite server only,
// so /api and /hls requests are proxied to the backend on the same origin.
const backend = process.env.BACKEND_URL || 'http://localhost:8080'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': backend,
      '/hls': backend,
    },
  },
})
