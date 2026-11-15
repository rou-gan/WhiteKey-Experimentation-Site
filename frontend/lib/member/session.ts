/** クライアントセッション。sessionStorage のみ（EDS-MEM-3 §6）。 */

export const SESSION_STORAGE_KEY = "whitekey.member.session";

export type ClientSession = {
  sessionToken: string;
  nickname: string;
  expiresAt: string;
};

type SessionStorageLike = {
  getItem(key: string): string | null;
  setItem(key: string, value: string): void;
  removeItem(key: string): void;
};

function storage(): SessionStorageLike | null {
  if (typeof globalThis.sessionStorage === "undefined") {
    return null;
  }
  return globalThis.sessionStorage;
}

export function saveSession(session: ClientSession, store = storage()): void {
  if (!store) {
    throw new Error("sessionStorage unavailable");
  }
  store.setItem(SESSION_STORAGE_KEY, JSON.stringify(session));
}

export function clearSession(store = storage()): void {
  store?.removeItem(SESSION_STORAGE_KEY);
}

export function getSession(
  now = () => new Date(),
  store = storage(),
): ClientSession | null {
  if (!store) {
    return null;
  }
  const raw = store.getItem(SESSION_STORAGE_KEY);
  if (!raw) {
    return null;
  }
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    clearSession(store);
    return null;
  }
  if (
    !parsed ||
    typeof parsed !== "object" ||
    typeof (parsed as ClientSession).sessionToken !== "string" ||
    typeof (parsed as ClientSession).nickname !== "string" ||
    typeof (parsed as ClientSession).expiresAt !== "string"
  ) {
    clearSession(store);
    return null;
  }
  const session = parsed as ClientSession;
  const expires = Date.parse(session.expiresAt);
  if (Number.isNaN(expires) || expires <= now().getTime()) {
    clearSession(store);
    return null;
  }
  return session;
}
