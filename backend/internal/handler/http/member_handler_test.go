package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	handlerhttp "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/handler/http"
	memberuc "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/usecase/member"
)

type stubMemberUC struct {
	registerRes  *memberuc.RegisterResult
	registerErr  error
	challengeRes *memberuc.ChallengeResult
	challengeErr error
	verifyRes    *memberuc.VerifyResult
	verifyErr    error
	deleteErr    error
}

func (s *stubMemberUC) Register(_ context.Context, _, _, _ string) (*memberuc.RegisterResult, error) {
	return s.registerRes, s.registerErr
}

func (s *stubMemberUC) IssueChallenge(_ context.Context, _ string) (*memberuc.ChallengeResult, error) {
	return s.challengeRes, s.challengeErr
}

func (s *stubMemberUC) VerifyResponse(_ context.Context, _, _ string) (*memberuc.VerifyResult, error) {
	return s.verifyRes, s.verifyErr
}

func (s *stubMemberUC) DeleteMe(_ context.Context, _ string) error {
	return s.deleteErr
}

// ケース: H-01 / 目的: POST /members 成功が 201 になること。
func TestMemberHandler_Register_Success(t *testing.T) {
	t.Parallel()

	h := handlerhttp.NewMemberHandler(&stubMemberUC{
		registerRes: &memberuc.RegisterResult{LoginID: "member01", Nickname: "しろかぎ"},
	})

	req := httptest.NewRequest(http.MethodPost, "/members", strings.NewReader(
		`{"loginId":"member01","password":"Abcd123!","nickname":"しろかぎ"}`,
	))
	rec := httptest.NewRecorder()
	h.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body["loginId"] != "member01" || body["nickname"] != "しろかぎ" {
		t.Fatalf("body = %#v", body)
	}
	if _, ok := body["password"]; ok {
		t.Fatalf("password must not appear: %#v", body)
	}
}

// ケース: H-02 / 目的: 登録衝突が 409 汎用になること。
func TestMemberHandler_Register_ConflictGeneric(t *testing.T) {
	t.Parallel()

	h := handlerhttp.NewMemberHandler(&stubMemberUC{registerErr: memberuc.ErrConflict})

	req := httptest.NewRequest(http.MethodPost, "/members", strings.NewReader(
		`{"loginId":"member01","password":"Abcd123!","nickname":"しろかぎ"}`,
	))
	rec := httptest.NewRecorder()
	h.Register(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body["message"] == "" {
		t.Fatalf("expected generic message, got %#v", body)
	}
	msg := strings.ToLower(body["message"])
	if strings.Contains(msg, "login") || strings.Contains(msg, "nick") {
		t.Fatalf("conflict message distinguishes field: %q", body["message"])
	}
}

// ケース: H-03 / 目的: POST /members/auth/challenge 成功が 200 になること。
func TestMemberHandler_IssueChallenge_Success(t *testing.T) {
	t.Parallel()

	h := handlerhttp.NewMemberHandler(&stubMemberUC{
		challengeRes: &memberuc.ChallengeResult{
			ChallengeID: "chal_1",
			Salt:        "c2FsdA==",
			Challenge:   "Y2hhbA==",
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/members/auth/challenge", strings.NewReader(
		`{"loginId":"member01"}`,
	))
	rec := httptest.NewRecorder()
	h.IssueChallenge(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body["challengeId"] == "" || body["salt"] == "" || body["challenge"] == "" {
		t.Fatalf("body = %#v", body)
	}
}

// ケース: H-04 / 目的: 検証失敗が 401 汎用になること。
func TestMemberHandler_VerifyResponse_AuthFailedGeneric(t *testing.T) {
	t.Parallel()

	h := handlerhttp.NewMemberHandler(&stubMemberUC{verifyErr: memberuc.ErrAuthFailed})

	req := httptest.NewRequest(http.MethodPost, "/members/auth/verify", strings.NewReader(
		`{"challengeId":"chal_1","response":"bad"}`,
	))
	rec := httptest.NewRecorder()
	h.VerifyResponse(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}

	raw := strings.ToLower(rec.Body.String())
	for _, reason := range []string{"expired", "mismatch", "consumed", "not found", "dummy", "password"} {
		if strings.Contains(raw, reason) {
			t.Fatalf("401 body distinguishes failure reason %q: %q", reason, rec.Body.String())
		}
	}
	if !strings.Contains(rec.Body.String(), `"message"`) {
		t.Fatalf("expected generic JSON message, got %q", rec.Body.String())
	}
}

// ケース: H-05 / 目的: DELETE /members/me 成功が 204 になること。
func TestMemberHandler_DeleteMe_Success(t *testing.T) {
	t.Parallel()

	h := handlerhttp.NewMemberHandler(&stubMemberUC{})

	req := httptest.NewRequest(http.MethodDelete, "/members/me", nil)
	req.Header.Set("Authorization", "Bearer sess-token")
	rec := httptest.NewRecorder()
	h.DeleteMe(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body should be empty, got %q", rec.Body.String())
	}
}
