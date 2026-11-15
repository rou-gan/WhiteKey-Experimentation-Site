import { defineConfig, devices } from "@playwright/test";

/**
 * ローカル E2E（IT-MEM-3）。
 * 前提: backend :8080 と frontend :3000 が起動済み。
 * 例: NEXT_PUBLIC_API_BASE_URL=http://localhost:8080 npm run dev
 */
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: 0,
  workers: 1,
  timeout: 90_000,
  expect: { timeout: 15_000 },
  reporter: [["list"]],
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL ?? "http://localhost:3000",
    trace: "on-first-retry",
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
});
