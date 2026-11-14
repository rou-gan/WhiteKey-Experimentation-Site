package member_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/member"
)

// ケース: D-N-01 / 目的: 妥当な値で active 会員を生成できること。
func TestNewActiveMember_Success(t *testing.T) {
	t.Parallel()

	m, err := member.NewActiveMember("member01", "Abcd123!", "しろかぎ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Status() != member.StatusActive {
		t.Fatalf("status = %q, want active", m.Status())
	}
	if m.LoginID() != "member01" {
		t.Fatalf("loginId = %q, want member01", m.LoginID())
	}
	if m.Nickname() != "しろかぎ" {
		t.Fatalf("nickname = %q, want しろかぎ", m.Nickname())
	}
	if m.PendingPassword() != "Abcd123!" {
		t.Fatalf("pending password mismatch")
	}
	if m.ID() == "" {
		t.Fatal("expected non-empty id")
	}
}

// ケース: D-N-02 / 目的: 7 文字のログインIDを拒否すること。
func TestNewActiveMember_RejectsShortLoginID(t *testing.T) {
	t.Parallel()

	m, err := member.NewActiveMember("member0", "Abcd123!", "しろかぎ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if m != nil {
		t.Fatalf("expected nil member, got %#v", m)
	}
}

// ケース: D-N-03 / 目的: 記号を含まないパスワードを拒否すること。
func TestNewActiveMember_RejectsPasswordWithoutSymbol(t *testing.T) {
	t.Parallel()

	m, err := member.NewActiveMember("member01", "Abcd1234", "しろかぎ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if m != nil {
		t.Fatalf("expected nil member, got %#v", m)
	}
}

// ケース: D-N-04 / 目的: 51 ルーンのニックネームを拒否すること。
func TestNewActiveMember_RejectsLongNickname(t *testing.T) {
	t.Parallel()

	nick := strings.Repeat("あ", 51)
	if utf8.RuneCountInString(nick) != 51 {
		t.Fatalf("fixture rune count = %d", utf8.RuneCountInString(nick))
	}
	m, err := member.NewActiveMember("member01", "Abcd123!", nick)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if m != nil {
		t.Fatalf("expected nil member, got %#v", m)
	}
}

// ケース: D-M-01, D-M-02 / 目的: マスク後に元値が残らず状態が deleted であること。
func TestMaskForLogicalDelete(t *testing.T) {
	t.Parallel()

	m, err := member.NewActiveMember("member01", "Abcd123!", "しろかぎ")
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := m.SetCredentials([]byte("salt-value-16b!"), []byte("secret-value-32-bytes-padded!!")); err != nil {
		t.Fatalf("set credentials: %v", err)
	}
	origLogin := m.LoginID()
	origNick := m.Nickname()
	origSalt := string(m.Salt())
	origSecret := string(m.Secret())

	if err := m.MaskForLogicalDelete(); err != nil {
		t.Fatalf("mask: %v", err)
	}
	if m.Status() != member.StatusDeleted {
		t.Fatalf("status = %q, want deleted", m.Status())
	}
	if m.LoginID() == origLogin {
		t.Fatal("loginId was not masked")
	}
	if m.Nickname() == origNick {
		t.Fatal("nickname was not masked")
	}
	if string(m.Salt()) == origSalt {
		t.Fatal("salt was not replaced")
	}
	if string(m.Secret()) == origSecret {
		t.Fatal("secret was not replaced")
	}
	if m.PendingPassword() != "" {
		t.Fatal("pending password should be cleared")
	}
	if m.DeletedAt() == nil {
		t.Fatal("expected deletedAt")
	}
}
