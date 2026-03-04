import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./specs",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: "html",
  outputDir: "./test-results",
  use: {
    baseURL: "http://127.0.0.1:4444",
    trace: "on-first-retry",
  },
  projects: [
    {
      name: "chromium",
      use: { browserName: "chromium" },
    },
  ],
  webServer: {
    command:
      "cd ../.. && npm run --prefix web build && rm -rf internal/embed/dist && cp -r web/dist internal/embed/dist && go run ./cmd/e2e-server --port 4444",
    port: 4444,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
