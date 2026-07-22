// Domain types for the authenticated user/session (entities layer: no UI, no API calls).
export interface AuthUser {
  id: string;
  email: string;
  name: string;
  permissions: string[]; // e.g. "site.enable", "audit.read"
}

export interface Session {
  user: AuthUser;
  accessToken: string;
}
