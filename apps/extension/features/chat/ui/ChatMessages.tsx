import type { ChatMessage } from '@charli/shared';
import { cn } from '@shared/lib';

interface ChatMessagesProps {
  messages: ChatMessage[];
}

// Presentational: renders the transcript (handles the empty state).
export function ChatMessages({ messages }: ChatMessagesProps) {
  if (messages.length === 0) {
    return (
      <p className="mt-8 text-center text-sm text-slate-400">
        Ask Charli anything about this page.
      </p>
    );
  }

  return (
    <div className="flex flex-col gap-3">
      {messages.map((m, i) => (
        <div
          key={i}
          className={cn(
            'max-w-[85%] rounded-lg px-3 py-2 text-sm',
            m.role === 'user'
              ? 'self-end bg-blue-600 text-white'
              : 'self-start bg-slate-100 text-slate-800',
          )}
        >
          {m.content}
        </div>
      ))}
    </div>
  );
}
