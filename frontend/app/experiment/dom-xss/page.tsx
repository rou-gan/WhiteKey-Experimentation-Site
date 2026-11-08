import { DomXssPreview } from "./DomXssPreview";

/**
 * DOM ベース XSS 対策の確認ページ。
 * クエリ `html` は埋め込み前にエスケープされる。
 */
export default async function DomXssExperimentPage({
  searchParams,
}: {
  searchParams: Promise<{ html?: string }>;
}) {
  const params = await searchParams;
  const html = params.html ?? "";

  return (
    <main style={{ padding: "2rem", fontFamily: "system-ui, sans-serif" }}>
      <h1>DOM XSS 実験</h1>
      <p>
        クエリ <code>?html=</code>{" "}
        の値を HTML 埋め込み前にエスケープして表示します（教育・UT 用）。
      </p>
      <DomXssPreview html={html} />
    </main>
  );
}
