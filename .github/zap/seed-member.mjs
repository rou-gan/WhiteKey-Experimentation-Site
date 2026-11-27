/**
 * ZAP 認証付きスキャン用: 使い捨て会員を登録し sessionToken を得る。
 * EDS-MEM-1 と同じ Argon2id + HMAC-SHA256（frontend/lib/member/auth.ts と同パラメータ）。
 *
 * 使い方:
 *   ZAP_TARGET_API=http://localhost:8080 node seed-member.mjs
 * 出力: 標準出力に sessionToken 1 行（ログには出さない）
 */
import { randomBytes } from "node:crypto";
import { argon2id, createHMAC, createSHA256 } from "hash-wasm";

const ARGON2_MEMORY_KIB = 19_456;
const ARGON2_ITERATIONS = 2;
const ARGON2_PARALLELISM = 1;
const ARGON2_HASH_LENGTH = 32;
const PASSWORD = "Abcd123!";

const apiBase = (process.env.ZAP_TARGET_API || "http://localhost:8080").replace(
  /\/$/,
  "",
);

function decodeBase64(value, label) {
  if (!/^[A-Za-z0-9+/]*={0,2}$/.test(value) || value.length % 4 !== 0) {
    throw new Error(`invalid base64 ${label}`);
  }
  const bin = Buffer.from(value, "base64");
  return new Uint8Array(bin);
}

function encodeBase64(bytes) {
  return Buffer.from(bytes).toString("base64");
}

function hexToBytes(hex) {
  if (hex.length % 2 !== 0) {
    throw new Error("invalid hex");
  }
  const out = new Uint8Array(hex.length / 2);
  for (let i = 0; i < out.length; i++) {
    out[i] = Number.parseInt(hex.slice(i * 2, i * 2 + 2), 16);
  }
  return out;
}

async function computeAuthResponse(password, saltB64, challengeB64) {
  const salt = decodeBase64(saltB64, "salt");
  const challenge = decodeBase64(challengeB64, "challenge");
  const secretHex = await argon2id({
    password,
    salt,
    parallelism: ARGON2_PARALLELISM,
    iterations: ARGON2_ITERATIONS,
    memorySize: ARGON2_MEMORY_KIB,
    hashLength: ARGON2_HASH_LENGTH,
    outputType: "hex",
  });
  const secret = hexToBytes(secretHex);
  const hasher = await createHMAC(createSHA256(), secret);
  hasher.update(challenge);
  return encodeBase64(hexToBytes(hasher.digest("hex")));
}

async function postJson(path, body) {
  const res = await fetch(`${apiBase}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: JSON.stringify(body),
  });
  const text = await res.text();
  let json = null;
  if (text) {
    try {
      json = JSON.parse(text);
    } catch {
      json = { raw: text };
    }
  }
  return { status: res.status, json };
}

/** loginId: 半角英数字のみ・8〜15 文字（EDS-MEM-1）。 */
function randLoginId() {
  return `zap${randomBytes(4).toString("hex")}`; // zap + 8 hex = 11
}

async function main() {
  const loginId = randLoginId();
  const nickname = `zapnick ${loginId}`;

  const reg = await postJson("/members", {
    loginId,
    password: PASSWORD,
    nickname,
  });
  if (reg.status !== 201) {
    console.error(`register failed: HTTP ${reg.status}`, JSON.stringify(reg.json));
    process.exit(1);
  }

  const chal = await postJson("/members/auth/challenge", { loginId });
  if (chal.status !== 200 || !chal.json?.challengeId) {
    console.error(`challenge failed: HTTP ${chal.status}`, JSON.stringify(chal.json));
    process.exit(1);
  }

  const response = await computeAuthResponse(
    PASSWORD,
    chal.json.salt,
    chal.json.challenge,
  );
  const ver = await postJson("/members/auth/verify", {
    challengeId: chal.json.challengeId,
    response,
  });
  if (ver.status !== 200 || !ver.json?.sessionToken) {
    console.error(`verify failed: HTTP ${ver.status}`, JSON.stringify(ver.json));
    process.exit(1);
  }

  // token 全文は stdout のみ（stderr には出さない）
  process.stdout.write(`${ver.json.sessionToken}\n`);
  console.error(`seed ok loginId=${loginId}`);
}

main().catch((err) => {
  console.error(err instanceof Error ? err.message : err);
  process.exit(1);
});
