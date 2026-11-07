package xssrender_test

import (
	"html"
	"strings"
	"testing"

	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/experiment/xssrender"
)

// XSS プローブは攻撃手順ではなく、エスケープ漏れを検出するためのテスト用入力。
var xssProbes = []struct {
	name    string
	input   string
	purpose string
}{
	{
		name:    "script_tag",
		input:   `<script>alert(1)</script>`,
		purpose: "script 要素が生のまま HTML に埋め込まれないこと",
	},
	{
		name:    "img_onerror",
		input:   `<img src=x onerror=alert(1)>`,
		purpose: "イベントハンドラ付きタグが生反射しないこと",
	},
	{
		name:    "attribute_breakout",
		input:   `" onclick="alert(1)`,
		purpose: "引用符による属性脱出がエスケープされること",
	},
}

// 目的: 信頼できない入力を HTML 埋め込み前にエスケープし、反射型 XSS を防ぐ。
func TestRenderGreeting_EscapesHTMLSpecialCharacters(t *testing.T) {
	t.Parallel()

	for _, tc := range xssProbes {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			t.Log(tc.purpose)

			out := xssrender.RenderGreeting(tc.input)

			if strings.Contains(out, tc.input) {
				t.Fatalf("raw user input reflected in HTML (XSS risk):\ninput=%q\nout=%q", tc.input, out)
			}

			want := "<h1>Hello, " + html.EscapeString(tc.input) + "</h1>"
			if out != want {
				t.Fatalf("escaped output mismatch\ngot:  %q\nwant: %q", out, want)
			}
		})
	}
}

// 目的: HTML 特殊文字を含まない平文は、意味を変えずにそのまま描画できること。
func TestRenderGreeting_SafePlainTextStillRenders(t *testing.T) {
	t.Parallel()

	out := xssrender.RenderGreeting("しろかぎ")
	want := "<h1>Hello, しろかぎ</h1>"
	if out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}
