package member

import "context"

// Repository persists members. Implemented in infrastructure.
type Repository interface {
	Save(ctx context.Context, m *Member) error
	FindActiveByLoginID(ctx context.Context, loginID string) (*Member, error)
	Update(ctx context.Context, m *Member) error
}
