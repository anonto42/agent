import '@testing-library/jest-dom/vitest';
import { webcrypto } from 'node:crypto';

// jsdom's crypto may lack randomUUID (used by the stream client); polyfill it.
if (!globalThis.crypto?.randomUUID) {
  // @ts-expect-error test-env polyfill
  globalThis.crypto = webcrypto;
}
