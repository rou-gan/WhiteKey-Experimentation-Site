package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domain "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/message"
)

type messagePostgresRepository struct {
	db *sql.DB
}

func NewMessagePostgresRepository(db *sql.DB) domain.Repository {
	return &messagePostgresRepository{db: db}
}

func (r *messagePostgresRepository) FindByID(ctx context.Context, id domain.ID) (*domain.Message, error) {
	var content string
	err := r.db.QueryRowContext(ctx, `SELECT content FROM messages WHERE id = $1`, string(id)).Scan(&content)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("message not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("query message: %w", err)
	}
	return domain.NewMessage(id, content)
}
