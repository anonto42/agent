import { Send, Square } from 'lucide-react';

interface ChatComposerProps {
  value: string;
  disabled: boolean;
  working: boolean;
  onChange: (value: string) => void;
  onSend: () => void;
  onStop: () => void;
}

// Presentational: input + send/stop button. Owns no state; calls back on
// interaction. While a task is working (L3 may take several turns), the send
// button is replaced by a stop button — the kill switch is also reachable via
// Escape (wired in ChatApp).
export function ChatComposer({ value, disabled, working, onChange, onSend, onStop }: ChatComposerProps) {
  return (
    <div className="flex items-center gap-2 rounded-2xl border border-border bg-card py-1.5 pr-1.5 pl-3 shadow-sm transition-all focus-within:border-transparent focus-within:ring-2 focus-within:ring-primary">
      <input
        className="flex-1 bg-transparent text-sm text-foreground placeholder:text-muted-foreground focus:outline-none"
        value={value}
        placeholder="Ask Charli…"
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={(e) => e.key === 'Enter' && onSend()}
      />
      {working ? (
        <button
          className="flex size-8 shrink-0 items-center justify-center rounded-xl bg-destructive text-destructive-foreground transition-colors hover:bg-destructive/90"
          onClick={onStop}
          title="Stop (Esc)"
        >
          <Square className="size-4" />
        </button>
      ) : (
        <button
          className="flex size-8 shrink-0 items-center justify-center rounded-xl bg-primary text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-40"
          onClick={onSend}
          disabled={disabled}
          title="Send"
        >
          <Send className="size-4" />
        </button>
      )}
    </div>
  );
}
