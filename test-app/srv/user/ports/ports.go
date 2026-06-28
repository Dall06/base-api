package ports

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
	"github.com/diegoaleon/test-app/srv/user/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type SignupInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
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

type UserHandler interface {
	Signup(c echo.Context) error
	Login(c echo.Context) error
	GetMe(c echo.Context) error
}

type DBConnector interface {
	GetDB() *bun.DB
	Close() error
}
