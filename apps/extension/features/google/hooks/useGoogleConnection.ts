import { useCallback, useEffect, useRef, useState } from 'react';
import { beginGoogleConnect, getGoogleConnectionStatus } from '../api';

const POLL_INTERVAL_MS = 2_000;
const POLL_TIMEOUT_MS = 2 * 60_000;

// Owns the Google connection status and the connect flow: open the consent
// URL in a new tab, then poll status until it flips to connected (or give up
// after POLL_TIMEOUT_MS). The OAuth round-trip itself happens entirely in
// that other tab; this hook never sees the redirect.
export function useGoogleConnection() {
  const [connected, setConnected] = useState(false);
  const [connecting, setConnecting] = useState(false);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const stopPolling = () => {
    if (pollRef.current) {
      clearInterval(pollRef.current);
      pollRef.current = null;
    }
  };

  useEffect(() => {
    void getGoogleConnectionStatus().then(setConnected);
    return stopPolling;
  }, []);

  const connect = useCallback(async () => {
    if (connected || connecting) return;
    setConnecting(true);
    try {
      const authUrl = await beginGoogleConnect();
      if (typeof chrome !== 'undefined' && chrome.tabs) {
        void chrome.tabs.create({ url: authUrl });
      }
    } catch {
      setConnecting(false);
      return;
    }

    const deadline = Date.now() + POLL_TIMEOUT_MS;
    stopPolling();
    pollRef.current = setInterval(() => {
      void getGoogleConnectionStatus().then((isConnected) => {
        if (isConnected) {
          setConnected(true);
          setConnecting(false);
          stopPolling();
        } else if (Date.now() > deadline) {
          setConnecting(false);
          stopPolling();
        }
      });
    }, POLL_INTERVAL_MS);
  }, [connected, connecting]);

  return { connected, connecting, connect };
}
