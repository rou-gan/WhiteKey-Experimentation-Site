package message

import (
	"context"

	domain "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/message"
)

type UseCase struct {
	repo domain.Repository
}

func NewUseCase(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	return uc.repo.FindByID(ctx, domain.ID(id))
}
