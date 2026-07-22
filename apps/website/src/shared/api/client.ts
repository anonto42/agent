// Thin client for the Go backend. The access token is kept in memory only —
// never in localStorage (see .agents/shared/scopes/website.md).
let accessToken: string | null = null;

export function setAccessToken(token: string | null): void {
  accessToken = token;
}

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/api/v1';

/** Standard backend response envelope (mirrors pkg/response in the Go backend). */
export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data?: T;
  error?: string;
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<ApiResponse<T>> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
      ...init?.headers,
    },
  });
  return (await res.json()) as ApiResponse<T>;
}
