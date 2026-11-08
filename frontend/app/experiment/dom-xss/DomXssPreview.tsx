"use client";

import { renderPreviewHTML } from "@/lib/experiment/domxss/render";

type Props = {
  html: string;
};

/**
 * INTENTIONAL DOM XSS sink:
 * renderPreviewHTML の未エスケープ出力を dangerouslySetInnerHTML へ渡す。
 */
export function DomXssPreview({ html }: Props) {
  const markup = renderPreviewHTML(html);

  return (
    <section
      aria-label="dom-xss-preview"
      // eslint-disable-next-line react/no-danger -- 意図的な DOM XSS 実験シンク
      dangerouslySetInnerHTML={{ __html: markup }}
    />
  );
}
