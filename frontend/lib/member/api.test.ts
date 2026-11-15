import { afterEach, describe, expect, it, vi } from "vitest";
import {
  MemberApiError,
  deleteMe,
  issueChallenge,
  registerMember,
  verifyChallenge,
} from "./api";

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("member api", () => {
  // ケース: C-01
  // 目的: 登録 201 で公開フィールドのみ返すこと。
  it("C-01 登録201で公開フィールドのみ返す", async () => {
    const fetchImpl = vi.fn(async () =>
      jsonResponse(201, {
        loginId: "member01",
        nickname: "しろかぎ",
        password: "should-not-leak",
      }),
    );
    const result = await registerMember(
      {
        loginId: "member01",
        password: "Abcd123!",
        nickname: "しろかぎ",
      },
      fetchImpl as unknown as typeof fetch,
    );
    expect(result).toEqual({ loginId: "member01", nickname: "しろかぎ" });
    expect(result).not.toHaveProperty("password");
  });

  // ケース: C-02
  // 目的: 登録 409 をエラーにすること。
  it("C-02 登録409をエラーにする", async () => {
    const fetchImpl = vi.fn(async () => jsonResponse(409, { error: "x" }));
    await expect(
      registerMember(
        {
          loginId: "member01",
          password: "Abcd123!",
          nickname: "しろかぎ",
        },
        fetchImpl as unknown as typeof fetch,
      ),
    ).rejects.toBeInstanceOf(MemberApiError);
  });

  // ケース: C-03
  // 目的: 挑戦 200 の JSON をそのまま返すこと。
  it("C-03 挑戦200のJSONを返す", async () => {
    const fetchImpl = vi.fn(async () =>
      jsonResponse(200, {
        challengeId: "chal_1",
        salt: "c2FsdA==",
        challenge: "Y2hhbA==",
      }),
    );
    const result = await issueChallenge(
      "member01",
      fetchImpl as unknown as typeof fetch,
    );
    expect(result).toEqual({
      challengeId: "chal_1",
      salt: "c2FsdA==",
      challenge: "Y2hhbA==",
    });
  });

  // ケース: C-04
  // 目的: verify 200 でセッション情報を返すこと。
  it("C-04 verify200でセッション情報を返す", async () => {
    const fetchImpl = vi.fn(async () =>
      jsonResponse(200, {
        sessionToken: "tok",
        nickname: "しろかぎ",
        expiresAt: "2026-01-01T00:00:00Z",
      }),
    );
    const result = await verifyChallenge(
      { challengeId: "chal_1", response: "resp" },
      fetchImpl as unknown as typeof fetch,
    );
    expect(result).toEqual({
      sessionToken: "tok",
      nickname: "しろかぎ",
      expiresAt: "2026-01-01T00:00:00Z",
    });
  });

  // ケース: C-05
  // 目的: verify 401 をエラーにすること。
  it("C-05 verify401をエラーにする", async () => {
    const fetchImpl = vi.fn(async () => jsonResponse(401, { error: "x" }));
    await expect(
      verifyChallenge(
        { challengeId: "chal_1", response: "resp" },
        fetchImpl as unknown as typeof fetch,
      ),
    ).rejects.toBeInstanceOf(MemberApiError);
  });

  // ケース: C-06
  // 目的: 退会 204 を成功とすること。
  it("C-06 退会204を成功とする", async () => {
    const fetchImpl = vi.fn(async () => new Response(null, { status: 204 }));
    await expect(
      deleteMe("tok", fetchImpl as unknown as typeof fetch),
    ).resolves.toBeUndefined();
  });

  // ケース: C-07
  // 目的: 退会で Bearer を送ること。
  it("C-07 退会でBearerを送る", async () => {
    const fetchImpl = vi.fn(async () => new Response(null, { status: 204 }));
    await deleteMe("tok", fetchImpl as unknown as typeof fetch);
    const init = fetchImpl.mock.calls[0]?.[1] as RequestInit;
    const headers = new Headers(init.headers);
    expect(headers.get("Authorization")).toBe("Bearer tok");
  });

  // ケース: C-08
  // 目的: verify ボディに password が無いこと。
  it("C-08 verifyボディにpasswordが無い", async () => {
    const fetchImpl = vi.fn(async () =>
      jsonResponse(200, {
        sessionToken: "tok",
        nickname: "しろかぎ",
        expiresAt: "2026-01-01T00:00:00Z",
      }),
    );
    await verifyChallenge(
      { challengeId: "chal_1", response: "resp" },
      fetchImpl as unknown as typeof fetch,
    );
    const init = fetchImpl.mock.calls[0]?.[1] as RequestInit;
    const body = JSON.parse(String(init.body)) as Record<string, unknown>;
    expect(body).toEqual({ challengeId: "chal_1", response: "resp" });
    expect(body).not.toHaveProperty("password");
  });
});
