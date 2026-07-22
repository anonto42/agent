// Every action Charli takes is recorded — this feature displays that log.
export interface AuditEntry {
  id: string;
  userId: string;
  tool: string; // the tool/skill the agent invoked
  args: Record<string, unknown>;
  result: 'ok' | 'denied' | 'error';
  site: string; // host the action ran on
  createdAt: string;
}
