import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  fullyParallel: true,
  forbidOnly: Boolean(process.env.CI),
  retries: process.env.CI ? 2 : 0,
  timeout: 30_000,
  reporter: process.env.CI
    ? [["github"], ["html", { outputFolder: "reports", open: "never" }]]
    : [["list"]],
});
