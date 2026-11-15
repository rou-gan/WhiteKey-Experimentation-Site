import { describe, expect, it } from "vitest";
import { validateMemberInput } from "./validation";

const valid = {
  loginId: "member01",
  password: "Abcd123!",
  nickname: "しろかぎ",
} as const;

// ケース: V-01
// 目的: 妥当な値を受理すること。
describe("validateMemberInput", () => {
  it("V-01 妥当な値を受理する", () => {
    expect(validateMemberInput({ ...valid })).toBe(true);
  });

  // ケース: V-02
  // 目的: 7 文字のログインIDを拒否すること。
  it("V-02 7文字のログインIDを拒否する", () => {
    expect(
      validateMemberInput({
        ...valid,
        loginId: "member0",
      }),
    ).toBe(false);
  });

  // ケース: V-03
  // 目的: 記号を含まないパスワードを拒否すること。
  it("V-03 記号なしパスワードを拒否する", () => {
    expect(
      validateMemberInput({
        ...valid,
        password: "Abcd1234",
      }),
    ).toBe(false);
  });

  // ケース: V-04
  // 目的: 51 ルーンのニックネームを拒否すること。
  it("V-04 51ルーンのニックネームを拒否する", () => {
    const nickname = "あ".repeat(51);
    expect(Array.from(nickname).length).toBe(51);
    expect(
      validateMemberInput({
        ...valid,
        nickname,
      }),
    ).toBe(false);
  });
});
