// Charli's perception (L1): read the visible text of the active tab so the
// backend can answer questions about the page the user is looking at.
//
// Returns '' when perception isn't available — restricted tabs (chrome://…),
// or non-extension contexts like unit tests and the E2E static server.
export async function readActivePageText(maxChars = 8000): Promise<string> {
  if (typeof chrome === 'undefined' || !chrome.tabs || !chrome.scripting) return '';

  try {
    const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
    if (!tab?.id) return '';

    const [injection] = await chrome.scripting.executeScript({
      target: { tabId: tab.id },
      func: () => document.body?.innerText ?? '',
    });

    const text = (injection?.result as string | undefined) ?? '';
    return text.slice(0, maxChars);
  } catch {
    return ''; // e.g. the tab disallows script injection
  }
}
