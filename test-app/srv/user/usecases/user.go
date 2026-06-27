package usecases

import (
	"context"
	"github.com/diegoaleon/test-app/pkg/errs"
	"github.com/diegoaleon/test-app/pkg/jwt"
	"github.com/diegoaleon/test-app/srv/user/domain"
	"github.com/diegoaleon/test-app/srv/user/ports"

	"github.com/google/uuid"
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
	existing, err := u.repo.GetByUsername(ctx, input.Username)
	if err == nil && existing != nil {
		return nil, errs.ConflictError("el usuario '%s' ya existe", input.Username)
	}

	id := uuid.New().String()
	user, err := domain.NewUser(id, input.Username, input.Password)
	if err != nil {
		return nil, errs.InternalError("error al cifrar la contraseña: %v", err)
	}

	if err := u.repo.Create(ctx, user); err != nil {
		return nil, errs.InternalError("error al guardar el usuario: %v", err)
	}

	tokenOut, err := u.jwtGenerator.Generate(jwt.GenerateInput{
		UserID: user.ID,
		Email:  user.Username,
		Role:   "user",
	})
	if err != nil {
		return nil, errs.InternalError("error al generar el token: %v", err)
	}

	return &ports.AuthResponse{
		User:  user,
		Token: tokenOut.Token,
	}, nil
}

func (u *userUsecase) Login(ctx context.Context, input ports.LoginInput) (*ports.AuthResponse, error) {
	user, err := u.repo.GetByUsername(ctx, input.Username)
	if err != nil {
		return nil, errs.NotFoundError("usuario no encontrado")
	}

	if !user.CheckPassword(input.Password) {
		return nil, errs.UnauthorizedError("credenciales inválidas")
	}

	tokenOut, err := u.jwtGenerator.Generate(jwt.GenerateInput{
		UserID: user.ID,
		Email:  user.Username,
		Role:   "user",
	})
	if err != nil {
		return nil, errs.InternalError("error al generar el token: %v", err)
	}

	return &ports.AuthResponse{
		User:  user,
		Token: tokenOut.Token,
	}, nil
}

func (u *userUsecase) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errs.NotFoundError("usuario con ID %s no encontrado", id)
	}
	return user, nil
}
