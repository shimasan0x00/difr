import type { Page } from "@playwright/test";

const BASE = "http://127.0.0.1:4444";

/** Delete all comments via API to isolate tests. */
export async function deleteAllComments(page: Page): Promise<void> {
  const res = await page.request.get(`${BASE}/api/comments`);
  const comments = await res.json();
  for (const c of comments) {
    await page.request.delete(`${BASE}/api/comments/${c.id}`);
  }
}
