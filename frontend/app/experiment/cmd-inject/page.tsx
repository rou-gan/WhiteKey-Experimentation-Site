import { buildPingArgs } from "@/lib/experiment/cmdinject/ping";

/**
 * OS コマンドインジェクション実験ページ（教育・UT 用）。
 * `buildPingArgs` の組み立て結果を表示する（このページでは OS 実行しない）。
 */
export default async function CmdInjectExperimentPage({
  searchParams,
}: {
  searchParams: Promise<{ host?: string }>;
}) {
  const params = await searchParams;
  const host = params.host ?? "";

  let args: string[] | null = null;
  let error = "";

  if (host) {
    try {
      args = buildPingArgs(host);
    } catch {
      error = "host rejected or invalid";
    }
  }

  return (
    <main style={{ padding: "2rem", fontFamily: "system-ui, sans-serif" }}>
      <h1>OS コマンドインジェクション実験</h1>
      <p>
        クエリ <code>?host=</code> から検証済み argv を組み立てて表示します（このページでは OS
        実行しません。HTTP API は <code>/api/experiment/ping</code>）。
      </p>
      {error ? <p role="status">{error}</p> : null}
      {args ? (
        <pre style={{ background: "#111", color: "#eee", padding: "1rem" }}>
          {JSON.stringify({ host, args }, null, 2)}
        </pre>
      ) : (
        <p>
          例: <code>/experiment/cmd-inject?host=example.com</code>
        </p>
      )}
    </main>
  );
}
