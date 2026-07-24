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

// L2/L3: the model proposes an action, the panel shows Approve/Reject, and
// approving drives confirm -> execute -> the DOM action actually running ->
// observe -> the loop continues to a final answer instead of stopping dead
// after "execute" (proving L3's propose/confirm/execute/observe cycle, not
// just a one-shot L2 exchange).
test('action loop: propose -> approve -> execute -> observe -> final answer', async ({ page }) => {
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

  // -> POST /observe -> the backend's loop takes one more turn -> a final
  // plain-text "chat" event, proving the task actually continued after
  // execute rather than ending there.
  await expect(page.getByText('All done — I clicked the Submit button for you.')).toBeVisible({
    timeout: 20_000,
  });
});

// Graded risk tiers: a RiskAuto tool (fill) must run without ever showing the
// Approve/Reject confirm card — only RiskConfirm tools (click, tested above)
// stop for the user.
test('risk tiers: a low-risk fill action auto-executes without confirmation', async ({ page }) => {
  await page.addInitScript(() => {
    // @ts-expect-error minimal stub of the chrome extension API
    window.chrome = {
      tabs: { query: async () => [{ id: 1 }] },
      scripting: { executeScript: async () => [{ result: true }] }, // pretend the fill succeeded
    };
  });

  await page.goto(PANEL);
  await page.getByPlaceholder('Ask Charli…').fill('please fill in my email');
  await page.getByTitle('Send').click();

  // Straight to "✓ Done." — no confirm card, no Approve/Reject buttons.
  await expect(page.getByText('✓ Done.')).toBeVisible({ timeout: 20_000 });
  await expect(page.getByRole('button', { name: 'Approve' })).not.toBeVisible();

  // The loop still continues afterward via observe, same as a confirmed action.
  await expect(page.getByText('All done — I filled that in for you.')).toBeVisible({
    timeout: 20_000,
  });
});

// L4: sheets_append is proposed and confirmed like any RiskConfirm action,
// but "performing" it means calling the backend's Google integration instead
// of touching the DOM. With no GOOGLE_CLIENT_ID/DATABASE_URL configured in
// this E2E environment, the integration reports itself unavailable — this
// proves that failure surfaces through the normal observe/loop machinery
// (a clear message, then the loop continues to a final answer) rather than
// crashing or hanging.
test('L4: sheets_append with no Google connection fails gracefully through the loop', async ({ page }) => {
  await page.addInitScript(() => {
    const store: Record<string, string> = {};
    // @ts-expect-error minimal stub of the chrome extension API
    window.chrome = {
      tabs: { query: async () => [{ id: 1 }] },
      scripting: { executeScript: async () => [{ result: true }] },
      // A real browser always has chrome.storage; getDeviceId() needs it to
      // produce a real id — an empty deviceId fails the backend's required
      // binding, which would mask the failure this test is actually after.
      storage: {
        local: {
          get: async (key: string) => ({ [key]: store[key] }),
          set: async (items: Record<string, string>) => Object.assign(store, items),
        },
      },
    };
  });

  await page.goto(PANEL);
  await page.getByPlaceholder('Ask Charli…').fill('add this row to my spreadsheet');
  await page.getByTitle('Send').click();

  await expect(page.getByText('Add this row to your spreadsheet?')).toBeVisible({ timeout: 20_000 });
  await page.getByRole('button', { name: 'Approve' }).click();

  // The append call fails ("not connected") — the panel must surface that
  // specific detail, not a generic action-failure message.
  await expect(page.getByText(/google sheets isn't connected yet/i)).toBeVisible({ timeout: 20_000 });

  // The loop still continues afterward via observe, same as any other action.
  await expect(page.getByText('Noted — could not add that to your spreadsheet.')).toBeVisible({
    timeout: 20_000,
  });
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
