import { resolve } from 'node:path';
import { defineConfig } from 'wxt';
import tailwindcss from '@tailwindcss/vite';

// WXT builds one codebase into a Chrome + Firefox MV3 extension.
// https://wxt.dev
export default defineConfig({
  modules: ['@wxt-dev/module-react'],
  vite: () => ({
    plugins: [tailwindcss()],
  }),
  // FSD aliases (same discipline as the website).
  alias: {
    '@features': resolve('features'),
    '@shared': resolve('shared'),
    '@entities': resolve('entities'),
  },
  manifest: {
    name: 'Charli',
    description: 'Charli — your flexible browser agent.',
    permissions: ['sidePanel', 'activeTab', 'storage', 'scripting'],
    host_permissions: ['<all_urls>'],
    action: {},
    side_panel: {
      default_path: 'sidepanel.html',
    },
  },
});
