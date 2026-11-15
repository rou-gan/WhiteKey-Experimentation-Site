"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
import { issueChallenge, verifyChallenge } from "@/lib/member/api";
import { computeAuthResponse } from "@/lib/member/auth";
import { loginErrorMessage } from "@/lib/member/errors";
import { saveSession } from "@/lib/member/session";
import { isValidLoginId, isValidPassword } from "@/lib/member/validation";
import styles from "../members.module.css";

export default function LoginPage() {
  const router = useRouter();
  const [loginId, setLoginId] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    if (!isValidLoginId(loginId) || !isValidPassword(password)) {
      setError(loginErrorMessage());
      return;
    }
    setPending(true);
    try {
      const challenge = await issueChallenge(loginId);
      const response = await computeAuthResponse(
        password,
        challenge.salt,
        challenge.challenge,
      );
      const verified = await verifyChallenge({
        challengeId: challenge.challengeId,
        response,
      });
      saveSession({
        sessionToken: verified.sessionToken,
        nickname: verified.nickname,
        expiresAt: verified.expiresAt,
      });
      router.push("/members/me");
    } catch {
      setError(loginErrorMessage());
    } finally {
      setPending(false);
    }
  }

  return (
    <main className={styles.page}>
      <p className={styles.brand}>しろかぎ</p>
      <h1 className={styles.title}>ログイン</h1>
      <form className={styles.form} onSubmit={onSubmit}>
        <label className={styles.label}>
          ログインID
          <input
            className={styles.input}
            name="loginId"
            autoComplete="username"
            value={loginId}
            onChange={(ev) => setLoginId(ev.target.value)}
            required
          />
        </label>
        <label className={styles.label}>
          パスワード
          <input
            className={styles.input}
            name="password"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(ev) => setPassword(ev.target.value)}
            required
          />
        </label>
        <div className={styles.actions}>
          <button className={styles.button} type="submit" disabled={pending}>
            ログイン
          </button>
        </div>
        {error ? (
          <p className={styles.error} role="alert">
            {error}
          </p>
        ) : null}
      </form>
      <p className={styles.nav}>
        <Link href="/members/register">登録へ</Link>
      </p>
    </main>
  );
}
