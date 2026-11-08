import { describe, expect, it } from "vitest";
import { escapeHTML, renderPreviewHTML } from "./render";

// XSS プローブは攻撃手順ではなく、エスケープ漏れを検出するためのテスト用入力。
const xssProbes = [
  {
    name: "script_tag",
    input: `<script>alert(1)</script>`,
    purpose: "script 要素が生のまま HTML に埋め込まれないこと",
  },
  {
    name: "img_onerror",
    input: `<img src=x onerror=alert(1)>`,
    purpose: "イベントハンドラ付きタグが生反射しないこと",
  },
  {
    name: "attribute_breakout",
    input: `" onclick="alert(1)`,
    purpose: "引用符による属性脱出がエスケープされること",
  },
] as const;

// 目的: 信頼できない入力を HTML 埋め込み前にエスケープし、DOM ベース XSS を防ぐ。
describe("renderPreviewHTML", () => {
  it.each(xssProbes)("$name — $purpose", ({ input }) => {
    const out = renderPreviewHTML(input);

    expect(out, "raw user input reflected in HTML (DOM XSS risk)").not.toContain(
      input,
    );

    const want = `<div class="preview">${escapeHTML(input)}</div>`;
    expect(out).toBe(want);
  });

  it("平文はそのまま描画できる", () => {
    expect(renderPreviewHTML("しろかぎ")).toBe(
      `<div class="preview">しろかぎ</div>`,
    );
  });
});
