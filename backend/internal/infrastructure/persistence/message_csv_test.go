package persistence_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	domain "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/message"
	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/infrastructure/persistence"
)

func writeCSV(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	return path
}

// 目的: CSV から ID 一致のメッセージを取得できること。
func TestMessageCSVRepository_FindByID_Success(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := writeCSV(t, dir, "messages.csv", "id,message\n1,Hello World\n")
	repo := persistence.NewMessageCSVRepository(path)

	msg, err := repo.FindByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content() != "Hello World" {
		t.Fatalf("content = %q", msg.Content())
	}
}

// 目的: 存在しない ID はエラー（未発見）になること。
func TestMessageCSVRepository_FindByID_NotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := writeCSV(t, dir, "messages.csv", "id,message\n1,Hello World\n")
	repo := persistence.NewMessageCSVRepository(path)

	msg, err := repo.FindByID(context.Background(), "999")
	if err == nil {
		t.Fatal("expected error")
	}
	if msg != nil {
		t.Fatalf("expected nil, got %#v", msg)
	}
}

// 目的: 特殊な ID 文字列を解釈せず、一致しなければ not found で安全に終えること。
func TestMessageCSVRepository_FindByID_MaliciousIDDoesNotBreakLookup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := writeCSV(t, dir, "messages.csv", "id,message\n1,Hello World\n")
	repo := persistence.NewMessageCSVRepository(path)

	payloads := []domain.ID{
		`1 OR 1=1`,
		`1; DROP TABLE messages;--`,
		`../etc/passwd`,
		domain.ID(strings.Repeat("9", 4096)),
		"1\x00hidden",
	}

	for _, id := range payloads {
		id := id
		t.Run(string(id)[:min(24, len(id))], func(t *testing.T) {
			t.Parallel()
			msg, err := repo.FindByID(context.Background(), id)
			if err == nil {
				t.Fatal("expected not found error for non-matching opaque id")
			}
			if msg != nil {
				t.Fatalf("expected nil message, got %#v", msg)
			}
		})
	}
}

// 目的: キャンセル済み context を検出して処理を中断すること。
func TestMessageCSVRepository_FindByID_ContextCanceled(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := writeCSV(t, dir, "messages.csv", "id,message\n1,Hello World\n")
	repo := persistence.NewMessageCSVRepository(path)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.FindByID(ctx, "1")
	if err == nil {
		t.Fatal("expected context error")
	}
}

// 目的: CSV ファイル不在時にエラーを返すこと。
func TestMessageCSVRepository_FindByID_MissingFile(t *testing.T) {
	t.Parallel()

	repo := persistence.NewMessageCSVRepository(filepath.Join(t.TempDir(), "missing.csv"))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := repo.FindByID(ctx, "1")
	if err == nil {
		t.Fatal("expected open error")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
