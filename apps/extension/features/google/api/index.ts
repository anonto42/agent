import type { GoogleConnectResponse, GoogleStatusResponse } from '@charli/shared';
import { getDeviceId } from '@shared/lib';

const BASE = 'http://localhost:8080/api/v1/integrations/google';

interface Envelope<T> {
  data?: T;
}

/** Starts an OAuth connection for this browser installation, returning the
 * Google consent URL to open in a new tab. */
export async function beginGoogleConnect(): Promise<string> {
  const deviceId = await getDeviceId();
  const res = await fetch(`${BASE}/connect`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deviceId }),
  });
  if (!res.ok) throw new Error(`connect failed (${res.status})`);
  const body = (await res.json()) as Envelope<GoogleConnectResponse>;
  if (!body.data?.authUrl) throw new Error('no auth URL in response');
  return body.data.authUrl;
}

/** Reports whether this browser installation has a completed Google connection. */
export async function getGoogleConnectionStatus(): Promise<boolean> {
  const deviceId = await getDeviceId();
  if (!deviceId) return false;
  const res = await fetch(`${BASE}/status?deviceId=${encodeURIComponent(deviceId)}`);
  if (!res.ok) return false;
  const body = (await res.json()) as Envelope<GoogleStatusResponse>;
  return body.data?.connected ?? false;
}
