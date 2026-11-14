package member_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sync"
	"testing"
	"time"

	domain "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/member"
	usecase "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/usecase/member"
	"golang.org/x/crypto/argon2"
)

type memRepo struct {
	mu        sync.Mutex
	byID      map[string]*domain.Member
	byLogin   map[string]*domain.Member
	saveCount int
	saveErr   error
}

func newMemRepo() *memRepo {
	return &memRepo{byID: map[string]*domain.Member{}, byLogin: map[string]*domain.Member{}}
}

func (r *memRepo) Save(_ context.Context, m *domain.Member) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saveCount++
	if r.saveErr != nil {
		return r.saveErr
	}
	r.byID[m.ID()] = m
	r.byLogin[m.LoginID()] = m
	return nil
}

func (r *memRepo) FindByID(_ context.Context, id string) (*domain.Member, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.byID[id]
	if !ok {
		return nil, usecase.ErrNotFound
	}
	return m, nil
}

func (r *memRepo) FindActiveByLoginID(_ context.Context, loginID string) (*domain.Member, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.byLogin[loginID]
	if !ok || m.Status() != domain.StatusActive {
		return nil, usecase.ErrNotFound
	}
	return m, nil
}

func (r *memRepo) Update(_ context.Context, m *domain.Member) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[m.ID()] = m
	// rebuild login index
	r.byLogin = map[string]*domain.Member{}
	for _, v := range r.byID {
		r.byLogin[v.LoginID()] = v
	}
	return nil
}

type memChallenges struct {
	mu   sync.Mutex
	byID map[string]*usecase.Challenge
}

func newMemChallenges() *memChallenges {
	return &memChallenges{byID: map[string]*usecase.Challenge{}}
}

func (s *memChallenges) Save(_ context.Context, c *usecase.Challenge) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *c
	cp.Salt = append([]byte(nil), c.Salt...)
	cp.Challenge = append([]byte(nil), c.Challenge...)
	s.byID[c.ID] = &cp
	return nil
}

