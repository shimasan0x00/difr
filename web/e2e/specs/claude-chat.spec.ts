import { test, expect } from "@playwright/test";

test.describe("Claude Chat", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await expect(page.locator('[id="diff-file-main.go"]')).toBeVisible();
  });

  test("displays chat panel with input and send button", async ({ page }) => {
    await expect(page.getByLabel("Message to Claude")).toBeVisible();
    await expect(page.getByRole("button", { name: "Send" })).toBeVisible();
  });

  test("sends a chat message and receives mock response", async ({
    page,
  }) => {
    await page.getByLabel("Message to Claude").fill("Hello Claude");
    await page.getByRole("button", { name: "Send" }).click();

    // Should show the user message
    await expect(page.getByText("Hello Claude")).toBeVisible();

    // Should receive mock response from the E2E server
    await expect(
      page.getByText("This is a mock response from Claude for E2E testing.")
    ).toBeVisible({ timeout: 10_000 });
  });

  test("triggers auto review and receives mock response", async ({
    page,
  }) => {
    await page.getByRole("button", { name: "Auto Review" }).click();

    // Should show review result
    await expect(
      page.getByText("Consider adding error handling")
    ).toBeVisible({ timeout: 10_000 });
  });
});
