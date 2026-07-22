import type { ChatMessage } from '@charli/shared';
import { charliSocket } from '@shared/api/socket';

let counter = 0;

// Sends a user message to Charli over the websocket and resolves the reply.
export async function sendChat(history: ChatMessage[], text: string): Promise<string> {
  void history; // full history goes to the agent loop in a later phase
  const id = `m${Date.now()}-${counter++}`;
  const reply = await charliSocket.send({ type: 'chat', id, content: text });
  return reply.content;
}
