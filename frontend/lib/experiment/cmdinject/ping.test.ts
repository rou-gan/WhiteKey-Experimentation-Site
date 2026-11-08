import { describe, expect, it } from "vitest";
import { buildPingArgs } from "./ping";

const injectionProbes = [
  {
    name: "semicolon",
    input: "127.0.0.1;id",
    purpose: "セミコロンによる追加コマンドが拒否されること",
  },
  {
    name: "pipe",
    input: "127.0.0.1|id",
    purpose: "パイプによる追加コマンドが拒否されること",
  },
  {
    name: "and",
    input: "127.0.0.1&&id",
    purpose: "&& による追加コマンドが拒否されること",
  },
  {
    name: "substitution",
    input: "$(id)",
    purpose: "コマンド置換が拒否されること",
  },
  {
    name: "backtick",
    input: "`id`",
    purpose: "バッククォート置換が拒否されること",
  },
] as const;

// 目的: シェルを介さず argv で ping すること（OS コマンドインジェクション防止）。
describe("buildPingArgs", () => {
  it("uses argv without shell", () => {
    const args = buildPingArgs("example.com");
    expect(args).toEqual(["ping", "-c", "1", "example.com"]);
  });

  it.each(injectionProbes)("$name — $purpose", ({ input }) => {
    expect(
      () => buildPingArgs(input),
      `accepted unsafe host ${JSON.stringify(input)} (command injection risk)`,
    ).toThrow();
  });
});
