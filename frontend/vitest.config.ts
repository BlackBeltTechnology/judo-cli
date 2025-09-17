import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    setupFiles: './src/setupTests.tsx',
    globals: true,
    css: {
      modules: {
        classNameStrategy: 'non-scoped',
      },
      postcss: {},
    },
    transformMode: {
      web: [/\.(css|scss)$/],
    },
    alias: {
      'xterm/css/xterm.css': './__mocks__/xterm-css-mock.js',
    },
  },
});
