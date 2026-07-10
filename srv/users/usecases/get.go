package usecases

import (
	"context"

	"base-api/srv/users/domain"
	"base-api/srv/users/ports"
)

type GetUseCase struct {
	repo ports.UserRepository
}

func NewGetUseCase(repo ports.UserRepository) ports.GetUseCase {
	return &GetUseCase{
		repo: repo,
	}
}

func (uc *GetUseCase) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return uc.repo.GetByID(ctx, id)
}
