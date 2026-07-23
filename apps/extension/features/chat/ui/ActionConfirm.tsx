import type { Action } from '@charli/shared';

interface ActionConfirmProps {
  action: Action;
  disabled: boolean;
  onApprove: () => void;
  onReject: () => void;
}

// Presentational: describes the proposed action and asks the user to decide.
// Owns no state; calls back on interaction. Nothing here executes anything —
// approving only tells the backend to re-check safety and reply with "execute".
export function ActionConfirm({ action, disabled, onApprove, onReject }: ActionConfirmProps) {
  const label =
    action.kind === 'fill'
      ? `Fill "${action.target ?? 'a field'}" with "${action.value ?? ''}"`
      : `Click "${action.target ?? 'a button'}"`;

  return (
    <div className="rounded-lg border border-amber-200 bg-amber-50 p-3 text-sm">
      <p className="mb-2 text-slate-700">{label}</p>
      <div className="flex gap-2">
        <button
          className="rounded-md bg-blue-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
          onClick={onApprove}
          disabled={disabled}
        >
          Approve
        </button>
        <button
          className="rounded-md border border-slate-200 px-3 py-1.5 text-xs font-medium text-slate-700 transition-colors hover:bg-slate-50 disabled:opacity-50"
          onClick={onReject}
          disabled={disabled}
        >
          Reject
        </button>
      </div>
    </div>
  );
}
