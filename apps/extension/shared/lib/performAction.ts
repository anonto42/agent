import type { Action } from '@charli/shared';
import { domFillFirstTextInput, domClickByText } from './domActions';

type ActionHandler = (tabId: number, action: Action) => Promise<boolean>;

// One entry per action kind the backend's tool registry can propose. Adding a
// new kind means adding a handler here, not another if-chain.
const actionHandlers: Record<string, ActionHandler> = {
  fill: async (tabId, action) => {
    const [injection] = await chrome.scripting.executeScript({
      target: { tabId },
      func: domFillFirstTextInput,
      args: [action.value ?? ''],
    });
    return Boolean(injection?.result);
  },
  click: async (tabId, action) => {
    const [injection] = await chrome.scripting.executeScript({
      target: { tabId },
      func: domClickByText,
      args: [action.target ?? ''],
    });
    return Boolean(injection?.result);
  },
};

// Runs an approved action on the active tab. Only called after the backend's
// safety engine has already approved it AND the user has confirmed it — this
// function performs, it never decides.
export async function performAction(action: Action): Promise<boolean> {
  if (typeof chrome === 'undefined' || !chrome.tabs || !chrome.scripting) return false;

  const handler = actionHandlers[action.kind];
  if (!handler) return false;

  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab?.id) return false;

  try {
    return await handler(tab.id, action);
  } catch {
    return false; // e.g. the tab disallows script injection
  }
}
