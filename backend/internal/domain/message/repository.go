package message

import "context"

type Repository interface {
	FindByID(ctx context.Context, id ID) (*Message, error)
}
