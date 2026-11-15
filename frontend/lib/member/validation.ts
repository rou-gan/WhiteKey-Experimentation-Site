/** EDS-MEM-1 §6 / EDS-MEM-3 §6 のクライアント側事前チェック。最終判定は API。 */

const LOGIN_ID_RE = /^[A-Za-z0-9]{8,15}$/;

const ALLOWED_PASSWORD_SYMBOLS = new Set(
  `!"#$%&'()*+,-./:;<=>?@[\\]^_\`{|}~`.split(""),
);

export type MemberInput = {
  loginId: string;
  password: string;
  nickname: string;
};

export function isValidLoginId(loginId: string): boolean {
  return LOGIN_ID_RE.test(loginId);
}

export function isValidPassword(password: string): boolean {
  if (password.length < 8 || password.length > 15) {
    return false;
  }
  let hasLetter = false;
  let hasDigit = false;
  let hasSymbol = false;
  for (let i = 0; i < password.length; i++) {
    const c = password.charCodeAt(i);
    if (c > 0x7f || c === 0x20) {
      return false;
    }
    const ch = password[i]!;
    if ((c >= 0x41 && c <= 0x5a) || (c >= 0x61 && c <= 0x7a)) {
      hasLetter = true;
    } else if (c >= 0x30 && c <= 0x39) {
      hasDigit = true;
    } else if (ALLOWED_PASSWORD_SYMBOLS.has(ch)) {
      hasSymbol = true;
    } else {
      return false;
    }
  }
  return hasLetter && hasDigit && hasSymbol;
}

export function isValidNickname(nickname: string): boolean {
  const runes = Array.from(nickname);
  if (runes.length < 1 || runes.length > 50) {
    return false;
  }
  return runes.some((r) => !/\s/u.test(r));
}

export function validateMemberInput(input: MemberInput): boolean {
  return (
    isValidLoginId(input.loginId) &&
    isValidPassword(input.password) &&
    isValidNickname(input.nickname)
  );
}
