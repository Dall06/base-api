package usecases

import (
	"context"

	"base-api/pkg/errs"
	"base-api/pkg/jwt"
	"base-api/srv/user/domain"
	"base-api/srv/user/ports"
)

type userUsecase struct {
	repo         ports.UserRepository
	jwtGenerator *jwt.Generator
}

func NewUserUsecase(repo ports.UserRepository, jwtGenerator *jwt.Generator) ports.UserUsecase {
	return &userUsecase{
		repo:         repo,
		jwtGenerator: jwtGenerator,
	}
}

func (u *userUsecase) Signup(ctx context.Context, input ports.SignupInput) (*ports.AuthResponse, error) {
	existing, err := u.repo.GetByEmail(ctx, input.Email)
	if err == nil && existing != nil {
		return nil, errs.ConflictError("el usuario con email '%s' ya existe", input.Email)
	}

	user, err := domain.NewUser("", input.Email, input.Password, input.Name)
	if err != nil {
		return nil, errs.InternalError("error al crear usuario: %v", err)
	}

	saved, err := u.repo.Create(ctx, user)
	if err != nil {
		return nil, errs.InternalError("error al guardar usuario: %v", err)
	}

	tokenOut, err := u.jwtGenerator.Generate(jwt.GenerateInput{
		UserID: saved.ID,
		Email:  saved.Email,
		Role:   "user",
	})
	if err != nil {
		return nil, errs.InternalError("error al generar token: %v", err)
	}

	return &ports.AuthResponse{
		User:  saved,
		Token: tokenOut.Token,
	}, nil
}

func (u *userUsecase) Login(ctx context.Context, input ports.LoginInput) (*ports.AuthResponse, error) {
	user, err := u.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, errs.UnauthorizedError("credenciales inválidas")
	}

	if !user.CheckPassword(input.Password) {
		return nil, errs.UnauthorizedError("credenciales inválidas")
	}

	tokenOut, err := u.jwtGenerator.Generate(jwt.GenerateInput{
		UserID: user.ID,
		Email:  user.Email,
		Role:   "user",
	})
	if err != nil {
		return nil, errs.InternalError("error al generar token: %v", err)
	}

	return &ports.AuthResponse{
		User:  user,
		Token: tokenOut.Token,
	}, nil
}

func (u *userUsecase) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return u.repo.GetByID(ctx, id)
}
