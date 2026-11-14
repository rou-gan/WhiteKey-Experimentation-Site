package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	domain "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/member"
	memberuc "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/usecase/member"
)

type memberPostgresRepository struct {
	db *sql.DB
}

// NewMemberPostgresRepository persists members in Postgres.
func NewMemberPostgresRepository(db *sql.DB) domain.Repository {
	return &memberPostgresRepository{db: db}
}

func (r *memberPostgresRepository) Save(ctx context.Context, m *domain.Member) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO members (id, login_id, nickname, salt, secret, status, deleted_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
	`, m.ID(), m.LoginID(), m.Nickname(), m.Salt(), m.Secret(), string(m.Status()), nullTime(m.DeletedAt()))
	if err != nil {
		if isUniqueViolation(err) {
			return memberuc.ErrConflict
		}
		return fmt.Errorf("insert member: %w", err)
	}
	return nil
}

func (r *memberPostgresRepository) FindByID(ctx context.Context, id string) (*domain.Member, error) {
	return r.scanOne(ctx, `
		SELECT id, login_id, nickname, status, salt, secret, deleted_at
		FROM members WHERE id = $1
	`, id)
}

func (r *memberPostgresRepository) FindActiveByLoginID(ctx context.Context, loginID string) (*domain.Member, error) {
	return r.scanOne(ctx, `
		SELECT id, login_id, nickname, status, salt, secret, deleted_at
		FROM members WHERE login_id = $1 AND status = 'active'
	`, loginID)
}

func (r *memberPostgresRepository) Update(ctx context.Context, m *domain.Member) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE members
		SET login_id = $2,
		    nickname = $3,
		    salt = $4,
		    secret = $5,
		    status = $6,
		    deleted_at = $7,
		    updated_at = now()
		WHERE id = $1
	`, m.ID(), m.LoginID(), m.Nickname(), m.Salt(), m.Secret(), string(m.Status()), nullTime(m.DeletedAt()))
	if err != nil {
		if isUniqueViolation(err) {
			return memberuc.ErrConflict
		}
		return fmt.Errorf("update member: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update member rows: %w", err)
	}
	if n == 0 {
		return memberuc.ErrNotFound
	}
	return nil
}

func (r *memberPostgresRepository) scanOne(ctx context.Context, query string, arg any) (*domain.Member, error) {
	var (
		id, loginID, nickname, status string
		salt, secret                  []byte
		deletedAt                     sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&id, &loginID, &nickname, &status, &salt, &secret, &deletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, memberuc.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query member: %w", err)
	}
	var del *time.Time
	if deletedAt.Valid {
		t := deletedAt.Time.UTC()
		del = &t
	}
	return domain.Reconstitute(id, loginID, nickname, domain.Status(status), salt, secret, del), nil
}

func nullTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC()
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
