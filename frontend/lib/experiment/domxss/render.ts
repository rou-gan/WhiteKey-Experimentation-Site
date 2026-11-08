/**
 * DOM ベース XSS 実験用の HTML 断片生成。
 *
 * INTENTIONAL VULNERABILITY (experiment):
 * 信頼できない入力をエスケープせず HTML に埋め込む。
 * ブラウザの innerHTML / dangerouslySetInnerHTML シンクへ渡すと DOM XSS になる。
 * UT で検知し、後続でサニタイズ／エスケープへ置き換える前提。
 */
export function renderPreviewHTML(userHTML: string): string {
  return `<div class="preview">${userHTML}</div>`;
}
