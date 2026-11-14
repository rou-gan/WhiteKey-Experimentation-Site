package persistence_test

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"testing"

	domain "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/member"
	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/infrastructure/persistence"
	memberuc "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/usecase/member"
)

func newTestMember(t *testing.T, loginID, nickname string) *domain.Member {
	t.Helper()
	m, err := domain.NewActiveMember(loginID, "Abcd123!", nickname)
	if err != nil {
		t.Fatalf("NewActiveMember: %v", err)
	}
	salt := make([]byte, 16)
	secret := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		t.Fatalf("salt: %v", err)
	}
	if _, err := rand.Read(secret); err != nil {
		t.Fatalf("secret: %v", err)
	}
	if err := m.SetCredentials(salt, secret); err != nil {
		t.Fatalf("SetCredentials: %v", err)
	}
	return m
}

func uniqueLogin(t *testing.T) string {
	t.Helper()
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("rand: %v", err)
	}
	// 8〜15 英数字
	return fmt.Sprintf("mem%02x%02x%02x", b[0], b[1], b[2])
}

// ケース: I-01 / 目的: 保存した会員を loginId で取得できること。
func TestMemberPostgresRepository_FindActiveByLoginID_Success(t *testing.T) {
	db := openTestDB(t)
	repo := persistence.NewMemberPostgresRepository(db)

	loginID := uniqueLogin(t)
	m := newTestMember(t, loginID, "nick"+loginID)
	if err := repo.Save(context.Background(), m); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindActiveByLoginID(context.Background(), loginID)
	if err != nil {
		t.Fatalf("FindActiveByLoginID: %v", err)
	}
	if got.Status() != domain.StatusActive {
		t.Fatalf("status=%s, want active", got.Status())
	}
	if got.LoginID() != loginID {
		t.Fatalf("loginID=%q, want %q", got.LoginID(), loginID)
	}
}

// ケース: I-02 / 目的: マスク削除後に旧 loginId で active 取得できないこと。
func TestMemberPostgresRepository_FindActiveByLoginID_AfterSoftDelete(t *testing.T) {
	db := openTestDB(t)
	repo := persistence.NewMemberPostgresRepository(db)

	loginID := uniqueLogin(t)
	m := newTestMember(t, loginID, "nick"+loginID)
	if err := repo.Save(context.Background(), m); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := m.MaskForLogicalDelete(); err != nil {
		t.Fatalf("MaskForLogicalDelete: %v", err)
	}
	if err := repo.Update(context.Background(), m); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.FindActiveByLoginID(context.Background(), loginID)
	if !errors.Is(err, memberuc.ErrNotFound) {
		t.Fatalf("err=%v, want ErrNotFound", err)
	}
	if got != nil {
		t.Fatalf("got=%v, want nil", got)
	}
}

// ケース: I-03 / 目的: 同じ loginId の二重保存が失敗すること。
func TestMemberPostgresRepository_Save_DuplicateLoginID(t *testing.T) {
	db := openTestDB(t)
	repo := persistence.NewMemberPostgresRepository(db)

	loginID := uniqueLogin(t)
	first := newTestMember(t, loginID, "nickA"+loginID[3:])
	if err := repo.Save(context.Background(), first); err != nil {
		t.Fatalf("Save first: %v", err)
	}

	second := newTestMember(t, loginID, "nickB"+loginID[3:])
	err := repo.Save(context.Background(), second)
	if !errors.Is(err, memberuc.ErrConflict) {
		t.Fatalf("err=%v, want ErrConflict", err)
	}
}
