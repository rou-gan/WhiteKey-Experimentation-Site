import { argon2id, createHMAC, createSHA256 } from "hash-wasm";

/** EDS-MEM-1 付録 A 最低ライン（弱めない）。 */
export const ARGON2_MEMORY_KIB = 19_456;
export const ARGON2_ITERATIONS = 2;
export const ARGON2_PARALLELISM = 1;
export const ARGON2_HASH_LENGTH = 32;

function decodeBase64(value: string, label: string): Uint8Array {
  // 厳密に拒否（緩い atob / Buffer デコードに任せない）。
  if (!/^[A-Za-z0-9+/]*={0,2}$/.test(value) || value.length % 4 !== 0) {
    throw new Error(`invalid base64 ${label}`);
  }
  try {
    const bin = atob(value);
    const out = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) {
      out[i] = bin.charCodeAt(i);
    }
    return out;
  } catch {
    throw new Error(`invalid base64 ${label}`);
  }
}

function encodeBase64(bytes: Uint8Array): string {
  let s = "";
  for (let i = 0; i < bytes.length; i++) {
    s += String.fromCharCode(bytes[i]!);
  }
  return btoa(s);
}

/**
 * secret = Argon2id(password, salt)
 * response = Base64(HMAC-SHA256(secret, challengeBytes))
 */
export async function computeAuthResponse(
  password: string,
  saltB64: string,
  challengeB64: string,
): Promise<string> {
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
  const macHex = hasher.digest("hex");
  return encodeBase64(hexToBytes(macHex));
}

function hexToBytes(hex: string): Uint8Array {
  if (hex.length % 2 !== 0) {
    throw new Error("invalid hex");
  }
  const out = new Uint8Array(hex.length / 2);
  for (let i = 0; i < out.length; i++) {
    out[i] = Number.parseInt(hex.slice(i * 2, i * 2 + 2), 16);
  }
  return out;
}
