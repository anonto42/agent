import { renderHook, act, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import type { ChatEvent } from '@charli/shared';

// Factory mock: control exactly what the "backend" sends back, without any
// real network/SSE involved.
vi.mock('../api', () => ({
  nextId: vi.fn(),
  sendChat: vi.fn(),
  confirmAction: vi.fn(),
  observeAction: vi.fn(),
  interruptTask: vi.fn(),
  onChatEvent: vi.fn(),
}));
vi.mock('@shared/lib', () => ({ performAction: vi.fn() }));

import { nextId, sendChat, confirmAction, observeAction, interruptTask, onChatEvent } from '../api';
import { performAction } from '@shared/lib';
import { useChat } from './useChat';

const mockedNextId = vi.mocked(nextId);
const mockedSendChat = vi.mocked(sendChat);
const mockedConfirmAction = vi.mocked(confirmAction);
const mockedObserveAction = vi.mocked(observeAction);
const mockedInterruptTask = vi.mocked(interruptTask);
const mockedOnChatEvent = vi.mocked(onChatEvent);
const mockedPerformAction = vi.mocked(performAction);

describe('useChat', () => {
  let emit: (event: ChatEvent) => void;

  beforeEach(() => {
    mockedNextId.mockReset().mockReturnValue('id-1');
    mockedSendChat.mockReset().mockResolvedValue(undefined);
    mockedConfirmAction.mockReset().mockResolvedValue(undefined);
    mockedObserveAction.mockReset().mockResolvedValue(undefined);
    mockedInterruptTask.mockReset().mockResolvedValue(undefined);
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

  it('surfaces a proposed action from an "action" event, executes on approve, and reports the outcome back', async () => {
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
    // L3: the extension must tell the backend the action succeeded so the
    // loop can decide its next step.
    await waitFor(() => expect(mockedObserveAction).toHaveBeenCalledWith('id-1', true, ''));
    // The task is still going (waiting on the backend's next turn) — the
    // composer must stay disabled rather than re-enabling between turns.
    expect(result.current.sending).toBe(true);
  });

  it('continues a multi-step task across several turns before finishing', async () => {
    const { result } = renderHook(() => useChat());
    act(() => result.current.setInput('do a two-step task'));
    await act(async () => {
      await result.current.send();
    });

    // Turn 1: propose -> approve -> execute -> observe.
    act(() =>
      emit({ type: 'action', id: 'id-1', content: 'Step 1?', action: { kind: 'click', target: 'Next' } }),
    );
    await act(async () => {
      await result.current.respondToAction(true);
    });
    act(() => emit({ type: 'execute', id: 'id-1', content: '', action: { kind: 'click', target: 'Next' } }));
    await waitFor(() => expect(mockedObserveAction).toHaveBeenCalledTimes(1));
    expect(result.current.sending).toBe(true);

    // Turn 2: the backend proposes another step off the same task id.
    act(() =>
      emit({ type: 'action', id: 'id-1', content: 'Step 2?', action: { kind: 'click', target: 'Finish' } }),
    );
    await waitFor(() => expect(result.current.pendingAction?.message).toBe('Step 2?'));
    expect(result.current.sending).toBe(true);

    // The task ends only once a terminal event arrives.
    act(() => emit({ type: 'chat', id: 'id-1', content: 'All done.' }));
    await waitFor(() => expect(result.current.sending).toBe(false));
    expect(result.current.messages.at(-1)).toEqual({ role: 'assistant', content: 'All done.' });
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

  it('stop() interrupts the in-progress task, and "interrupted" ends it', async () => {
    const { result } = renderHook(() => useChat());
    act(() => result.current.setInput('do something long'));
    await act(async () => {
      await result.current.send();
    });
    expect(result.current.sending).toBe(true);

    await act(async () => {
      await result.current.stop();
    });
    expect(mockedInterruptTask).toHaveBeenCalledWith('id-1');

    act(() => emit({ type: 'interrupted', id: 'id-1', content: 'Stopped.' }));

    await waitFor(() => expect(result.current.sending).toBe(false));
    expect(result.current.messages.at(-1)).toEqual({ role: 'assistant', content: 'Stopped.' });
  });
});
