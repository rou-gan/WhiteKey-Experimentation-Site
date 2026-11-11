package message_test

import (
	"context"
	"errors"
	"testing"

	domain "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/message"
	usecase "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/usecase/message"
)

type stubRepo struct {
	msg *domain.Message
	err error
	got domain.ID
}

func (s *stubRepo) FindByID(_ context.Context, id domain.ID) (*domain.Message, error) {
	s.got = id
	if s.err != nil {
		return nil, s.err
	}
	return s.msg, nil
}

// ケース: U-01 / 目的: リポジトリから取得したメッセージをユースケースがそのまま返すこと。
func TestUseCase_GetByID_Success(t *testing.T) {
	t.Parallel()

	want, err := domain.NewMessage("1", "Hello World")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	repo := &stubRepo{msg: want}
	uc := usecase.NewUseCase(repo)

	got, err := uc.GetByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.got != "1" {
		t.Fatalf("repo got id %q, want 1", repo.got)
	}
	if got.Content() != "Hello World" {
		t.Fatalf("content = %q", got.Content())
	}
}

// ケース: U-02 / 目的: リポジトリエラーを握りつぶさず呼び出し元へ伝播すること。
func TestUseCase_GetByID_PropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	repo := &stubRepo{err: errors.New("not found")}
	uc := usecase.NewUseCase(repo)

	got, err := uc.GetByID(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if got != nil {
		t.Fatalf("expected nil message, got %#v", got)
	}
}

// ケース: U-03 / 目的: 悪意のある ID を解釈・改変せず、不透明値としてリポジトリへ転送すること。
func TestUseCase_GetByID_ForwardsMaliciousIDAsOpaqueValue(t *testing.T) {
	t.Parallel()

	malicious := `1'; DROP TABLE messages;--`
	repo := &stubRepo{err: errors.New("not found")}
	uc := usecase.NewUseCase(repo)

	_, _ = uc.GetByID(context.Background(), malicious)
	if string(repo.got) != malicious {
		t.Fatalf("repo got %q, want opaque malicious id unchanged", repo.got)
	}
}
