/**
 * OS コマンドインジェクション実験用の ping argv 組み立て。
 * シェルを介さず、入力検証後に引数配列のみを返す。
 * ユニットテストでは実行しない。
 */
const SAFE_HOST = /^[A-Za-z0-9][A-Za-z0-9.-]{0,253}$/;

export function buildPingArgs(host: string): string[] {
  const trimmed = host.trim();
  if (!trimmed) {
    throw new Error("host is required");
  }
  if (!SAFE_HOST.test(trimmed)) {
    throw new Error("invalid host");
  }
  return ["ping", "-c", "1", trimmed];
}
