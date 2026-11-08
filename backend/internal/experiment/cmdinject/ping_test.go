package cmdinject_test

import (
	"strings"
	"testing"

	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/experiment/cmdinject"
)

// インジェクション用プローブは攻撃手順ではなく、シェルメタ文字の拒否漏れを検出する入力。
var injectionProbes = []struct {
	name    string
	input   string
	purpose string
}{
	{
		name:    "semicolon",
		input:   "127.0.0.1;id",
		purpose: "セミコロンによる追加コマンドが拒否されること",
	},
	{
		name:    "pipe",
		input:   "127.0.0.1|id",
		purpose: "パイプによる追加コマンドが拒否されること",
	},
	{
		name:    "and",
		input:   "127.0.0.1&&id",
		purpose: "&& による追加コマンドが拒否されること",
	},
	{
		name:    "substitution",
		input:   "$(id)",
		purpose: "コマンド置換が拒否されること",
	},
	{
		name:    "backtick",
		input:   "`id`",
		purpose: "バッククォート置換が拒否されること",
	},
}

// 目的: シェルを介さず argv で ping すること（OS コマンドインジェクション防止）。
func TestPingArgs_UsesArgvWithoutShell(t *testing.T) {
	t.Parallel()

	args, err := cmdinject.PingArgs("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"ping", "-c", "1", "example.com"}
	if len(args) != len(want) {
		t.Fatalf("args=%v, want %v (must not use sh -c)", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("args=%v, want %v (must not use sh -c)", args, want)
		}
	}
}

// 目的: シェルメタ文字を含む host を拒否すること。
func TestPingArgs_RejectsShellMetacharacters(t *testing.T) {
	t.Parallel()

	for _, tc := range injectionProbes {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			t.Log(tc.purpose)

			args, err := cmdinject.PingArgs(tc.input)
			if err == nil {
				joined := strings.Join(args, " ")
				t.Fatalf("accepted unsafe host %q (command injection risk); args=%v joined=%q", tc.input, args, joined)
			}
		})
	}
}
