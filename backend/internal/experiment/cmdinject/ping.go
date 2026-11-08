package cmdinject

import (
	"fmt"
	"os/exec"
	"regexp"
)

// ホスト名／IPv4 向けの許可文字（シェルメタ文字を含まない）。
var safeHost = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]{0,253}$`)

// PingArgs はホスト名から ping 実行用の argv を組み立てる。
// シェルを介さず、入力検証後に引数配列のみを返す（OS コマンドインジェクション防止）。
func PingArgs(host string) ([]string, error) {
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}
	if !safeHost.MatchString(host) {
		return nil, fmt.Errorf("invalid host")
	}
	return []string{"ping", "-c", "1", host}, nil
}

// RunPing は PingArgs の結果を実行する（実験ハンドラ用）。
func RunPing(host string) ([]byte, error) {
	args, err := PingArgs(host)
	if err != nil {
		return nil, err
	}
	// args は PingArgs で許可リスト検証済みの固定バイナリ + 位置引数のみ。
	cmd := exec.Command(args[0], args[1:]...) // #nosec G204 -- host allowlist-validated argv, no shell
	return cmd.CombinedOutput()
}
