package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	memberuc "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/usecase/member"
)

type memberChallengePostgresStore struct {
	db *sql.DB
}

// NewMemberChallengePostgresStore persists auth challenges in Postgres.
func NewMemberChallengePostgresStore(db *sql.DB) memberuc.ChallengeStore {
	return &memberChallengePostgresStore{db: db}
}

func (s *memberChallengePostgresStore) Save(ctx context.Context, c *memberuc.Challenge) error {
	var memberID any
	if c.MemberID != "" {
		memberID = c.MemberID
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO member_auth_challenges (id, login_id, member_id, salt, challenge, expires_at, consumed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NULL, now())
	`, c.ID, c.LoginID, memberID, c.Salt, c.Challenge, c.ExpiresAt.UTC())
	if err != nil {
		return fmt.Errorf("insert challenge: %w", err)
	}
	return nil
}

func (s *memberChallengePostgresStore) FindByID(ctx context.Context, id string) (*memberuc.Challenge, error) {
	var (
		loginID               string
		memberID              sql.NullString
		salt, challenge       []byte
		expiresAt             sql.NullTime
		consumedAt            sql.NullTime
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT login_id, member_id, salt, challenge, expires_at, consumed_at
		FROM member_auth_challenges WHERE id = $1
	`, id).Scan(&loginID, &memberID, &salt, &challenge, &expiresAt, &consumedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, memberuc.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query challenge: %w", err)
	}
	ch := &memberuc.Challenge{
		ID:        id,
		LoginID:   loginID,
		Salt:      append([]byte(nil), salt...),
		Challenge: append([]byte(nil), challenge...),
		ExpiresAt: expiresAt.Time.UTC(),
		Consumed:  consumedAt.Valid,
	}
	if memberID.Valid {
		ch.MemberID = memberID.String
	}
	return ch, nil
}

func (s *memberChallengePostgresStore) MarkConsumed(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE member_auth_challenges
		SET consumed_at = now()
		WHERE id = $1 AND consumed_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("consume challenge: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("consume challenge rows: %w", err)
	}
	if n == 0 {
		return memberuc.ErrNotFound
	}
	return nil
}
