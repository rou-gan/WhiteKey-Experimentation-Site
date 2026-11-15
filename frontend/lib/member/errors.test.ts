import { describe, expect, it } from "vitest";
import {
  DELETE_ERROR_MESSAGE,
  LOGIN_ERROR_MESSAGE,
  REGISTER_ERROR_MESSAGE,
  deleteErrorMessage,
  loginErrorMessage,
  registerErrorMessage,
} from "./errors";

describe("member error messages", () => {
  // ケース: E-01
  // 目的: 登録の 400 と 409 が同じ文言になること。
  it("E-01 登録の400と409が同じ文言になる", () => {
    expect(registerErrorMessage(400)).toBe(REGISTER_ERROR_MESSAGE);
    expect(registerErrorMessage(409)).toBe(REGISTER_ERROR_MESSAGE);
    expect(REGISTER_ERROR_MESSAGE).toBe("登録できませんでした。");
  });

  // ケース: E-02
  // 目的: ログイン失敗を固定文言にすること。
  it("E-02 ログイン失敗を固定文言にする", () => {
    expect(loginErrorMessage(401)).toBe(LOGIN_ERROR_MESSAGE);
    expect(LOGIN_ERROR_MESSAGE).toBe("ログインできませんでした。");
  });

  // ケース: E-03
  // 目的: 退会失敗を固定文言にすること。
  it("E-03 退会失敗を固定文言にする", () => {
    expect(deleteErrorMessage(404)).toBe(DELETE_ERROR_MESSAGE);
    expect(DELETE_ERROR_MESSAGE).toBe("退会できませんでした。");
  });
});
