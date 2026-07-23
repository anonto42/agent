// A deterministic stand-in for the LLM provider (OpenAI-compatible), so the E2E
// runs offline with no API key. The backend points its LLM client here.
//
// - If the request carries page context (L1), the reply echoes it back so the
//   test can prove the page text made it all the way to the model.
// - If the user's message mentions "click" and the conversation doesn't yet
//   contain an observation, the reply proposes a click action (L2/L3), so the
//   test can drive the full propose -> confirm -> execute -> observe loop.
// - Once an observation is present (the backend's L3 loop reported the action
//   executed), reply with a final plain-text answer instead of proposing the
//   same action again — otherwise the loop would repeat forever, since this
//   mock always looks at the same original user message, not just the latest
//   turn.
import { createServer } from 'node:http';

const PORT = 8099;
const REPLY = 'Charli here — this is a test reply.';
const PAGE_MARKER = 'The user is viewing this page:';
const OBSERVATION_MARKER = 'Observation:';

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
        const hasObservation = messages.some(
          (x) => typeof x.content === 'string' && x.content.startsWith(OBSERVATION_MARKER),
        );
        const userMsg = messages.find((x) => x.role === 'user');

        if (userMsg && /click/i.test(userMsg.content) && !hasObservation) {
          content = JSON.stringify({
            action: { kind: 'click', target: 'Submit' },
            message: 'Click the Submit button?',
          });
        } else if (hasObservation) {
          content = 'All done — I clicked the Submit button for you.';
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
