package message_test

import (
	"strings"
	"testing"

	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/message"
)

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

func TestNewMessage_RejectsEmptyValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		id      message.ID
		content string
	}{
		{name: "empty id", id: "", content: "Hello"},
		{name: "empty content", id: "1", content: ""},
		{name: "both empty", id: "", content: ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
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

func TestNewMessage_AcceptsSpecialCharactersAsData(t *testing.T) {
	t.Parallel()

	// Domain treats these as opaque data; they must not be rejected as "code".
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
