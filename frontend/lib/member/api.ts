import { apiBaseUrl } from "@/lib/api";

export type RegisterResult = {
  loginId: string;
  nickname: string;
};

export type ChallengeResult = {
  challengeId: string;
  salt: string;
  challenge: string;
};

export type VerifyResult = {
  sessionToken: string;
  nickname: string;
  expiresAt: string;
};

export class MemberApiError extends Error {
  readonly status: number;

  constructor(status: number, message = "member api error") {
    super(message);
    this.name = "MemberApiError";
    this.status = status;
  }
}

type FetchLike = typeof fetch;

async function parseJson(res: Response): Promise<unknown> {
  return (await res.json()) as unknown;
}

export async function registerMember(
  input: { loginId: string; password: string; nickname: string },
  fetchImpl: FetchLike = fetch,
): Promise<RegisterResult> {
  const res = await fetchImpl(`${apiBaseUrl()}/members`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      loginId: input.loginId,
      password: input.password,
      nickname: input.nickname,
    }),
    cache: "no-store",
  });
  if (res.status !== 201) {
    throw new MemberApiError(res.status);
  }
  const body = (await parseJson(res)) as RegisterResult;
  return { loginId: body.loginId, nickname: body.nickname };
}

export async function issueChallenge(
  loginId: string,
  fetchImpl: FetchLike = fetch,
): Promise<ChallengeResult> {
  const res = await fetchImpl(`${apiBaseUrl()}/members/auth/challenge`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ loginId }),
    cache: "no-store",
  });
  if (res.status !== 200) {
    throw new MemberApiError(res.status);
  }
  const body = (await parseJson(res)) as ChallengeResult;
  return {
    challengeId: body.challengeId,
    salt: body.salt,
    challenge: body.challenge,
  };
}

export async function verifyChallenge(
  input: { challengeId: string; response: string },
  fetchImpl: FetchLike = fetch,
): Promise<VerifyResult> {
  const res = await fetchImpl(`${apiBaseUrl()}/members/auth/verify`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      challengeId: input.challengeId,
      response: input.response,
    }),
    cache: "no-store",
  });
  if (res.status !== 200) {
    throw new MemberApiError(res.status);
  }
  const body = (await parseJson(res)) as VerifyResult;
  return {
    sessionToken: body.sessionToken,
    nickname: body.nickname,
    expiresAt: body.expiresAt,
  };
}

export async function deleteMe(
  sessionToken: string,
  fetchImpl: FetchLike = fetch,
): Promise<void> {
  const res = await fetchImpl(`${apiBaseUrl()}/members/me`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${sessionToken}` },
    cache: "no-store",
  });
  if (res.status !== 204) {
    throw new MemberApiError(res.status);
  }
}
