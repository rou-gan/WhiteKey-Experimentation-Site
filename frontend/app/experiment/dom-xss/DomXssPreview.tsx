"use client";

import { renderPreviewHTML } from "@/lib/experiment/domxss/render";

type Props = {
  html: string;
};

/**
 * クエリ由来の文字列をエスケープ済み HTML 断片として埋め込む。
 * （DOM XSS 実験ページ。埋め込み前エスケープは renderPreviewHTML が担当）
 */
export function DomXssPreview({ html }: Props) {
  const markup = renderPreviewHTML(html);

  return (
    <section
      aria-label="dom-xss-preview"
      // エスケープ済み断片の表示用。生のユーザー入力は渡さない。
      dangerouslySetInnerHTML={{ __html: markup }}
    />
  );
}
