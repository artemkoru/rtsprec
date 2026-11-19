import { defineConfig } from '@playwright/test';

// E2E tests run against the Vite dev server, which proxies /api and /hls to the
// backend. E2E_BASE_URL points at another dev server; E2E_CHANNEL=chrome uses an
// installed Google Chrome instead of the bundled Chromium.
const baseURL = process.env.E2E_BASE_URL || 'http://localhost:5173';

export default defineConfig({
  testDir: './tests',
  globalSetup: './tests/global-setup.js',
  fullyParallel: false,
  workers: 1,
  reporter: [['list']],
  use: {
    baseURL,
    trace: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { channel: process.env.E2E_CHANNEL || undefined },
    },
  ],
  // Start the dev server only when no explicit E2E_BASE_URL is given; otherwise the
  // tests run against the server at that address and globalSetup checks it.
  webServer: process.env.E2E_BASE_URL
    ? undefined
    : {
        command: 'npm run dev -- --port 5173 --strictPort',
        url: baseURL,
        reuseExistingServer: true,
      },
});
