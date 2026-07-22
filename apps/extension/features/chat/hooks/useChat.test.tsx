import { renderHook, act, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';

// Factory mock so the real api (and its stream client) never loads.
vi.mock('../api', () => ({ sendChat: vi.fn() }));
import { sendChat } from '../api';
import { useChat } from './useChat';

const mockedSendChat = vi.mocked(sendChat);

describe('useChat', () => {
  beforeEach(() => mockedSendChat.mockReset());

  it('adds the user message then the assistant reply', async () => {
    mockedSendChat.mockResolvedValue('hi from charli');
    const { result } = renderHook(() => useChat());

    act(() => result.current.setInput('hello'));
    await act(async () => {
      await result.current.send();
    });

    await waitFor(() => {
      expect(result.current.messages).toEqual([
        { role: 'user', content: 'hello' },
        { role: 'assistant', content: 'hi from charli' },
      ]);
    });
    expect(result.current.error).toBeNull();
  });

  it('surfaces an error when the api fails, keeping the user message', async () => {
    mockedSendChat.mockRejectedValueOnce(new Error('boom'));
    const { result } = renderHook(() => useChat());

    act(() => result.current.setInput('hi'));
    await act(async () => {
      // send() handles the error internally; guard the outer await regardless.
      await result.current.send().catch(() => {});
    });

    await waitFor(() => expect(result.current.error).toBe('boom'));
    expect(result.current.messages).toEqual([{ role: 'user', content: 'hi' }]);
  });

  it('ignores empty/whitespace input', async () => {
    const { result } = renderHook(() => useChat());
    act(() => result.current.setInput('   '));
    await act(async () => {
      await result.current.send();
    });
    expect(result.current.messages).toEqual([]);
    expect(mockedSendChat).not.toHaveBeenCalled();
  });
});
