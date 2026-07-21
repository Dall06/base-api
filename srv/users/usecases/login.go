package usecases

import (
	"context"

	"base-api/pkg/crypto"
	"base-api/pkg/errs"
	"base-api/pkg/jwt"
	"base-api/srv/users/domain"
	"base-api/srv/users/ports"
)

type LoginUseCase struct {
	repo   ports.UserRepository
	jwtGen *jwt.Generator
}

func NewLoginUseCase(repo ports.UserRepository, jwtGen *jwt.Generator) ports.LoginUseCase {
	return &LoginUseCase{
		repo:   repo,
		jwtGen: jwtGen,
	}
}

func (uc *LoginUseCase) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := uc.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errs.UnauthorizedError("credenciales inválidas")
	}

	if !crypto.CheckPassword(user.PasswordHash, req.Password) {
		return nil, errs.UnauthorizedError("credenciales inválidas")
	}

	tokenOut, err := uc.jwtGen.Generate(jwt.GenerateInput{
		UserID: user.ID,
		Email:  user.Email,
		Role:   "user",
	})
	if err != nil {
		return nil, errs.InternalError("error al generar token: %v", err)
	}

	return &domain.AuthResponse{
		User:  user,
		Token: tokenOut.Token,
	}, nil
}
