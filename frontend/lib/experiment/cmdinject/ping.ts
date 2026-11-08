/**
 * OS コマンドインジェクション実験用の ping argv 組み立て。
 *
 * INTENTIONAL VULNERABILITY (experiment):
 * ユーザー入力をシェル文字列へ連結し `sh -c` で渡す形を返す。
 * Node の child_process で実行すると OS コマンドインジェクションになる。
 * UT で検知し、後続で「検証 + 引数配列（シェルなし）」へ置き換える前提。
 *
 * 注: このモジュールはコマンド文字列／argv の組み立てのみを担当する。
 * ユニットテストでは実行しない。
 */
export function buildPingArgs(host: string): string[] {
  if (!host.trim()) {
    throw new Error("host is required");
  }
  // 危険: シェル経由 + 文字列連結
  return ["sh", "-c", `ping -c 1 ${host}`];
}
