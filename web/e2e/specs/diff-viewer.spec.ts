import { test, expect } from "@playwright/test";

test.describe("DiffViewer", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    // Wait for diff to load
    await expect(page.locator('[id="diff-file-main.go"]')).toBeVisible();
  });

  test("displays file headers with paths", async ({ page }) => {
    await expect(page.locator('[id="diff-file-main.go"]')).toBeVisible();
    await expect(page.locator('[id="diff-file-utils.go"]')).toBeVisible();
    await expect(page.locator('[id="diff-file-old.go"]')).toBeVisible();
  });

  test("displays correct status badges", async ({ page }) => {
    await expect(
      page.getByLabel("File status: modified")
    ).toBeVisible();
    await expect(page.getByLabel("File status: added")).toBeVisible();
    await expect(
      page.getByLabel("File status: deleted")
    ).toBeVisible();
  });

  test("renders split view by default", async ({ page }) => {
    await expect(page.getByTestId("split-view").first()).toBeVisible();
  });

  test("renders added lines with correct data attribute", async ({
    page,
  }) => {
    const addedLines = page.locator('[data-line-type="add"]');
    await expect(addedLines.first()).toBeVisible();
  });

  test("renders deleted lines with correct data attribute", async ({
    page,
  }) => {
    const deletedLines = page.locator('[data-line-type="delete"]');
    await expect(deletedLines.first()).toBeVisible();
  });

  test("switches to unified view and back to split", async ({ page }) => {
    // Click Unified button
    await page.getByRole("button", { name: "Unified" }).click();
    // Split view should disappear
    await expect(page.getByTestId("split-view")).toHaveCount(0);

    // Click Split button to go back
    await page.getByRole("button", { name: "Split" }).click();
    await expect(page.getByTestId("split-view").first()).toBeVisible();
  });
});
