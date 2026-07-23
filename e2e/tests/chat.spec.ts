import { test, expect } from '@playwright/test';

const PANEL = 'http://localhost:5099/sidepanel.html';

// The whole MVP in one test: type into the real panel, and the (mock) LLM's
// reply must come back over SSE and render as an assistant bubble.
test('chat round-trip: a message gets an assistant reply end-to-end', async ({ page }) => {
  await page.goto(PANEL);

  await page.getByPlaceholder('Ask Charli…').fill('hello');
  await page.getByTitle('Send').click();

  await expect(page.getByText('Charli here — this is a test reply.')).toBeVisible({
    timeout: 20_000,
  });
});

// L1: the panel reads the active tab's text and sends it along; the (mock) LLM
// echoes it, proving page context flows perceive -> send -> backend -> prompt.
test('page perception: the active page text reaches the model', async ({ page }) => {
  // Stub the extension APIs the panel uses to read the active tab.
  await page.addInitScript(() => {
    // @ts-expect-error minimal stub of the chrome extension API
    window.chrome = {
      tabs: { query: async () => [{ id: 1 }] },
      scripting: { executeScript: async () => [{ result: 'The secret word is platypus.' }] },
    };
  });

  await page.goto(PANEL);
  await page.getByPlaceholder('Ask Charli…').fill('what does the page say?');
  await page.getByTitle('Send').click();

  await expect(page.getByText(/platypus/)).toBeVisible({ timeout: 20_000 });
});
