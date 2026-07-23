import { useCallback, useEffect, useRef, useState } from 'react';
import type { Action, ChatEvent, ChatMessage } from '@charli/shared';
import { performAction } from '@shared/lib';
import { sendChat, confirmAction, observeAction, interruptTask, onChatEvent, nextId } from '../api';

export interface PendingAction {
  id: string;
  message: string;
  action: Action;
}

const SEND_TIMEOUT_MS = 30_000;

// Owns all chat state, effects, and handlers — including the L2/L3
// propose -> confirm -> execute -> observe -> (repeat | finish) lifecycle.
// A task can take several turns (each a propose/confirm/execute/observe
// round), so `sending` stays true for the whole task, not just its first
// reply. The UI just renders what this returns; no business logic lives in
// the components.
export function useChat() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [sending, setSending] = useState(false);
  const [confirming, setConfirming] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pendingAction, setPendingAction] = useState<PendingAction | null>(null);
  const [taskID, setTaskID] = useState<string | null>(null);

  // The id of the task we're currently running (shared by every turn of it).
  const awaitingID = useRef<string | null>(null);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearWatchdog = () => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
  };

  // (Re)arms the "Charli did not respond in time" watchdog for id. Called at
  // send() and again on every non-terminal event, so a long multi-turn task
  // gets a fresh window per turn instead of one fixed deadline for the whole
  // task.
  const armWatchdog = (id: string) => {
    clearWatchdog();
    timeoutRef.current = setTimeout(() => {
      if (awaitingID.current === id) {
        finishTask();
        setError('Charli did not respond in time.');
      }
    }, SEND_TIMEOUT_MS);
  };

  const finishTask = () => {
    awaitingID.current = null;
    clearWatchdog();
    setSending(false);
    setTaskID(null);
  };

  useEffect(() => {
    return onChatEvent((event: ChatEvent) => {
      const isCurrentTask = event.id === awaitingID.current;

      switch (event.type) {
        case 'chat':
          if (isCurrentTask) finishTask();
          setMessages((m) => [...m, { role: 'assistant', content: event.content }]);
          break;
        case 'error':
          if (isCurrentTask) finishTask();
          setError(event.content || 'Charli could not answer right now.');
          break;
        case 'action':
          if (isCurrentTask) armWatchdog(event.id);
          setTaskID(event.id);
          setMessages((m) => [...m, { role: 'assistant', content: event.content }]);
          if (event.action) {
            setPendingAction({ id: event.id, message: event.content, action: event.action });
          }
          break;
        case 'execute':
          setConfirming(false);
          setPendingAction(null);
          if (isCurrentTask) armWatchdog(event.id); // give the next turn (observe -> step) a fresh window
          if (event.action) {
            const action = event.action;
            const id = event.id;
            void performAction(action).then((ok) => {
              setMessages((m) => [
                ...m,
                {
                  role: 'assistant',
                  content: ok ? '✓ Done.' : '⚠️ I could not complete that action on the page.',
                },
              ]);
              void observeAction(id, ok, ok ? '' : 'the DOM action reported failure').catch(() => {
                if (isCurrentTask) finishTask();
                setError('Lost contact with Charli mid-task.');
              });
            });
          }
          break;
        case 'cancelled':
          if (isCurrentTask) finishTask();
          setConfirming(false);
          setPendingAction(null);
          setMessages((m) => [...m, { role: 'assistant', content: event.content || 'Cancelled.' }]);
          break;
        case 'interrupted':
          if (isCurrentTask) finishTask();
          setConfirming(false);
          setPendingAction(null);
          setMessages((m) => [...m, { role: 'assistant', content: event.content || 'Stopped.' }]);
          break;
      }
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const send = useCallback(async () => {
    const text = input.trim();
    if (!text || sending) return;

    const id = nextId();
    setMessages((m) => [...m, { role: 'user', content: text }]);
    setInput('');
    setSending(true);
    setError(null);
    setTaskID(id);
    awaitingID.current = id;
    armWatchdog(id);

    try {
      await sendChat(id, text);
    } catch (e) {
      finishTask();
      setError(e instanceof Error ? e.message : 'Failed to reach Charli');
    }
  }, [input, sending]);

  const respondToAction = useCallback(
    async (approved: boolean) => {
      if (!pendingAction || confirming) return;
      setConfirming(true);
      try {
        await confirmAction(pendingAction.id, approved);
      } catch (e) {
        setConfirming(false);
        setError(e instanceof Error ? e.message : 'Failed to reach Charli');
      }
    },
    [pendingAction, confirming],
  );

  // The L3 kill switch: stops the in-progress task, best-effort. If the
  // request itself fails, the task still ends eventually via its own turn
  // limit or watchdog.
  const stop = useCallback(async () => {
    if (!taskID) return;
    try {
      await interruptTask(taskID);
    } catch {
      // best-effort
    }
  }, [taskID]);

  return {
    messages,
    input,
    setInput,
    sending,
    error,
    send,
    pendingAction,
    confirming,
    respondToAction,
    stop,
  };
}
