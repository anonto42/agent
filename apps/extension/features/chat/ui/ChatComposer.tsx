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
    <div className="flex items-center gap-2">
      <input
        className="flex-1 rounded-lg border border-slate-200 px-3 py-2 text-sm placeholder:text-slate-400 focus:border-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
        value={value}
        placeholder="Ask Charli…"
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={(e) => e.key === 'Enter' && onSend()}
      />
      {working ? (
        <button
          className="flex size-9 items-center justify-center rounded-lg bg-red-600 text-white transition-colors hover:bg-red-700"
          onClick={onStop}
          title="Stop (Esc)"
        >
          <Square className="size-4" />
        </button>
      ) : (
        <button
          className="flex size-9 items-center justify-center rounded-lg bg-blue-600 text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
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
