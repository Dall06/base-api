package usecases

import (
	"context"

	"base-api/pkg/errs"
	"base-api/pkg/jwt"
	"base-api/srv/users/domain"
	"base-api/srv/users/ports"

	"github.com/google/uuid"
)

type SignupUseCase struct {
	repo   ports.UserRepository
	jwtGen *jwt.Generator
}

func NewSignupUseCase(repo ports.UserRepository, jwtGen *jwt.Generator) ports.SignupUseCase {
	return &SignupUseCase{
		repo:   repo,
		jwtGen: jwtGen,
	}
}

func (uc *SignupUseCase) Signup(ctx context.Context, req domain.SignupRequest) (*domain.AuthResponse, error) {
	existing, err := uc.repo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, errs.ConflictError("el usuario con email '%s' ya existe", req.Email)
	}

	id := uuid.New().String()
	user, err := domain.NewUser(id, req.Email, req.Password, req.Name)
	if err != nil {
		return nil, errs.InternalError("error al crear usuario: %v", err)
	}

	saved, err := uc.repo.Create(ctx, user)
	if err != nil {
		return nil, errs.InternalError("error al guardar usuario: %v", err)
	}

	tokenOut, err := uc.jwtGen.Generate(jwt.GenerateInput{
		UserID: saved.ID,
		Email:  saved.Email,
		Role:   "user",
	})
	if err != nil {
		return nil, errs.InternalError("error al generar token: %v", err)
	}

	return &domain.AuthResponse{
		User:  saved,
		Token: tokenOut.Token,
	}, nil
}
