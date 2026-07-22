import type { ChatMessage } from '@charli/shared';
import { charliStream } from '@shared/api/stream';

let counter = 0;

// Sends a user message to Charli (POST /chat) and resolves the reply that
// arrives back over the SSE stream.
export async function sendChat(history: ChatMessage[], text: string): Promise<string> {
  void history; // full history goes to the agent loop in a later phase
  const id = `m${Date.now()}-${counter++}`;
  const reply = await charliStream.send({ type: 'chat', id, content: text });
  return reply.content;
}
