package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domain "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/message"
	handlerhttp "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/handler/http"
	messageuc "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/usecase/message"
)

type stubRepo struct {
	msg *domain.Message
	err error
}

func (s *stubRepo) FindByID(_ context.Context, _ domain.ID) (*domain.Message, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.msg, nil
}

func newHandler(t *testing.T, repo *stubRepo) *handlerhttp.MessageHandler {
	t.Helper()
	return handlerhttp.NewMessageHandler(messageuc.NewUseCase(repo))
}

// 目的: 正常系で JSON メッセージと Content-Type が返ること。
func TestMessageHandler_GetRoot_Success(t *testing.T) {
	t.Parallel()

	msg, err := domain.NewMessage("1", "Hello World")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	h := newHandler(t, &stubRepo{msg: msg})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.GetRoot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("content-type = %q", ct)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body["message"] != "Hello World" {
		t.Fatalf("body = %#v", body)
	}
}

// 目的: GET 以外のメソッドを 405 で拒否すること。
func TestMessageHandler_GetRoot_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	h := newHandler(t, &stubRepo{})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	h.GetRoot(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

// 目的: 内部エラー詳細（パス・秘密・スタック等）がレスポンスへ漏洩しないこと。
func TestMessageHandler_GetRoot_ErrorDoesNotLeakInternals(t *testing.T) {
	t.Parallel()

	internal := errors.New(`open csv: open /app/data/messages.csv: secret=super-secret stack=goroutine 1 [running]`)
	h := newHandler(t, &stubRepo{err: internal})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.GetRoot(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}

	body := rec.Body.String()
	forbidden := []string{
		"/app/data/messages.csv",
		"super-secret",
		"goroutine",
		"open csv",
		"stack",
	}
	for _, s := range forbidden {
		if strings.Contains(body, s) {
			t.Fatalf("response leaked %q: %q", s, body)
		}
	}
	if !strings.Contains(body, "internal server error") {
		t.Fatalf("expected generic error message, got %q", body)
	}
}

// 目的: 将来の認証・認可拒否テストの置き場を確保する（PoC では未実装のため Skip）。
func TestMessageHandler_AuthRejection_Placeholder(t *testing.T) {
	t.Skip("auth/authorization is not implemented in PoC; add mock-based deny tests when introduced")
}
