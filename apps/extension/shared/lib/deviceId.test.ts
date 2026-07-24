import { describe, it, expect, vi, afterEach } from 'vitest';
import { getDeviceId } from './deviceId';

afterEach(() => {
  (globalThis as { chrome?: unknown }).chrome = undefined;
});

function stubChromeStorage(initial: Record<string, unknown> = {}) {
  const store = { ...initial };
  (globalThis as { chrome?: unknown }).chrome = {
    storage: {
      local: {
        get: vi.fn().mockImplementation(async (key: string) => ({ [key]: store[key] })),
        set: vi.fn().mockImplementation(async (items: Record<string, unknown>) => {
          Object.assign(store, items);
        }),
      },
    },
  };
  return store;
}

describe('getDeviceId', () => {
  it('creates and persists a new id on first call', async () => {
    const store = stubChromeStorage();

    const id = await getDeviceId();

    expect(id).toMatch(/^[0-9a-f-]{36}$/);
    expect(store.charli_device_id).toBe(id);
  });

  it('returns the same id on subsequent calls, without creating a new one', async () => {
    stubChromeStorage({ charli_device_id: 'existing-id' });

    const id = await getDeviceId();

    expect(id).toBe('existing-id');
  });

  it('returns an empty string when the extension APIs are unavailable', async () => {
    expect(await getDeviceId()).toBe('');
  });
});
