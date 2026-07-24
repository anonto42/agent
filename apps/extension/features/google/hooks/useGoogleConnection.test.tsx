import { renderHook, act, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';

vi.mock('../api', () => ({
  beginGoogleConnect: vi.fn(),
  getGoogleConnectionStatus: vi.fn(),
}));

import { beginGoogleConnect, getGoogleConnectionStatus } from '../api';
import { useGoogleConnection } from './useGoogleConnection';

const mockedBegin = vi.mocked(beginGoogleConnect);
const mockedStatus = vi.mocked(getGoogleConnectionStatus);

describe('useGoogleConnection', () => {
  beforeEach(() => {
    mockedBegin.mockReset();
    mockedStatus.mockReset().mockResolvedValue(false);
  });

  it('checks connection status on mount', async () => {
    mockedStatus.mockResolvedValue(true);
    const { result } = renderHook(() => useGoogleConnection());

    await waitFor(() => expect(result.current.connected).toBe(true));
  });

  it('connect() begins the OAuth flow and polls until connected', async () => {
    mockedBegin.mockResolvedValue('https://accounts.google.com/o/oauth2/auth?state=abc');
    mockedStatus.mockResolvedValueOnce(false).mockResolvedValue(true);

    const { result } = renderHook(() => useGoogleConnection());
    await waitFor(() => expect(result.current.connected).toBe(false));

    await act(async () => {
      await result.current.connect();
    });
    expect(mockedBegin).toHaveBeenCalledTimes(1);
    expect(result.current.connecting).toBe(true);

    await waitFor(() => expect(result.current.connected).toBe(true), { timeout: 5000 });
    expect(result.current.connecting).toBe(false);
  });

  it('does not start a new connect attempt while already connecting', async () => {
    mockedBegin.mockResolvedValue('https://accounts.google.com/o/oauth2/auth?state=abc');

    const { result } = renderHook(() => useGoogleConnection());
    await waitFor(() => expect(result.current.connected).toBe(false));

    await act(async () => {
      await result.current.connect();
    });
    await act(async () => {
      await result.current.connect();
    });

    expect(mockedBegin).toHaveBeenCalledTimes(1);
  });
});
