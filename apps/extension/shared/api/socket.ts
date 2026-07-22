// Minimal websocket client to the Go backend, with request/response
// correlation by message id.
//
// Phase 0 note: the side panel connects directly. When content scripts also
// need the backend (L1+), this moves into the background service worker and the
// panel talks to it via runtime messaging.
export interface WireMessage {
  type: string;
  id: string;
  content: string;
}

const WS_URL = 'ws://localhost:8080/ws';
const TIMEOUT_MS = 10_000;

class CharliSocket {
  private ws: WebSocket | null = null;
  private pending = new Map<string, (msg: WireMessage) => void>();
  private queue: WireMessage[] = [];

  private ensure(): WebSocket {
    const existing = this.ws;
    if (
      existing &&
      (existing.readyState === WebSocket.OPEN || existing.readyState === WebSocket.CONNECTING)
    ) {
      return existing;
    }

    const ws = new WebSocket(WS_URL);
    ws.onopen = () => {
      for (const m of this.queue) ws.send(JSON.stringify(m));
      this.queue = [];
    };
    ws.onmessage = (ev) => {
      const msg = JSON.parse(ev.data as string) as WireMessage;
      const resolve = this.pending.get(msg.id);
      if (resolve) {
        this.pending.delete(msg.id);
        resolve(msg);
      }
    };
    ws.onclose = () => {
      if (this.ws === ws) this.ws = null;
    };
    this.ws = ws;
    return ws;
  }

  send(message: WireMessage): Promise<WireMessage> {
    const ws = this.ensure();
    return new Promise((resolve, reject) => {
      this.pending.set(message.id, resolve);

      if (ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify(message));
      else this.queue.push(message);

      setTimeout(() => {
        if (this.pending.delete(message.id)) reject(new Error('Charli did not respond'));
      }, TIMEOUT_MS);
    });
  }
}

export const charliSocket = new CharliSocket();
