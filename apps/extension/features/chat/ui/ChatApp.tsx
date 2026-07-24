import { useEffect } from 'react';
import { Sparkles } from 'lucide-react';
import { useChat } from '../hooks/useChat';
import { ChatMessages } from './ChatMessages';
import { ChatComposer } from './ChatComposer';
import { ActionConfirm } from './ActionConfirm';
import { GoogleConnectButton } from '@features/google/ui/GoogleConnectButton';

// Container: delegates state to useChat and composes presentational children.
// (The clean container-hook pattern from omni.dns.)
export function ChatApp() {
  const { messages, input, setInput, sending, error, send, pendingAction, confirming, respondToAction, stop } =
    useChat();

  // The Esc kill switch (agent-safety.md): stops an in-progress task from
  // anywhere in the panel, not just via the Stop button.
  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape' && sending) {
        void stop();
      }
    }
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [sending, stop]);

  return (
    <div className="flex h-screen flex-col bg-background text-foreground">
      <header className="flex items-center gap-2 border-b border-border bg-card px-4 py-3">
        <span className="flex size-7 items-center justify-center rounded-lg bg-primary text-primary-foreground">
          <Sparkles className="size-4" />
        </span>
        <h1 className="text-sm font-semibold">Charli</h1>
        <GoogleConnectButton />
      </header>

      <div className="flex flex-1 flex-col gap-1 overflow-y-auto px-4 py-3">
        <ChatMessages messages={messages} />
        {pendingAction && (
          <ActionConfirm
            action={pendingAction.action}
            disabled={confirming}
            onApprove={() => respondToAction(true)}
            onReject={() => respondToAction(false)}
          />
        )}
      </div>

      {error && <p className="px-4 pb-2 text-xs text-destructive">{error}</p>}

      <div className="border-t border-border bg-background p-3">
        <ChatComposer
          value={input}
          disabled={sending || Boolean(pendingAction)}
          working={sending}
          onChange={setInput}
          onSend={send}
          onStop={stop}
        />
      </div>
    </div>
  );
}
