import { useGoogleConnection } from '../hooks/useGoogleConnection';

// Presentational-ish: a small header affordance showing/controlling the
// Google Sheets connection (L4). No settings page, no spreadsheet picker —
// the user identifies the sheet by URL/ID in chat, same as any instruction.
export function GoogleConnectButton() {
  const { connected, connecting, connect } = useGoogleConnection();

  if (connected) {
    return (
      <span className="ml-auto rounded-full bg-secondary/10 px-2 py-1 text-xs font-medium text-secondary">
        Sheets ✓
      </span>
    );
  }

  return (
    <button
      className="ml-auto rounded-full border border-border px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-muted disabled:opacity-50"
      onClick={() => void connect()}
      disabled={connecting}
      title="Connect Google Sheets"
    >
      {connecting ? 'Connecting…' : 'Connect Sheets'}
    </button>
  );
}
