/** EDS-MEM-3 §5 / §9 の固定汎用文言。HTTP 本文は写さない。 */

export const REGISTER_ERROR_MESSAGE = "登録できませんでした。";
export const LOGIN_ERROR_MESSAGE = "ログインできませんでした。";
export const DELETE_ERROR_MESSAGE = "退会できませんでした。";

export function registerErrorMessage(_status?: number): string {
  return REGISTER_ERROR_MESSAGE;
}

export function loginErrorMessage(_status?: number): string {
  return LOGIN_ERROR_MESSAGE;
}

export function deleteErrorMessage(_status?: number): string {
  return DELETE_ERROR_MESSAGE;
}
