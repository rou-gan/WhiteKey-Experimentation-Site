import { DomXssPreview } from "./DomXssPreview";

/**
 * DOM ベース XSS 実験ページ。
 * クエリ `html` をクライアント側で HTML シンクへ渡す（意図的に不安全）。
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
        の値をクライアントで HTML として埋め込みます（教育・UT 用の意図的脆弱実装）。
      </p>
      <DomXssPreview html={html} />
    </main>
  );
}
