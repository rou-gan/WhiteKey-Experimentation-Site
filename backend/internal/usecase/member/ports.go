package member

import (
	"context"
	"time"
)

// Challenge is a one-time auth challenge.
type Challenge struct {
	ID        string
	LoginID   string
	MemberID  string // empty when dummy
	Salt      []byte
	Challenge []byte
	ExpiresAt time.Time
	Consumed  bool
}

// ChallengeStore persists auth challenges.
type ChallengeStore interface {
	Save(ctx context.Context, c *Challenge) error
	FindByID(ctx context.Context, id string) (*Challenge, error)
	MarkConsumed(ctx context.Context, id string) error
}

// Session is an authenticated session.
type Session struct {
	Token     string
	TokenHash []byte
	MemberID  string
	ExpiresAt time.Time
}

// SessionStore persists sessions.
type SessionStore interface {
	Save(ctx context.Context, s *Session) error
	FindByToken(ctx context.Context, token string) (*Session, error)
	RevokeAllForMember(ctx context.Context, memberID string) error
}
