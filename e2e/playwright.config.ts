import { defineConfig } from "@playwright/test";

// Both functions are plain HTTP servers, so Playwright can run them itself.
// Pointing E2E_API_URL at something already running (a deployed stage, or a
// `make run-api` in another shell) skips that and tests what is there instead.
const managed = !process.env.E2E_API_URL && !process.env.E2E_PUBLIC_URL;

// `port` and not `url`: every route of the api function answers 401 without a
// key, and Playwright waits for a response under 400.
const functions = [
  { name: "api", port: 8080 },
  { name: "public", port: 8081 },
];

export default defineConfig({
  testDir: "./tests",
  fullyParallel: true,
  forbidOnly: Boolean(process.env.CI),
  retries: process.env.CI ? 2 : 0,
  timeout: 30_000,
  reporter: process.env.CI
    ? [["github"], ["html", { outputFolder: "reports", open: "never" }]]
    : [["list"]],
  webServer: managed
    ? functions.map(({ name, port }) => ({
        command: `go run ./lambda/${name}`,
        // the module root, this file sits in e2e/
        cwd: "..",
        env: { LISTEN_ADDR: `:${port}` },
        port,
        // `go run` compiles first, which is the slow part on a cold cache
        timeout: 180_000,
        reuseExistingServer: !process.env.CI,
        stdout: "pipe" as const,
        stderr: "pipe" as const,
      }))
    : undefined,
});
