import type { ChatMessage } from '@charli/shared';
import { Bot, MessageCircle, User } from 'lucide-react';
import { cn } from '@shared/lib';

interface ChatMessagesProps {
  messages: ChatMessage[];
}

// Presentational: renders the transcript (handles the empty state).
export function ChatMessages({ messages }: ChatMessagesProps) {
  if (messages.length === 0) {
    return (
      <div className="mt-8 flex flex-col items-center gap-3 text-center">
        <span className="flex size-12 items-center justify-center rounded-full bg-muted">
          <MessageCircle className="size-6 text-muted-foreground" />
        </span>
        <p className="text-sm text-muted-foreground">Ask Charli anything about this page.</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-1">
      {messages.map((m, i) => {
        const isUser = m.role === 'user';
        return (
          <div key={i} className={cn('my-1.5 flex items-start gap-2', isUser && 'flex-row-reverse')}>
            <span
              className={cn(
                'flex size-7 shrink-0 items-center justify-center rounded-full',
                isUser ? 'bg-primary' : 'bg-accent',
              )}
            >
              {isUser ? (
                <User className="size-3.5 text-primary-foreground" />
              ) : (
                <Bot className="size-3.5 text-accent-foreground" />
              )}
            </span>
            <div
              className={cn(
                'max-w-[80%] rounded-lg px-3 py-2 text-sm shadow-sm',
                isUser
                  ? 'bg-primary text-primary-foreground'
                  : 'border border-border bg-card text-foreground',
              )}
            >
              {m.content}
            </div>
          </div>
        );
      })}
    </div>
  );
}
