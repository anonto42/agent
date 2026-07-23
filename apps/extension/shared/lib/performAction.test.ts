import { describe, it, expect, vi, afterEach } from 'vitest';
import { performAction } from './performAction';
import { domFillFirstTextInput, domClickByText } from './domActions';

afterEach(() => {
  (globalThis as { chrome?: unknown }).chrome = undefined;
});

function stubChrome(executeScriptResult: unknown) {
  const executeScript = vi.fn().mockResolvedValue([{ result: executeScriptResult }]);
  (globalThis as { chrome?: unknown }).chrome = {
    tabs: { query: vi.fn().mockResolvedValue([{ id: 42 }]) },
    scripting: { executeScript },
  };
  return executeScript;
}

describe('performAction', () => {
  it('calls executeScript with the fill function and the action value', async () => {
    const executeScript = stubChrome(true);
    const ok = await performAction({ kind: 'fill', value: 'hello' });

    expect(ok).toBe(true);
    expect(executeScript).toHaveBeenCalledWith(
      expect.objectContaining({
        target: { tabId: 42 },
        func: domFillFirstTextInput,
        args: ['hello'],
      }),
    );
  });

  it('calls executeScript with the click function and the action target', async () => {
    const executeScript = stubChrome(true);
    const ok = await performAction({ kind: 'click', target: 'Submit' });

    expect(ok).toBe(true);
    expect(executeScript).toHaveBeenCalledWith(
      expect.objectContaining({ func: domClickByText, args: ['Submit'] }),
    );
  });

  it('returns false when the extension APIs are unavailable', async () => {
    expect(await performAction({ kind: 'click', target: 'x' })).toBe(false);
  });

  it('returns false when the DOM action itself reports failure', async () => {
    stubChrome(false);
    expect(await performAction({ kind: 'fill', value: 'x' })).toBe(false);
  });
});
