import { describe, expect, it } from "vitest";
import { computeAuthResponse } from "./auth";

/**
 * 固定ベクトル（EDS-MEM-1 付録 A: m=19456, t=2, p=1, 出力32B）。
 * salt = 16×0x01, challenge = 32×0x02, password = Abcd123!
 * 期待 response は Go `argon2.IDKey` + HMAC-SHA256 と同系で固定した仕様値。
 */
const FIXED = {
  password: "Abcd123!",
  saltB64: "AQEBAQEBAQEBAQEBAQEBAQ==",
  challengeB64: "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI=",
  responseB64: "gifbJsUwx1O6tovoAo/hM3VQ1kGv6nMbTJsUkXFb/Uw=",
} as const;

describe("computeAuthResponse", () => {
  // ケース: A-01
  // 目的: 固定ベクトルで応答が期待 Base64 と一致すること。
  it("A-01 固定ベクトルで応答が一致する", async () => {
    const response = await computeAuthResponse(
      FIXED.password,
      FIXED.saltB64,
      FIXED.challengeB64,
    );
    expect(response).toBe(FIXED.responseB64);
  }, 30_000);

  // ケース: A-02
  // 目的: 不正 Base64 の salt では応答を返さないこと。
  it("A-02 不正Base64のsaltでは応答を返さない", async () => {
    await expect(
      computeAuthResponse(FIXED.password, "@@@", FIXED.challengeB64),
    ).rejects.toThrow(/invalid base64 salt/);
  });
});
