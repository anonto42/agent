// Realtime client for the Go backend: an SSE stream for messages coming DOWN
// (server -> client) and POST /chat for messages going UP (client -> server),
// correlated by message id.
//
// Phase 0 note: the side panel connects directly. When content scripts also
// need the backend (L1+), this moves into the background service worker and the
// panel talks to it via runtime messaging.
export interface WireMessage {
  type: string;
  id: string;
  content: string;
}

const BASE = 'http://localhost:8080/api/v1';
const TIMEOUT_MS = 10_000;

class CharliStream {
  private es: EventSource | null = null;
  private ready: Promise<void> | null = null;
  private readonly session = crypto.randomUUID();
  private pending = new Map<string, (msg: WireMessage) => void>();

  // Opens the SSE stream once and resolves when it is connected (so the
  // server-side subscription exists before we POST).
  private connect(): Promise<void> {
    if (this.ready && this.es && this.es.readyState !== EventSource.CLOSED) {
      return this.ready;
    }
    this.ready = new Promise<void>((resolve) => {
      const es = new EventSource(`${BASE}/events?session=${this.session}`);
      es.onopen = () => resolve();
      es.onmessage = (ev) => {
        const msg = JSON.parse(ev.data as string) as WireMessage;
        const resolvePending = this.pending.get(msg.id);
        if (resolvePending) {
          this.pending.delete(msg.id);
          resolvePending(msg);
        }
      };
      // EventSource reconnects automatically on error; nothing to do here.
      this.es = es;
    });
    return this.ready;
  }

  async send(message: WireMessage): Promise<WireMessage> {
    await this.connect();

    return new Promise<WireMessage>((resolve, reject) => {
      this.pending.set(message.id, resolve);

      const timer = setTimeout(() => {
        if (this.pending.delete(message.id)) reject(new Error('Charli did not respond'));
      }, TIMEOUT_MS);

      fetch(`${BASE}/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session: this.session,
          id: message.id,
          content: message.content,
        }),
      })
        .then((res) => {
          if (!res.ok) throw new Error(`chat failed (${res.status})`);
        })
        .catch((err) => {
          clearTimeout(timer);
          this.pending.delete(message.id);
          reject(err instanceof Error ? err : new Error('send failed'));
        });
    });
  }
}

export const charliStream = new CharliStream();
