"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { deleteMe, MemberApiError } from "@/lib/member/api";
import { deleteErrorMessage } from "@/lib/member/errors";
import { clearSession, getSession, type ClientSession } from "@/lib/member/session";
import styles from "../members.module.css";

export default function MemberHomePage() {
  const router = useRouter();
  const [session, setSession] = useState<ClientSession | null>(null);
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const current = getSession();
    if (!current) {
      router.replace("/members/login");
      return;
    }
    setSession(current);
    setReady(true);
  }, [router]);

  function onLogout() {
    clearSession();
    router.push("/members/login");
  }

  async function onDelete() {
    if (!session) {
      return;
    }
    const ok = window.confirm("退会しますか？この操作は取り消せません。");
    if (!ok) {
      return;
    }
    setError("");
    setPending(true);
    try {
      await deleteMe(session.sessionToken);
      clearSession();
      router.push("/members/login");
    } catch (err) {
      if (err instanceof MemberApiError && err.status === 401) {
        clearSession();
        router.push("/members/login");
        return;
      }
      setError(deleteErrorMessage());
    } finally {
      setPending(false);
    }
  }

  if (!ready || !session) {
    return (
      <main className={styles.page}>
        <p className={styles.brand}>しろかぎ</p>
      </main>
    );
  }

  return (
    <main className={styles.page}>
      <p className={styles.brand}>しろかぎ</p>
      <h1 className={styles.title}>会員ホーム</h1>
      <p className={styles.nickname}>{session.nickname}</p>
      <div className={styles.actions}>
        <button className={styles.button} type="button" onClick={onLogout}>
          ログアウト
        </button>
        <button
          className={styles.button}
          type="button"
          onClick={onDelete}
          disabled={pending}
        >
          退会
        </button>
      </div>
      {error ? (
        <p className={styles.error} role="alert">
          {error}
        </p>
      ) : null}
    </main>
  );
}
