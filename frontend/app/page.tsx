import { fetchRootMessage } from "@/lib/api";
import styles from "./page.module.css";

export default async function HomePage() {
  let message = "";
  let error = "";

  try {
    const data = await fetchRootMessage();
    message = data.message;
  } catch {
    error = "メッセージを取得できませんでした。";
  }

  return (
    <main className={styles.hero}>
      <div className={styles.veil} aria-hidden />
      <section className={styles.copy}>
        <p className={styles.brand}>しろかぎ</p>
        <h1 className={styles.headline}>Design by Security</h1>
        <p className={styles.lead}>
          コンテナ・依存関係・SBOM を低コストで検証するセキュリティ実験場。
        </p>
        <div className={styles.status} role="status">
          {error ? (
            <p className={styles.error}>{error}</p>
          ) : (
            <p className={styles.message}>{message}</p>
          )}
        </div>
        <p className={styles.memberLinks}>
          <a href="/members/register">登録</a>
          {" · "}
          <a href="/members/login">ログイン</a>
        </p>
      </section>
    </main>
  );
}
