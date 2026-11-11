package message_test

import (
	"strings"
	"testing"

	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/message"
)

// ケース: D-01 / 目的: 妥当な ID・本文で Message を生成できること。
func TestNewMessage_Success(t *testing.T) {
	t.Parallel()

	msg, err := message.NewMessage("1", "Hello World")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID() != "1" {
		t.Fatalf("id = %q, want 1", msg.ID())
	}
	if msg.Content() != "Hello World" {
		t.Fatalf("content = %q, want Hello World", msg.Content())
	}
}

// ケース: D-02, D-03 / 目的: 空の ID または本文をドメイン層で拒否すること。
func TestNewMessage_RejectsEmptyValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		id      message.ID
		content string
		purpose string
	}{
		{name: "empty id", id: "", content: "Hello", purpose: "D-02: 空 ID を拒否する"},
		{name: "empty content", id: "1", content: "", purpose: "D-03: 空本文を拒否する"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			t.Log(tc.purpose)
			msg, err := message.NewMessage(tc.id, tc.content)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if msg != nil {
				t.Fatalf("expected nil message, got %#v", msg)
			}
		})
	}
}

// ケース: D-04 / 目的: SQL/XSS 風の文字列も「コード」として弾かず、不透明なデータとして保持すること。
func TestNewMessage_AcceptsSpecialCharactersAsData(t *testing.T) {
	t.Parallel()

	payloads := []string{
		"Hello'; DROP TABLE messages;--",
		`<script>alert("xss")</script>`,
		"line1\nline2\r\n",
		strings.Repeat("A", 10_000),
	}

	for _, content := range payloads {
		content := content
		t.Run(content[:min(32, len(content))], func(t *testing.T) {
			t.Parallel()
			msg, err := message.NewMessage("1", content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if msg.Content() != content {
				t.Fatalf("content mismatch")
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
