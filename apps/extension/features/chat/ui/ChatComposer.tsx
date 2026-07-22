import { Send } from 'lucide-react';

interface ChatComposerProps {
  value: string;
  disabled: boolean;
  onChange: (value: string) => void;
  onSend: () => void;
}

// Presentational: input + send button. Owns no state; calls back on interaction.
export function ChatComposer({ value, disabled, onChange, onSend }: ChatComposerProps) {
  return (
    <div className="flex items-center gap-2">
      <input
        className="flex-1 rounded-lg border border-slate-200 px-3 py-2 text-sm placeholder:text-slate-400 focus:border-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
        value={value}
        placeholder="Ask Charli…"
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={(e) => e.key === 'Enter' && onSend()}
      />
      <button
        className="flex size-9 items-center justify-center rounded-lg bg-blue-600 text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
        onClick={onSend}
        disabled={disabled}
        title="Send"
      >
        <Send className="size-4" />
      </button>
    </div>
  );
}
