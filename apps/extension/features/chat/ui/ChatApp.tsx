import { Sparkles } from 'lucide-react';
import { useChat } from '../hooks/useChat';
import { ChatMessages } from './ChatMessages';
import { ChatComposer } from './ChatComposer';

// Container: delegates state to useChat and composes presentational children.
// (The clean container-hook pattern from omni.dns.)
export function ChatApp() {
  const { messages, input, setInput, sending, error, send } = useChat();

  return (
    <div className="flex h-screen flex-col bg-white text-slate-800">
      <header className="flex items-center gap-2 border-b border-slate-200 px-4 py-3">
        <span className="flex size-7 items-center justify-center rounded-lg bg-blue-600 text-white">
          <Sparkles className="size-4" />
        </span>
        <h1 className="text-sm font-semibold">Charli</h1>
      </header>

      <div className="flex-1 overflow-y-auto px-4 py-3">
        <ChatMessages messages={messages} />
      </div>

      {error && <p className="px-4 pb-2 text-xs text-red-500">{error}</p>}

      <div className="border-t border-slate-200 p-3">
        <ChatComposer value={input} disabled={sending} onChange={setInput} onSend={send} />
      </div>
    </div>
  );
}
