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

// L2: the model proposes an action, the panel shows Approve/Reject, and
// approving drives confirm -> execute -> the DOM action actually running.
test('action loop: propose -> approve -> execute performs the DOM action', async ({ page }) => {
  await page.addInitScript(() => {
    // @ts-expect-error minimal stub of the chrome extension API
    window.chrome = {
      tabs: { query: async () => [{ id: 1 }] },
      scripting: { executeScript: async () => [{ result: true }] }, // pretend the click succeeded
    };
  });

  await page.goto(PANEL);
  await page.getByPlaceholder('Ask Charli…').fill('please click the button');
  await page.getByTitle('Send').click();

  // The mock LLM proposes a click action; the panel must show the confirm UI.
  await expect(page.getByText('Click the Submit button?')).toBeVisible({ timeout: 20_000 });
  await expect(page.getByText('Click "Submit"')).toBeVisible();

  await page.getByRole('button', { name: 'Approve' }).click();

  // Approving -> POST /confirm -> "execute" event -> performAction (stubbed
  // true) -> the panel reports success.
  await expect(page.getByText('✓ Done.')).toBeVisible({ timeout: 20_000 });
});

// L2 (reject path): rejecting must cancel, and nothing should be performed.
test('action loop: propose -> reject cancels without executing', async ({ page }) => {
  let executed = false;
  await page.exposeFunction('__markExecuted', () => {
    executed = true;
  });
  await page.addInitScript(() => {
    // @ts-expect-error minimal stub of the chrome extension API
    window.chrome = {
      tabs: { query: async () => [{ id: 1 }] },
      scripting: {
        // readActivePageText (L1) also calls executeScript, with no `args` —
        // only flag calls that carry args, i.e. the domClickByText/domFill call.
        executeScript: async (params: { args?: unknown[] }) => {
          if (params.args) {
            // @ts-expect-error test hook injected above
            window.__markExecuted();
          }
          return [{ result: true }];
        },
      },
    };
  });

  await page.goto(PANEL);
  await page.getByPlaceholder('Ask Charli…').fill('please click the button');
  await page.getByTitle('Send').click();

  await expect(page.getByRole('button', { name: 'Reject' })).toBeVisible({ timeout: 20_000 });
  await page.getByRole('button', { name: 'Reject' }).click();

  await expect(page.getByText('Cancelled.')).toBeVisible({ timeout: 20_000 });
  expect(executed).toBe(false);
});
