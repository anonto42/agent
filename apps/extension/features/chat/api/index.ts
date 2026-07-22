import type { ChatMessage } from '@charli/shared';

// Sends a user message to Charli and resolves the assistant's reply.
//
// TODO(phase-0): route through the background worker -> websocket -> Go backend.
// For now it echoes locally so the UI is exercisable before the transport exists.
export async function sendChat(history: ChatMessage[], text: string): Promise<string> {
  void history;
  await new Promise((resolve) => setTimeout(resolve, 150));
  return `You said: ${text}`;
}
