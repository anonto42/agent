import type { Action } from '@charli/shared';
import { domFillFirstTextInput, domClickByText } from './domActions';

// Runs an approved action on the active tab. Only called after the backend's
// safety engine has already approved it AND the user has confirmed it — this
// function performs, it never decides.
export async function performAction(action: Action): Promise<boolean> {
  if (typeof chrome === 'undefined' || !chrome.tabs || !chrome.scripting) return false;

  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab?.id) return false;

  try {
    if (action.kind === 'fill') {
      const [injection] = await chrome.scripting.executeScript({
        target: { tabId: tab.id },
        func: domFillFirstTextInput,
        args: [action.value ?? ''],
      });
      return Boolean(injection?.result);
    }
    if (action.kind === 'click') {
      const [injection] = await chrome.scripting.executeScript({
        target: { tabId: tab.id },
        func: domClickByText,
        args: [action.target ?? ''],
      });
      return Boolean(injection?.result);
    }
  } catch {
    return false; // e.g. the tab disallows script injection
  }
  return false;
}
