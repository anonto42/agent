import { resolve } from 'node:path';
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

const dir = import.meta.dirname;

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./vitest.setup.ts'],
    include: ['features/**/*.test.{ts,tsx}', 'shared/**/*.test.{ts,tsx}'],
  },
  resolve: {
    alias: {
      '@features': resolve(dir, 'features'),
      '@shared': resolve(dir, 'shared'),
      '@entities': resolve(dir, 'entities'),
    },
  },
});
