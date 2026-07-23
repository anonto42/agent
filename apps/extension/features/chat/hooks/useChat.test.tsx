import { renderHook, act, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import type { ChatEvent } from '@charli/shared';

// Factory mock: control exactly what the "backend" sends back, without any
// real network/SSE involved.
vi.mock('../api', () => ({
  nextId: vi.fn(),
  sendChat: vi.fn(),
  confirmAction: vi.fn(),
  onChatEvent: vi.fn(),
}));
vi.mock('@shared/lib', () => ({ performAction: vi.fn() }));

import { nextId, sendChat, confirmAction, onChatEvent } from '../api';
import { performAction } from '@shared/lib';
import { useChat } from './useChat';

const mockedNextId = vi.mocked(nextId);
const mockedSendChat = vi.mocked(sendChat);
const mockedConfirmAction = vi.mocked(confirmAction);
const mockedOnChatEvent = vi.mocked(onChatEvent);
const mockedPerformAction = vi.mocked(performAction);

describe('useChat', () => {
  let emit: (event: ChatEvent) => void;

  beforeEach(() => {
    mockedNextId.mockReset().mockReturnValue('id-1');
    mockedSendChat.mockReset().mockResolvedValue(undefined);
    mockedConfirmAction.mockReset().mockResolvedValue(undefined);
    mockedPerformAction.mockReset().mockResolvedValue(true);
    mockedOnChatEvent.mockReset().mockImplementation((listener) => {
      emit = listener;
      return () => {};
    });
  });

  it('adds the user message, then the assistant reply on a "chat" event', async () => {
    const { result } = renderHook(() => useChat());

    act(() => result.current.setInput('hello'));
    await act(async () => {
      await result.current.send();
    });
    expect(result.current.sending).toBe(true);
    expect(mockedSendChat).toHaveBeenCalledWith('id-1', 'hello');

    act(() => emit({ type: 'chat', id: 'id-1', content: 'hi from charli' }));

    await waitFor(() => expect(result.current.sending).toBe(false));
    expect(result.current.messages).toEqual([
      { role: 'user', content: 'hello' },
      { role: 'assistant', content: 'hi from charli' },
    ]);
  });

  it('surfaces an error event, keeping the user message', async () => {
    const { result } = renderHook(() => useChat());
    act(() => result.current.setInput('hi'));
    await act(async () => {
      await result.current.send();
    });

    act(() => emit({ type: 'error', id: 'id-1', content: 'boom' }));

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

  it('surfaces a proposed action from an "action" event and clears it on approve+execute', async () => {
    const { result } = renderHook(() => useChat());
    act(() => result.current.setInput('click submit'));
    await act(async () => {
      await result.current.send();
    });

    act(() =>
      emit({
        type: 'action',
        id: 'id-1',
        content: 'Click submit?',
        action: { kind: 'click', target: 'Submit' },
      }),
    );

    await waitFor(() => expect(result.current.pendingAction).not.toBeNull());
    expect(result.current.pendingAction).toEqual({
      id: 'id-1',
      message: 'Click submit?',
      action: { kind: 'click', target: 'Submit' },
    });

    await act(async () => {
      await result.current.respondToAction(true);
    });
    expect(mockedConfirmAction).toHaveBeenCalledWith('id-1', true);

    act(() => emit({ type: 'execute', id: 'id-1', content: '', action: { kind: 'click', target: 'Submit' } }));

    await waitFor(() => expect(result.current.pendingAction).toBeNull());
    await waitFor(() => expect(mockedPerformAction).toHaveBeenCalledWith({ kind: 'click', target: 'Submit' }));
    await waitFor(() =>
      expect(result.current.messages.at(-1)).toEqual({ role: 'assistant', content: '✓ Done.' }),
    );
  });

  it('clears the pending action and shows a message on "cancelled"', async () => {
    const { result } = renderHook(() => useChat());
    act(() => result.current.setInput('fill the form'));
    await act(async () => {
      await result.current.send();
    });
    act(() =>
      emit({ type: 'action', id: 'id-1', content: 'Fill it?', action: { kind: 'fill', value: 'x' } }),
    );
    await waitFor(() => expect(result.current.pendingAction).not.toBeNull());

    await act(async () => {
      await result.current.respondToAction(false);
    });
    act(() => emit({ type: 'cancelled', id: 'id-1', content: 'Cancelled.' }));

    await waitFor(() => expect(result.current.pendingAction).toBeNull());
    expect(result.current.messages.at(-1)).toEqual({ role: 'assistant', content: 'Cancelled.' });
  });
});
