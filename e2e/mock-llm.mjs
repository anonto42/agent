// A deterministic stand-in for the LLM provider (OpenAI-compatible), so the E2E
// runs offline with no API key. The backend points its LLM client here.
import { createServer } from 'node:http';

const PORT = 8099;
const REPLY = 'Charli here — this is a test reply.';

createServer((req, res) => {
  if (req.method === 'POST' && req.url === '/v1/chat/completions') {
    let body = '';
    req.on('data', (c) => (body += c));
    req.on('end', () => {
      res.setHeader('Content-Type', 'application/json');
      res.end(
        JSON.stringify({
          choices: [{ message: { role: 'assistant', content: REPLY } }],
        }),
      );
    });
    return;
  }
  res.statusCode = 404;
  res.end();
}).listen(PORT, () => console.log(`mock-llm listening on ${PORT}`));
