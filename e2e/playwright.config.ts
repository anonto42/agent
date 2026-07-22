import { defineConfig } from '@playwright/test';

// End-to-end: the REAL built extension panel → the REAL Go backend → a mock LLM.
// Playwright boots all three servers, then drives the panel like a user.
export default defineConfig({
  testDir: './tests',
  timeout: 60_000,
  fullyParallel: false,
  workers: 1,
  use: {
    browserName: 'chromium',
    viewport: { width: 400, height: 680 },
  },
  webServer: [
    {
      command: 'node mock-llm.mjs',
      port: 8099,
      reuseExistingServer: false,
      timeout: 30_000,
    },
    {
      // Prebuilt binary (via moon dep backend:build) so Playwright can stop it
      // cleanly — `go run` would orphan its child binary on teardown.
      command: './bin/api',
      cwd: '../apps/backend',
      env: {
        PORT: '8080',
        LLM_BASE_URL: 'http://localhost:8099/v1',
        LLM_MODEL: 'mock',
        LLM_API_KEY: 'test',
      },
      url: 'http://localhost:8080/api/v1/health',
      reuseExistingServer: false,
      timeout: 60_000,
    },
    {
      // Serve the built extension side panel as static files.
      command: 'python3 -m http.server 5099',
      cwd: '../apps/extension/.output/chrome-mv3',
      port: 5099,
      reuseExistingServer: false,
      timeout: 30_000,
    },
  ],
});
