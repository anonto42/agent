import type { ChatEvent } from '@charli/shared';
import { charliStream } from '@shared/api/stream';
import { readActivePageText } from '@shared/lib';

let counter = 0;

/** Generates a fresh id used to correlate a sent message with its events. */
export function nextId(): string {
  return `m${Date.now()}-${counter++}`;
}

/** Sends a user message, along with the active tab's text (L1 perception). */
export async function sendChat(id: string, content: string): Promise<void> {
  const page = await readActivePageText();
  await charliStream.post(id, content, page);
}

/** Approves or rejects a previously proposed action (L2). */
export async function confirmAction(id: string, approved: boolean): Promise<void> {
  await charliStream.confirm(id, approved);
}

/** Reports whether an executed action succeeded, continuing the agent loop (L3). */
export async function observeAction(id: string, success: boolean, detail = ''): Promise<void> {
  await charliStream.observe(id, success, detail);
}

/** Stops an in-progress multi-step task (L3 kill switch). */
export async function interruptTask(id: string): Promise<void> {
  await charliStream.interrupt(id);
}

/** Subscribes to every event on this session's stream. Returns an unsubscribe fn. */
export function onChatEvent(listener: (event: ChatEvent) => void): () => void {
  return charliStream.onEvent(listener);
}
