import { useCallback, useState } from 'react';
import type { ChatMessage } from '@charli/shared';
import { sendChat } from '../api';

// Owns all chat state, effects, and handlers. The UI just renders what this
// returns — no business logic lives in the components.
export function useChat() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [sending, setSending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const send = useCallback(async () => {
    const text = input.trim();
    if (!text || sending) return;

    const next: ChatMessage[] = [...messages, { role: 'user', content: text }];
    setMessages(next);
    setInput('');
    setSending(true);
    setError(null);
    try {
      const reply = await sendChat(next, text);
      setMessages((m) => [...m, { role: 'assistant', content: reply }]);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to reach Charli');
    } finally {
      setSending(false);
    }
  }, [input, sending, messages]);

  return { messages, input, setInput, sending, error, send };
}
