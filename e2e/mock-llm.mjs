// A deterministic stand-in for the LLM provider (OpenAI-compatible), so the E2E
// runs offline with no API key. The backend points its LLM client here.
//
// - If the request carries page context (L1), the reply echoes it back so the
//   test can prove the page text made it all the way to the model.
// - If the user's message mentions "click", the reply proposes a click action
//   (L2), so the test can drive the full propose -> confirm -> execute loop.
import { createServer } from 'node:http';

const PORT = 8099;
const REPLY = 'Charli here — this is a test reply.';
const PAGE_MARKER = 'The user is viewing this page:';

createServer((req, res) => {
  if (req.method === 'POST' && req.url === '/v1/chat/completions') {
    let body = '';
    req.on('data', (c) => (body += c));
    req.on('end', () => {
      let content = REPLY;
      try {
        const { messages = [] } = JSON.parse(body);
        const pageMsg = messages.find(
          (x) => typeof x.content === 'string' && x.content.includes(PAGE_MARKER),
        );
        const userMsg = messages.find((x) => x.role === 'user');

        if (userMsg && /click/i.test(userMsg.content)) {
          content = JSON.stringify({
            action: { kind: 'click', target: 'Submit' },
            message: 'Click the Submit button?',
          });
        } else if (pageMsg) {
          const seen = pageMsg.content.slice(pageMsg.content.indexOf(PAGE_MARKER) + PAGE_MARKER.length).trim();
          content = `PAGE_SEEN: ${seen}`;
        }
      } catch {
        // ignore malformed bodies, fall back to the default reply
      }
      res.setHeader('Content-Type', 'application/json');
      res.end(JSON.stringify({ choices: [{ message: { role: 'assistant', content } }] }));
    });
    return;
  }
  res.statusCode = 404;
  res.end();
}).listen(PORT, () => console.log(`mock-llm listening on ${PORT}`));
