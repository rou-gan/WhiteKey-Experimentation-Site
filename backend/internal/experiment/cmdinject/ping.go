package cmdinject

import (
	"fmt"
	"os/exec"
)

// PingArgs はホスト名から ping 実行用の argv を組み立てる。
//
// INTENTIONAL VULNERABILITY (experiment):
// ユーザー入力をシェル文字列へ連結し `sh -c` で実行する形を返す。
// `;` `|` `$(...)` などが注入されると OS コマンドインジェクションになる。
// UT で検知し、後続で「検証 + 引数配列（シェルなし）」へ置き換える前提。
func PingArgs(host string) ([]string, error) {
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}
	// 危険: シェル経由 + 文字列連結
	return []string{"sh", "-c", "ping -c 1 " + host}, nil
}

// RunPing は PingArgs の結果を実行する（実験ハンドラ用）。
func RunPing(host string) ([]byte, error) {
	args, err := PingArgs(host)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(args[0], args[1:]...) // #nosec G204 -- 意図的なコマンドインジェクション実験
	return cmd.CombinedOutput()
}
