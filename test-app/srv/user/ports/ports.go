package ports

import (
	"context"
	"github.com/diegoaleon/test-app/srv/user/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}

type SignupInput struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=50"`
}

type LoginInput struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	User  *domain.User `json:"user"`
	Token string       `json:"token"`
}

type UserUsecase interface {
	Signup(ctx context.Context, input SignupInput) (*AuthResponse, error)
	Login(ctx context.Context, input LoginInput) (*AuthResponse, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
}
