import { useCallback, useEffect, useRef, useState } from 'react';
import type { Action, ChatEvent, ChatMessage } from '@charli/shared';
import { performAction } from '@shared/lib';
import { sendChat, confirmAction, onChatEvent, nextId } from '../api';

export interface PendingAction {
  id: string;
  message: string;
  action: Action;
}

const SEND_TIMEOUT_MS = 30_000;

// Owns all chat state, effects, and handlers — including the L2
// propose -> confirm -> execute lifecycle. The UI just renders what this
// returns; no business logic lives in the components.
export function useChat() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [sending, setSending] = useState(false);
  const [confirming, setConfirming] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pendingAction, setPendingAction] = useState<PendingAction | null>(null);

  // The id we're currently waiting on a first response for (chat/error/action).
  const awaitingID = useRef<string | null>(null);

  useEffect(() => {
    return onChatEvent((event: ChatEvent) => {
      if (event.id === awaitingID.current) {
        awaitingID.current = null;
        setSending(false);
      }

      switch (event.type) {
        case 'chat':
          setMessages((m) => [...m, { role: 'assistant', content: event.content }]);
          break;
        case 'error':
          setError(event.content || 'Charli could not answer right now.');
          break;
        case 'action':
          setMessages((m) => [...m, { role: 'assistant', content: event.content }]);
          if (event.action) {
            setPendingAction({ id: event.id, message: event.content, action: event.action });
          }
          break;
        case 'execute':
          setConfirming(false);
          setPendingAction(null);
          if (event.action) {
            const action = event.action;
            void performAction(action).then((ok) => {
              setMessages((m) => [
                ...m,
                {
                  role: 'assistant',
                  content: ok ? '✓ Done.' : '⚠️ I could not complete that action on the page.',
                },
              ]);
            });
          }
          break;
        case 'cancelled':
          setConfirming(false);
          setPendingAction(null);
          setMessages((m) => [...m, { role: 'assistant', content: event.content || 'Cancelled.' }]);
          break;
      }
    });
  }, []);

  const send = useCallback(async () => {
    const text = input.trim();
    if (!text || sending) return;

    const id = nextId();
    setMessages((m) => [...m, { role: 'user', content: text }]);
    setInput('');
    setSending(true);
    setError(null);
    awaitingID.current = id;

    setTimeout(() => {
      if (awaitingID.current === id) {
        awaitingID.current = null;
        setSending(false);
        setError('Charli did not respond in time.');
      }
    }, SEND_TIMEOUT_MS);

    try {
      await sendChat(id, text);
    } catch (e) {
      awaitingID.current = null;
      setSending(false);
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

  return { messages, input, setInput, sending, error, send, pendingAction, confirming, respondToAction };
}
