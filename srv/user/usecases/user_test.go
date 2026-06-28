package usecases

import (
	"context"
	"testing"
	"time"

	"base-api/pkg/errs"
	"base-api/pkg/jwt"
	"base-api/srv/user/domain"
	"base-api/srv/user/ports"
	"base-api/srv/user/repositories"

	"github.com/stretchr/testify/assert"
)

func TestUserUsecase_SignupAndLogin(t *testing.T) {
	jwtGen := jwt.NewGenerator(jwt.Config{
		Secret:     "test-secret",
		Expiration: 1 * time.Hour,
	})

	tests := []struct {
		name          string
		action        string // "signup" o "login"
		email         string
		password      string
		nameVal       string
		expectedErr   error
		setupRepoUser *domain.User // pre-load user for login test cases
	}{
		{
			name:        "Signup exitoso",
			action:      "signup",
			email:       "juan@example.com",
			password:    "password123",
			nameVal:     "Juan Perez",
			expectedErr: nil,
		},
		{
			name:        "Signup fallido - usuario duplicado",
			action:      "signup",
			email:       "juan@example.com",
			password:    "otrapassword",
			nameVal:     "Juan Perez",
			expectedErr: errs.ErrConflict,
		},
		{
			name:        "Login exitoso",
			action:      "login",
			email:       "pedro@example.com",
			password:    "pedropass",
			expectedErr: nil,
			setupRepoUser: func() *domain.User {
				u, _ := domain.NewUser("1", "pedro@example.com", "pedropass", "Pedro")
				return u
			}(),
		},
		{
			name:        "Login fallido - contraseña incorrecta",
			action:      "login",
			email:       "pedro@example.com",
			password:    "claveincorrecta",
			expectedErr: errs.ErrUnauthorized,
			setupRepoUser: func() *domain.User {
				u, _ := domain.NewUser("1", "pedro@example.com", "pedropass", "Pedro")
				return u
			}(),
		},
		{
			name:        "Login fallido - usuario inexistente",
			action:      "login",
			email:       "inexistente@example.com",
			password:    "algunaclave",
			expectedErr: errs.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repositories.NewMemoryUserRepository()
			uc := NewUserUsecase(repo, jwtGen)
			ctx := context.Background()

			// Pre-cargar usuario si aplica
			if tt.setupRepoUser != nil {
				_, _ = repo.Create(ctx, tt.setupRepoUser)
			}

			// Para el caso de duplicados en signup, pre-crear el usuario
			if tt.action == "signup" && tt.expectedErr == errs.ErrConflict {
				u, _ := domain.NewUser("existing-id", tt.email, "pwd", "Juan")
				_, _ = repo.Create(ctx, u)
			}

			var res *ports.AuthResponse
			var err error

			if tt.action == "signup" {
				res, err = uc.Signup(ctx, ports.SignupInput{
					Email:    tt.email,
					Password: tt.password,
					Name:     tt.nameVal,
				})
			} else {
				res, err = uc.Login(ctx, ports.LoginInput{
					Email:    tt.email,
					Password: tt.password,
				})
			}

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.NotEmpty(t, res.Token)
				assert.Equal(t, tt.email, res.User.Email)
			}
		})
	}
}
