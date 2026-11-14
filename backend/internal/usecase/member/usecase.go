package member

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	domain "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/member"
)

const (
	challengeTTL = 5 * time.Minute
	sessionTTL   = 8 * time.Hour
)

// UseCase orchestrates member registration, challenge-response auth, and deletion.
type UseCase struct {
	repo      domain.Repository
	challenges ChallengeStore
	sessions  SessionStore
	now       func() time.Time
}

// NewUseCase wires dependencies.
func NewUseCase(repo domain.Repository, challenges ChallengeStore, sessions SessionStore) *UseCase {
	return &UseCase{
		repo:       repo,
		challenges: challenges,
		sessions:   sessions,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

// RegisterResult is the public registration response.
type RegisterResult struct {
	LoginID  string
	Nickname string
}

// Register creates an active member (EDS F-01).
func (uc *UseCase) Register(ctx context.Context, loginID, password, nickname string) (*RegisterResult, error) {
	m, err := domain.NewActiveMember(loginID, password, nickname)
	if err != nil {
		return nil, err
	}
	salt, err := newSalt()
	if err != nil {
		return nil, err
	}
	secret := deriveSecret(m.PendingPassword(), salt)
	if err := m.SetCredentials(salt, secret); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, m); err != nil {
		if errors.Is(err, ErrConflict) {
			return nil, ErrConflict
		}
		return nil, err
	}
	return &RegisterResult{LoginID: m.LoginID(), Nickname: m.Nickname()}, nil
}

// ChallengeResult is returned to clients for response computation.
type ChallengeResult struct {
	ChallengeID string
	Salt        string // Base64
	Challenge   string // Base64
}

// IssueChallenge issues a challenge (EDS F-02). Unknown login IDs get dummy values.
func (uc *UseCase) IssueChallenge(ctx context.Context, loginID string) (*ChallengeResult, error) {
	if err := domain.ValidateLoginID(loginID); err != nil {
		return nil, err
	}
	id, err := newChallengeID()
	if err != nil {
		return nil, err
	}
	challengeBytes, err := newChallengeBytes()
	if err != nil {
		return nil, err
	}

	var (
		salt     []byte
		memberID string
	)
	m, err := uc.repo.FindActiveByLoginID(ctx, loginID)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return nil, err
		}
		salt, err = newSalt()
		if err != nil {
			return nil, err
		}
	} else {
		salt = m.Salt()
		memberID = m.ID()
	}

	expires := uc.now().Add(challengeTTL)
	ch := &Challenge{
		ID:        id,
		LoginID:   loginID,
		MemberID:  memberID,
		Salt:      salt,
		Challenge: challengeBytes,
		ExpiresAt: expires,
	}
	if err := uc.challenges.Save(ctx, ch); err != nil {
		return nil, err
	}
	return &ChallengeResult{
		ChallengeID: id,
		Salt:        base64.StdEncoding.EncodeToString(salt),
		Challenge:   base64.StdEncoding.EncodeToString(challengeBytes),
	}, nil
}

// VerifyResult is returned on successful authentication.
type VerifyResult struct {
	SessionToken string
	Nickname     string
	ExpiresAt    time.Time
}

// VerifyResponse validates a challenge response (EDS F-03).
func (uc *UseCase) VerifyResponse(ctx context.Context, challengeID, responseB64 string) (*VerifyResult, error) {
	fail := func() (*VerifyResult, error) { return nil, ErrAuthFailed }

	ch, err := uc.challenges.FindByID(ctx, challengeID)
	if err != nil || ch == nil {
		return fail()
	}
	if ch.Consumed || uc.now().After(ch.ExpiresAt) || ch.MemberID == "" {
		return fail()
	}
	m, err := uc.repo.FindByID(ctx, ch.MemberID)
	if err != nil || m == nil || m.Status() != domain.StatusActive {
		return fail()
	}
	if !responsesEqual(responseB64, m.Secret(), ch.Challenge) {
		return fail()
	}
	if err := uc.challenges.MarkConsumed(ctx, ch.ID); err != nil {
		return fail()
	}
	token, hash, err := newSessionToken()
	if err != nil {
		return nil, err
	}
	expires := uc.now().Add(sessionTTL)
	sess := &Session{
		Token:     token,
		TokenHash: hash,
		MemberID:  m.ID(),
		ExpiresAt: expires,
	}
	if err := uc.sessions.Save(ctx, sess); err != nil {
		return nil, err
	}
	return &VerifyResult{
		SessionToken: token,
		Nickname:     m.Nickname(),
		ExpiresAt:    expires,
	}, nil
}

// DeleteMe logically deletes the member bound to the session (EDS F-04).
func (uc *UseCase) DeleteMe(ctx context.Context, sessionToken string) error {
	sess, err := uc.sessions.FindByToken(ctx, sessionToken)
	if err != nil || sess == nil {
		return ErrUnauthorized
	}
	if uc.now().After(sess.ExpiresAt) {
		return ErrUnauthorized
	}
	m, err := uc.repo.FindByID(ctx, sess.MemberID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	if m.Status() != domain.StatusActive {
		return ErrNotFound
	}
	if err := m.MaskForLogicalDelete(); err != nil {
		return err
	}
	if err := uc.repo.Update(ctx, m); err != nil {
		return err
	}
	return uc.sessions.RevokeAllForMember(ctx, m.ID())
}
