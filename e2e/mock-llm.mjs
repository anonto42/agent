// A deterministic stand-in for the LLM provider (OpenAI-compatible), so the E2E
// runs offline with no API key. The backend points its LLM client here.
//
// If the request carries page context (L1), the reply echoes it back so the
// test can prove the page text made it all the way to the model.
import { createServer } from 'node:http';

const PORT = 8099;
const REPLY = 'Charli here — this is a test reply.';
const PAGE_MARKER = 'The user is viewing this page:';

createServer((req, res) => {
  if (req.method === 'POST' && req.url === '/v1/chat/completions') {
    let body = '';
    req.on('data', (c) => (body += c));
    req.on('end', () => {
      let seen = '';
      try {
        const { messages = [] } = JSON.parse(body);
        const m = messages.find(
          (x) => typeof x.content === 'string' && x.content.includes(PAGE_MARKER),
        );
        if (m) seen = m.content.slice(m.content.indexOf(PAGE_MARKER) + PAGE_MARKER.length).trim();
      } catch {
        // ignore malformed bodies
      }
      const content = seen ? `PAGE_SEEN: ${seen}` : REPLY;
      res.setHeader('Content-Type', 'application/json');
      res.end(JSON.stringify({ choices: [{ message: { role: 'assistant', content } }] }));
    });
    return;
  }
  res.statusCode = 404;
  res.end();
}).listen(PORT, () => console.log(`mock-llm listening on ${PORT}`));
