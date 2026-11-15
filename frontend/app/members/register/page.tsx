"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
import { registerMember } from "@/lib/member/api";
import { registerErrorMessage } from "@/lib/member/errors";
import { validateMemberInput } from "@/lib/member/validation";
import styles from "../members.module.css";

export default function RegisterPage() {
  const router = useRouter();
  const [loginId, setLoginId] = useState("");
  const [password, setPassword] = useState("");
  const [nickname, setNickname] = useState("");
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    if (!validateMemberInput({ loginId, password, nickname })) {
      setError(registerErrorMessage());
      return;
    }
    setPending(true);
    try {
      await registerMember({ loginId, password, nickname });
      router.push("/members/login");
    } catch {
      setError(registerErrorMessage());
    } finally {
      setPending(false);
    }
  }

  return (
    <main className={styles.page}>
      <p className={styles.brand}>しろかぎ</p>
      <h1 className={styles.title}>会員登録</h1>
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
            autoComplete="new-password"
            value={password}
            onChange={(ev) => setPassword(ev.target.value)}
            required
          />
        </label>
        <label className={styles.label}>
          ニックネーム
          <input
            className={styles.input}
            name="nickname"
            autoComplete="nickname"
            value={nickname}
            onChange={(ev) => setNickname(ev.target.value)}
            required
          />
        </label>
        <div className={styles.actions}>
          <button className={styles.button} type="submit" disabled={pending}>
            登録
          </button>
        </div>
        {error ? (
          <p className={styles.error} role="alert">
            {error}
          </p>
        ) : null}
      </form>
      <p className={styles.nav}>
        <Link href="/members/login">ログインへ</Link>
      </p>
    </main>
  );
}
