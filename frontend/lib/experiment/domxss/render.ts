/**
 * DOM ベース XSS 実験用の HTML 断片生成。
 * 信頼できない入力は HTML エスケープしてから埋め込み、
 * innerHTML / dangerouslySetInnerHTML シンクでもスクリプト実行を防ぐ。
 */
export function escapeHTML(value: string): string {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

export function renderPreviewHTML(userHTML: string): string {
  return `<div class="preview">${escapeHTML(userHTML)}</div>`;
}
