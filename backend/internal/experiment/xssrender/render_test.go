package xssrender_test

import (
	"html"
	"strings"
	"testing"

	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/experiment/xssrender"
)

// Probe strings are security-test fixtures (not attack instructions).
// They check that untrusted input is escaped before HTML embedding.
var xssProbes = []struct {
	name  string
	input string
}{
	{
		name:  "script_tag",
		input: `<script>alert(1)</script>`,
	},
	{
		name:  "img_onerror",
		input: `<img src=x onerror=alert(1)>`,
	},
	{
		name:  "attribute_breakout",
		input: `" onclick="alert(1)`,
	},
}

func TestRenderGreeting_EscapesHTMLSpecialCharacters(t *testing.T) {
	t.Parallel()

	for _, tc := range xssProbes {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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

func TestRenderGreeting_SafePlainTextStillRenders(t *testing.T) {
	t.Parallel()

	out := xssrender.RenderGreeting("しろかぎ")
	want := "<h1>Hello, しろかぎ</h1>"
	if out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}
