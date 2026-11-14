package persistence

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"

	memberuc "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/usecase/member"
)

type memberSessionPostgresStore struct {
	db *sql.DB
}

// NewMemberSessionPostgresStore persists sessions in Postgres.
func NewMemberSessionPostgresStore(db *sql.DB) memberuc.SessionStore {
	return &memberSessionPostgresStore{db: db}
}

func (s *memberSessionPostgresStore) Save(ctx context.Context, sess *memberuc.Session) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO member_sessions (token_hash, member_id, expires_at, revoked_at, created_at)
		VALUES ($1, $2, $3, NULL, now())
	`, sess.TokenHash, sess.MemberID, sess.ExpiresAt.UTC())
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (s *memberSessionPostgresStore) FindByToken(ctx context.Context, token string) (*memberuc.Session, error) {
	sum := sha256.Sum256([]byte(token))
	hash := sum[:]

	var (
		memberID  string
		expiresAt sql.NullTime
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT member_id, expires_at
		FROM member_sessions
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, hash).Scan(&memberID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, memberuc.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query session: %w", err)
	}
	return &memberuc.Session{
		Token:     token,
		TokenHash: append([]byte(nil), hash...),
		MemberID:  memberID,
		ExpiresAt: expiresAt.Time.UTC(),
	}, nil
}

func (s *memberSessionPostgresStore) RevokeAllForMember(ctx context.Context, memberID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE member_sessions
		SET revoked_at = now()
		WHERE member_id = $1 AND revoked_at IS NULL
	`, memberID)
	if err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}
	return nil
}
