import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 6969,
    proxy: {
      '/api': 'http://localhost:6969',
      '/ws': {
        target: 'ws://localhost:6969',
        ws: true,
      },
    },
  },
  build: {
    outDir: 'build',
  },
});
