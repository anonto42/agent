// A persistent identifier for this browser installation (L4), distinct from
// @shared/api/stream.ts's `session` (which is per-page-load and never
// persisted). Google connections need to survive panel reloads, so they're
// keyed by this instead — first real use of the `storage` permission already
// granted in wxt.config.ts.
const STORAGE_KEY = 'charli_device_id';

export async function getDeviceId(): Promise<string> {
  if (typeof chrome === 'undefined' || !chrome.storage) return '';

  const stored = await chrome.storage.local.get(STORAGE_KEY);
  const existing = stored[STORAGE_KEY] as string | undefined;
  if (existing) return existing;

  const id = crypto.randomUUID();
  await chrome.storage.local.set({ [STORAGE_KEY]: id });
  return id;
}
