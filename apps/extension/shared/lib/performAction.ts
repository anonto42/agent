import type { Action } from '@charli/shared';
import { domFillFirstTextInput, domClickByText } from './domActions';
import { getDeviceId } from './deviceId';

const BASE = 'http://localhost:8080/api/v1';

export interface ActionResult {
  success: boolean;
  detail?: string;
}

type ActionHandler = (tabId: number, action: Action) => Promise<ActionResult>;

// One entry per action kind the backend's tool registry can propose. Adding a
// new kind means adding a handler here, not another if-chain. Most kinds act
// on the DOM (fill, click); sheets_append (L4) instead calls a backend API —
// "performing" it just happens to mean that instead of touching the page.
const actionHandlers: Record<string, ActionHandler> = {
  fill: async (tabId, action) => {
    const [injection] = await chrome.scripting.executeScript({
      target: { tabId },
      func: domFillFirstTextInput,
      args: [action.value ?? ''],
    });
    return { success: Boolean(injection?.result) };
  },
  click: async (tabId, action) => {
    const [injection] = await chrome.scripting.executeScript({
      target: { tabId },
      func: domClickByText,
      args: [action.target ?? ''],
    });
    return { success: Boolean(injection?.result) };
  },
  sheets_append: async (_tabId, action) => {
    const deviceId = await getDeviceId();
    const res = await fetch(`${BASE}/integrations/google/append`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        deviceId,
        spreadsheetId: action.spreadsheetId ?? '',
        values: action.values ?? [],
      }),
    });
    const body = (await res.json()) as { data?: { success: boolean; detail?: string } };
    if (!body.data) return { success: false, detail: 'unexpected response from the server' };
    return { success: body.data.success, detail: body.data.detail };
  },
};

// Runs an approved action. Only called after the backend's safety engine has
// already approved it AND the user has confirmed it — this function
// performs, it never decides.
export async function performAction(action: Action): Promise<ActionResult> {
  if (typeof chrome === 'undefined' || !chrome.tabs || !chrome.scripting) return { success: false };

  const handler = actionHandlers[action.kind];
  if (!handler) return { success: false };

  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab?.id) return { success: false };

  try {
    return await handler(tab.id, action);
  } catch {
    return { success: false }; // e.g. the tab disallows script injection, or a network error
  }
}
