import { test, expect } from "@playwright/test";
import { deleteAllComments } from "../helpers/test-utils";

test.describe.configure({ mode: "serial" });
test.describe("Comments", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await expect(page.getByText("main.go")).toBeVisible();
    await deleteAllComments(page);
  });

  test("shows comment button on line hover and opens form", async ({
    page,
  }) => {
    // Hover a line to reveal the "+" button
    const addButton = page.getByLabel("Add comment").first();
    await addButton.click({ force: true });

    // CommentForm should appear
    await expect(page.getByLabel("Comment body")).toBeVisible();
  });

  test("creates a comment via form", async ({ page }) => {
    const addButton = page.getByLabel("Add comment").first();
    await addButton.click({ force: true });

    await page.getByLabel("Comment body").fill("Test comment from Playwright");
    await page.getByRole("button", { name: "Add Comment", exact: true }).click();

    // Comment should appear inline
    await expect(
      page.getByText("Test comment from Playwright")
    ).toBeVisible();
  });

  test("cancels comment form without creating", async ({ page }) => {
    const addButton = page.getByLabel("Add comment").first();
    await addButton.click({ force: true });

    await page.getByLabel("Comment body").fill("Should not be saved");
    await page.getByRole("button", { name: "Cancel" }).click();

    // Form should disappear
    await expect(page.getByLabel("Comment body")).toHaveCount(0);
    // Comment should not be created
    await expect(page.getByText("Should not be saved")).toHaveCount(0);
  });

  test("deletes a comment", async ({ page }) => {
    // First create a comment
    const addButton = page.getByLabel("Add comment").first();
    await addButton.click({ force: true });
    await page.getByLabel("Comment body").fill("Comment to delete");
    await page.getByRole("button", { name: "Add Comment", exact: true }).click();
    await expect(page.getByText("Comment to delete")).toBeVisible();

    // Auto-accept the confirm dialog when it appears
    page.on("dialog", (d) => d.accept());

    // Delete the comment
    await page.getByRole("button", { name: "Delete" }).first().click();

    // Verify deletion via API (authoritative check, avoids split-view dual rendering)
    await expect(async () => {
      const res = await page.request.get(
        "http://127.0.0.1:4444/api/comments"
      );
      const comments = await res.json();
      expect(comments).toHaveLength(0);
    }).toPass({ timeout: 5000 });
  });
});
