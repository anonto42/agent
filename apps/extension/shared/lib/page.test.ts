import { describe, it, expect, vi, afterEach } from 'vitest';
import { readActivePageText } from './page';

afterEach(() => {
  (globalThis as { chrome?: unknown }).chrome = undefined;
});

describe('readActivePageText', () => {
  it('returns the active tab text, truncated to maxChars', async () => {
    (globalThis as { chrome?: unknown }).chrome = {
      tabs: { query: vi.fn().mockResolvedValue([{ id: 7 }]) },
      scripting: { executeScript: vi.fn().mockResolvedValue([{ result: 'hello page world' }]) },
    };

    const text = await readActivePageText(5);
    expect(text).toBe('hello');
  });

  it('returns empty string when the extension APIs are unavailable', async () => {
    expect(await readActivePageText()).toBe('');
  });
});
