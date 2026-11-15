import { afterEach, describe, expect, it } from "vitest";
import {
  SESSION_STORAGE_KEY,
  clearSession,
  getSession,
  saveSession,
} from "./session";

function memoryStore() {
  const map = new Map<string, string>();
  return {
    getItem(key: string) {
      return map.has(key) ? map.get(key)! : null;
    },
    setItem(key: string, value: string) {
      map.set(key, value);
    },
    removeItem(key: string) {
      map.delete(key);
    },
  };
}

describe("client session", () => {
  afterEach(() => {
    // no shared global store in these tests
  });

  // ケース: S-01
  // 目的: 保存直後に同一値を取得できること。
  it("S-01 保存直後に同一値を取得できる", () => {
    const store = memoryStore();
    const expiresAt = new Date(Date.now() + 60_000).toISOString();
    saveSession(
      { sessionToken: "abc", nickname: "しろかぎ", expiresAt },
      store,
    );
    expect(store.getItem(SESSION_STORAGE_KEY)).not.toBeNull();
    expect(getSession(() => new Date(), store)).toEqual({
      sessionToken: "abc",
      nickname: "しろかぎ",
      expiresAt,
    });
  });

  // ケース: S-02
  // 目的: clear 後にセッションが無いこと。
  it("S-02 clear後にセッションが無い", () => {
    const store = memoryStore();
    const expiresAt = new Date(Date.now() + 60_000).toISOString();
    saveSession(
      { sessionToken: "abc", nickname: "しろかぎ", expiresAt },
      store,
    );
    clearSession(store);
    expect(getSession(() => new Date(), store)).toBeNull();
  });

  // ケース: S-03
  // 目的: 期限切れを未認証にすること。
  it("S-03 期限切れを未認証にする", () => {
    const store = memoryStore();
    const expiresAt = new Date(Date.now() - 1_000).toISOString();
    saveSession(
      { sessionToken: "abc", nickname: "しろかぎ", expiresAt },
      store,
    );
    expect(getSession(() => new Date(), store)).toBeNull();
  });
});
