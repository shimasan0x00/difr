import { test, expect } from "@playwright/test";

test.describe("App States", () => {
  test("displays difr header title", async ({ page }) => {
    await page.goto("/");
    await expect(page.getByText("difr").first()).toBeVisible();
  });

  test("shows loading state when API is slow", async ({ page }) => {
    // Intercept diff API and delay response
    await page.route("**/api/diff", async (route) => {
      await new Promise((r) => setTimeout(r, 2000));
      await route.continue();
    });

    await page.goto("/");
    await expect(page.getByText("Loading diff...")).toBeVisible();
  });

  test("shows error state when API fails", async ({ page }) => {
    // Intercept diff API and return 500
    await page.route("**/api/diff", (route) =>
      route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: "Internal Server Error" }),
      })
    );

    await page.goto("/");
    await expect(page.getByText(/Error:/)).toBeVisible({ timeout: 10_000 });
  });
});
