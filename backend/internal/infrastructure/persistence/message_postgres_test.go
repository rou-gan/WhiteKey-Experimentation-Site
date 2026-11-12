package persistence_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	domain "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/message"
	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/infrastructure/persistence"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	if err := persistence.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// ケース: I-01 / 目的: DB から ID 一致のメッセージを取得できること。
func TestMessagePostgresRepository_FindByID_Success(t *testing.T) {
	db := openTestDB(t)
	repo := persistence.NewMessagePostgresRepository(db)

	msg, err := repo.FindByID(context.Background(), domain.ID("1"))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if msg.Content() != "Hello World" {
		t.Fatalf("content=%q, want Hello World", msg.Content())
	}
}

// ケース: I-02 / 目的: 存在しない ID は未発見エラーになること。
func TestMessagePostgresRepository_FindByID_NotFound(t *testing.T) {
	db := openTestDB(t)
	repo := persistence.NewMessagePostgresRepository(db)

	msg, err := repo.FindByID(context.Background(), domain.ID("999"))
	if err == nil {
		t.Fatal("expected error")
	}
	if msg != nil {
		t.Fatalf("msg=%v, want nil", msg)
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("err=%v, want not found", err)
	}
}

// ケース: I-03 / 目的: 特殊な ID を解釈せず一致しなければ not found で終えること。
func TestMessagePostgresRepository_FindByID_MaliciousIDDoesNotBreakLookup(t *testing.T) {
	db := openTestDB(t)
	repo := persistence.NewMessagePostgresRepository(db)

	ids := []domain.ID{
		"1 OR 1=1",
		"1; DROP TABLE messages;--",
		"' OR ''='",
	}
	for _, id := range ids {
		msg, err := repo.FindByID(context.Background(), id)
		if err == nil {
			t.Fatalf("id=%q: expected error", id)
		}
		if msg != nil {
			t.Fatalf("id=%q: msg=%v, want nil", id, msg)
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Fatalf("id=%q: err=%v, want not found", id, err)
		}
	}
}

// ケース: I-04 / 目的: キャンセル済み context を検出して中断すること。
func TestMessagePostgresRepository_FindByID_ContextCanceled(t *testing.T) {
	db := openTestDB(t)
	repo := persistence.NewMessagePostgresRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	msg, err := repo.FindByID(ctx, domain.ID("1"))
	if err == nil {
		t.Fatal("expected error")
	}
	if msg != nil {
		t.Fatalf("msg=%v, want nil", msg)
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("err=%v, want context canceled", err)
	}
}