func (s *memChallenges) FindByID(_ context.Context, id string) (*usecase.Challenge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.byID[id]
	if !ok {
		return nil, usecase.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (s *memChallenges) MarkConsumed(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.byID[id]
	if !ok {
		return usecase.ErrNotFound
	}
	c.Consumed = true
	return nil
}

type memSessions struct {
	mu      sync.Mutex
	byToken map[string]*usecase.Session
	revoked map[string]bool
}

func newMemSessions() *memSessions {
	return &memSessions{byToken: map[string]*usecase.Session{}, revoked: map[string]bool{}}
}

func (s *memSessions) Save(_ context.Context, sess *usecase.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *sess
	s.byToken[sess.Token] = &cp
	return nil
}

func (s *memSessions) FindByToken(_ context.Context, token string) (*usecase.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.byToken[token]
	if !ok || s.revoked[sess.MemberID] {
		return nil, usecase.ErrNotFound
	}
	cp := *sess
	return &cp, nil
}

func (s *memSessions) RevokeAllForMember(_ context.Context, memberID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revoked[memberID] = true
	return nil
}

func derive(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 2, 19456, 1, 32)
}

func responseFor(secret, challenge []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(challenge)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// ケース: U-R-01 / 目的: 登録成功時に公開フィールドのみ返すこと。
func TestRegister_Success(t *testing.T) {
	t.Parallel()
	repo := newMemRepo()
	uc := usecase.NewUseCase(repo, newMemChallenges(), newMemSessions())

	got, err := uc.Register(context.Background(), "member01", "Abcd123!", "しろかぎ")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.LoginID != "member01" || got.Nickname != "しろかぎ" {
		t.Fatalf("got %+v", got)
	}
	if repo.saveCount != 1 {
		t.Fatalf("saveCount=%d", repo.saveCount)
	}
}

// ケース: U-R-02 / 目的: 無効パスワードではリポジトリを呼ばないこと。
func TestRegister_InvalidPasswordDoesNotSave(t *testing.T) {
	t.Parallel()
	repo := newMemRepo()
	uc := usecase.NewUseCase(repo, newMemChallenges(), newMemSessions())

	_, err := uc.Register(context.Background(), "member01", "short", "しろかぎ")
	if err == nil {
		t.Fatal("expected error")
	}
	if repo.saveCount != 0 {
		t.Fatalf("saveCount=%d", repo.saveCount)
	}
}

// ケース: U-R-03 / 目的: 重複を衝突エラーとして伝播すること。
func TestRegister_PropagatesConflict(t *testing.T) {
	t.Parallel()
	repo := newMemRepo()
	repo.saveErr = usecase.ErrConflict
	uc := usecase.NewUseCase(repo, newMemChallenges(), newMemSessions())

	got, err := uc.Register(context.Background(), "member01", "Abcd123!", "しろかぎ")
	if !errors.Is(err, usecase.ErrConflict) {
		t.Fatalf("err=%v", err)
	}
	if got != nil {
		t.Fatalf("got=%+v", got)
	}
}

// ケース: U-C-01 / 目的: 登録済み会員へ挑戦を発行できること。
func TestIssueChallenge_Registered(t *testing.T) {
	t.Parallel()
	repo := newMemRepo()
	ch := newMemChallenges()
	uc := usecase.NewUseCase(repo, ch, newMemSessions())
	if _, err := uc.Register(context.Background(), "member01", "Abcd123!", "しろかぎ"); err != nil {
		t.Fatalf("register: %v", err)
	}

	got, err := uc.IssueChallenge(context.Background(), "member01")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.ChallengeID == "" || got.Salt == "" || got.Challenge == "" {
		t.Fatalf("empty fields: %+v", got)
	}
}

// ケース: U-C-02 / 目的: 未登録でも成功形で返すこと。
func TestIssueChallenge_UnknownStillSucceeds(t *testing.T) {
	t.Parallel()
	uc := usecase.NewUseCase(newMemRepo(), newMemChallenges(), newMemSessions())

	got, err := uc.IssueChallenge(context.Background(), "unknown01")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.ChallengeID == "" || got.Salt == "" || got.Challenge == "" {
		t.Fatalf("empty fields: %+v", got)
	}
}

// ケース: U-C-03 / 目的: 短すぎる loginId を拒否すること。
func TestIssueChallenge_RejectsShortLoginID(t *testing.T) {
	t.Parallel()
	uc := usecase.NewUseCase(newMemRepo(), newMemChallenges(), newMemSessions())

	_, err := uc.IssueChallenge(context.Background(), "short")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ケース: U-V-01 / 目的: 正しい応答でセッションを発行すること。
func TestVerifyResponse_Success(t *testing.T) {
	t.Parallel()
	repo := newMemRepo()
	chStore := newMemChallenges()
	uc := usecase.NewUseCase(repo, chStore, newMemSessions())
	if _, err := uc.Register(context.Background(), "member01", "Abcd123!", "しろかぎ"); err != nil {
		t.Fatalf("register: %v", err)
	}
	chal, err := uc.IssueChallenge(context.Background(), "member01")
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	salt, _ := base64.StdEncoding.DecodeString(chal.Salt)
	challenge, _ := base64.StdEncoding.DecodeString(chal.Challenge)
	secret := derive("Abcd123!", salt)
	resp := responseFor(secret, challenge)

	got, err := uc.VerifyResponse(context.Background(), chal.ChallengeID, resp)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got.SessionToken == "" || got.Nickname != "しろかぎ" || got.ExpiresAt.IsZero() {
		t.Fatalf("got %+v", got)
	}
}

// ケース: U-V-02 / 目的: 誤った応答を認証失敗にすること。
func TestVerifyResponse_WrongResponse(t *testing.T) {
	t.Parallel()
	repo := newMemRepo()
	uc := usecase.NewUseCase(repo, newMemChallenges(), newMemSessions())
	if _, err := uc.Register(context.Background(), "member01", "Abcd123!", "しろかぎ"); err != nil {
		t.Fatalf("register: %v", err)
	}
	chal, err := uc.IssueChallenge(context.Background(), "member01")
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}

	_, err = uc.VerifyResponse(context.Background(), chal.ChallengeID, base64.StdEncoding.EncodeToString([]byte("nope")))
	if !errors.Is(err, usecase.ErrAuthFailed) {
		t.Fatalf("err=%v", err)
	}
}

// ケース: U-V-03 / 目的: 一度成功した挑戦の再検証を失敗させること。
func TestVerifyResponse_ConsumedChallenge(t *testing.T) {
	t.Parallel()
	repo := newMemRepo()
	uc := usecase.NewUseCase(repo, newMemChallenges(), newMemSessions())
	if _, err := uc.Register(context.Background(), "member01", "Abcd123!", "しろかぎ"); err != nil {
		t.Fatalf("register: %v", err)
	}
	chal, err := uc.IssueChallenge(context.Background(), "member01")
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	salt, _ := base64.StdEncoding.DecodeString(chal.Salt)
	challenge, _ := base64.StdEncoding.DecodeString(chal.Challenge)
	resp := responseFor(derive("Abcd123!", salt), challenge)

	if _, err := uc.VerifyResponse(context.Background(), chal.ChallengeID, resp); err != nil {
		t.Fatalf("first verify: %v", err)
	}
	_, err = uc.VerifyResponse(context.Background(), chal.ChallengeID, resp)
	if !errors.Is(err, usecase.ErrAuthFailed) {
		t.Fatalf("err=%v", err)
	}
}

// ケース: U-D-01 / 目的: 削除成功時にマスク保存とセッション無効化が行われること。
func TestDeleteMe_Success(t *testing.T) {
	t.Parallel()
	repo := newMemRepo()
	sessions := newMemSessions()
	uc := usecase.NewUseCase(repo, newMemChallenges(), sessions)
	if _, err := uc.Register(context.Background(), "member01", "Abcd123!", "しろかぎ"); err != nil {
		t.Fatalf("register: %v", err)
	}
	chal, _ := uc.IssueChallenge(context.Background(), "member01")
	salt, _ := base64.StdEncoding.DecodeString(chal.Salt)
	challenge, _ := base64.StdEncoding.DecodeString(chal.Challenge)
	ver, err := uc.VerifyResponse(context.Background(), chal.ChallengeID, responseFor(derive("Abcd123!", salt), challenge))
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if err := uc.DeleteMe(context.Background(), ver.SessionToken); err != nil {
		t.Fatalf("delete: %v", err)
	}
	id := firstID(repo)
	m, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if m.Status() != domain.StatusDeleted {
		t.Fatalf("status=%s", m.Status())
	}
	if m.LoginID() == "member01" {
		t.Fatal("loginId not masked")
	}
	if _, err := sessions.FindByToken(context.Background(), ver.SessionToken); err == nil {
		t.Fatal("expected session revoked")
	}
}

func firstID(repo *memRepo) string {
	for id := range repo.byID {
		return id
	}
	return ""
}

// ケース: U-D-02 / 目的: 削除済み会員の削除を失敗させること。
func TestDeleteMe_AlreadyDeleted(t *testing.T) {
	t.Parallel()
	repo := newMemRepo()
	sessions := newMemSessions()
	uc := usecase.NewUseCase(repo, newMemChallenges(), sessions)
	if _, err := uc.Register(context.Background(), "member01", "Abcd123!", "しろかぎ"); err != nil {
		t.Fatalf("register: %v", err)
	}
	chal, _ := uc.IssueChallenge(context.Background(), "member01")
	salt, _ := base64.StdEncoding.DecodeString(chal.Salt)
	challenge, _ := base64.StdEncoding.DecodeString(chal.Challenge)
	ver, err := uc.VerifyResponse(context.Background(), chal.ChallengeID, responseFor(derive("Abcd123!", salt), challenge))
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if err := uc.DeleteMe(context.Background(), ver.SessionToken); err != nil {
		t.Fatalf("first delete: %v", err)
	}

	// session still present in map but revoked — FindByToken fails → unauthorized
	// Use a fresh session entry pointing at deleted member to hit ErrNotFound path.
	id := firstID(repo)
	sessions.byToken["tok"] = &usecase.Session{
		Token:     "tok",
		MemberID:  id,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	sessions.revoked[id] = false

	err = uc.DeleteMe(context.Background(), "tok")
	if !errors.Is(err, usecase.ErrNotFound) {
		t.Fatalf("err=%v", err)
	}
}
