import { test, expect } from '@playwright/test';

// The whole MVP in one test: type into the real panel, and the (mock) LLM's
// reply must come back over SSE and render as an assistant bubble.
test('chat round-trip: a message gets an assistant reply end-to-end', async ({ page }) => {
  await page.goto('http://localhost:5099/sidepanel.html');

  await page.getByPlaceholder('Ask Charli…').fill('hello');
  await page.getByTitle('Send').click();

  await expect(page.getByText('Charli here — this is a test reply.')).toBeVisible({
    timeout: 20_000,
  });
});
