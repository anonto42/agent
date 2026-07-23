import type { ChatMessage } from '@charli/shared';
import { charliStream } from '@shared/api/stream';
import { readActivePageText } from '@shared/lib';

let counter = 0;

// Sends a user message to Charli, along with the text of the page the user is
// currently viewing (L1 perception), and resolves the reply from the SSE stream.
export async function sendChat(history: ChatMessage[], text: string): Promise<string> {
  void history; // full history goes to the agent loop in a later phase
  const page = await readActivePageText();
  const id = `m${Date.now()}-${counter++}`;
  const reply = await charliStream.send({ type: 'chat', id, content: text }, page);
  return reply.content;
}
