// Charli's hands (L2): the actual DOM manipulation for a confirmed action.
//
// These functions are passed directly to chrome.scripting.executeScript, which
// serializes them and runs them INSIDE the target page. Because of that they
// must be fully self-contained — no closures over other module state, only
// `document` (a global in that page context) and their own arguments. That
// same property makes them plain, directly testable functions: in a jsdom
// test, `document` is the test's own global, so calling them normally exercises
// real DOM behaviour without needing the chrome API at all.

/** Fills the first visible text-like input/textarea on the page. */
export function domFillFirstTextInput(value: string): boolean {
  const el = document.querySelector<HTMLInputElement | HTMLTextAreaElement>(
    'input[type="text"], input:not([type]), input[type="email"], input[type="search"], textarea',
  );
  if (!el) return false;

  el.focus();
  el.value = value;
  el.dispatchEvent(new Event('input', { bubbles: true }));
  el.dispatchEvent(new Event('change', { bubbles: true }));
  return true;
}

/** Clicks the first button/link/role=button whose visible text matches `target`. */
export function domClickByText(target: string): boolean {
  const needle = target.trim().toLowerCase();
  if (!needle) return false;

  const candidates = Array.from(
    document.querySelectorAll<HTMLElement>(
      'button, a, [role="button"], input[type="submit"], input[type="button"]',
    ),
  );
  const match = candidates.find((el) => {
    const text = (el.textContent || (el as HTMLInputElement).value || '').trim().toLowerCase();
    return text.includes(needle);
  });
  if (!match) return false;

  match.scrollIntoView?.({ block: 'center' });
  match.click();
  return true;
}
