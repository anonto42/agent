import type { ChatEvent, ChatRequest, ConfirmRequest, ObserveRequest, InterruptRequest } from '@charli/shared';

// Realtime client for the Go backend: an SSE stream for messages coming DOWN
// (server -> client) and POST /chat + POST /confirm for messages going UP
// (client -> server). Message shapes come from the shared contract.
//
// One user message can produce MORE THAN ONE event over time (an "action"
// proposal, then later an "execute"/"cancelled" once confirmed) — so this is a
// plain pub/sub, not a one-shot request/response promise. Callers match events
// to their own request by `id`.
//
// Phase 0 note: the side panel connects directly. When content scripts also
// need the backend (L1+), this moves into the background service worker and the
// panel talks to it via runtime messaging.

const BASE = 'http://localhost:8080/api/v1';

class CharliStream {
  private es: EventSource | null = null;
  private ready: Promise<void> | null = null;
  private readonly session = crypto.randomUUID();
  private listeners = new Set<(event: ChatEvent) => void>();

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
        const event = JSON.parse(ev.data as string) as ChatEvent;
        for (const listener of this.listeners) listener(event);
      };
      // EventSource reconnects automatically on error; nothing to do here.
      this.es = es;
    });
    return this.ready;
  }

  /** Subscribes to every event on this session's stream. Returns an unsubscribe fn. */
  onEvent(listener: (event: ChatEvent) => void): () => void {
    this.listeners.add(listener);
    void this.connect();
    return () => this.listeners.delete(listener);
  }

  async post(id: string, content: string, page: string): Promise<void> {
    await this.connect();
    const body: ChatRequest = { session: this.session, id, content, page };
    const res = await fetch(`${BASE}/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (!res.ok) throw new Error(`chat failed (${res.status})`);
  }

  async confirm(id: string, approved: boolean): Promise<void> {
    await this.connect();
    const body: ConfirmRequest = { session: this.session, id, approved };
    const res = await fetch(`${BASE}/confirm`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (!res.ok) throw new Error(`confirm failed (${res.status})`);
  }

  /** Reports whether an executed action succeeded, continuing the agent loop (L3). */
  async observe(id: string, success: boolean, detail: string): Promise<void> {
    await this.connect();
    const body: ObserveRequest = { session: this.session, id, success, detail };
    const res = await fetch(`${BASE}/observe`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (!res.ok) throw new Error(`observe failed (${res.status})`);
  }

  /** Stops an in-progress multi-step task (L3 kill switch). */
  async interrupt(id: string): Promise<void> {
    await this.connect();
    const body: InterruptRequest = { session: this.session, id };
    const res = await fetch(`${BASE}/interrupt`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (!res.ok) throw new Error(`interrupt failed (${res.status})`);
  }
}

export const charliStream = new CharliStream();
