import { describe, it, expect, vi, afterEach } from 'vitest';
import { performAction } from './performAction';
import { domFillFirstTextInput, domClickByText } from './domActions';

afterEach(() => {
  (globalThis as { chrome?: unknown }).chrome = undefined;
  vi.unstubAllGlobals();
});

function stubChrome(executeScriptResult: unknown) {
  const executeScript = vi.fn().mockResolvedValue([{ result: executeScriptResult }]);
  (globalThis as { chrome?: unknown }).chrome = {
    tabs: { query: vi.fn().mockResolvedValue([{ id: 42 }]) },
    scripting: { executeScript },
    storage: { local: { get: vi.fn().mockResolvedValue({}), set: vi.fn().mockResolvedValue(undefined) } },
  };
  return executeScript;
}

describe('performAction', () => {
  it('calls executeScript with the fill function and the action value', async () => {
    const executeScript = stubChrome(true);
    const result = await performAction({ kind: 'fill', value: 'hello' });

    expect(result).toEqual({ success: true });
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
    const result = await performAction({ kind: 'click', target: 'Submit' });

    expect(result).toEqual({ success: true });
    expect(executeScript).toHaveBeenCalledWith(
      expect.objectContaining({ func: domClickByText, args: ['Submit'] }),
    );
  });

  it('reports failure when the extension APIs are unavailable', async () => {
    expect(await performAction({ kind: 'click', target: 'x' })).toEqual({ success: false });
  });

  it('reports failure when the DOM action itself reports failure', async () => {
    stubChrome(false);
    expect(await performAction({ kind: 'fill', value: 'x' })).toEqual({ success: false });
  });

  it('sheets_append calls the backend append endpoint with the device id, and reports the result', async () => {
    stubChrome(true);
    const fetchMock = vi.fn().mockResolvedValue({
      json: () => Promise.resolve({ data: { success: true } }),
    });
    vi.stubGlobal('fetch', fetchMock);

    const result = await performAction({ kind: 'sheets_append', spreadsheetId: 'abc123', values: ['a', 'b'] });

    expect(result).toEqual({ success: true, detail: undefined });
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toBe('http://localhost:8080/api/v1/integrations/google/append');
    expect(init.method).toBe('POST');
    const body = JSON.parse(init.body as string);
    expect(body).toMatchObject({ spreadsheetId: 'abc123', values: ['a', 'b'] });
    expect(body.deviceId).toMatch(/^[0-9a-f-]{36}$/);
  });

  it('sheets_append surfaces a failure detail from the backend', async () => {
    stubChrome(true);
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        json: () => Promise.resolve({ data: { success: false, detail: "google sheets isn't connected yet" } }),
      }),
    );

    const result = await performAction({ kind: 'sheets_append', spreadsheetId: 'abc123', values: ['a'] });

    expect(result).toEqual({ success: false, detail: "google sheets isn't connected yet" });
  });
});
