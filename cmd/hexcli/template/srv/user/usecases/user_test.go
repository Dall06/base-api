package usecases

import (
	"context"
	"testing"
	"time"

	"template-placeholder/pkg/errs"
	"template-placeholder/pkg/jwt"
	"template-placeholder/srv/user/domain"
	"template-placeholder/srv/user/ports"
	"template-placeholder/srv/user/repositories"

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
		username      string
		password      string
		expectedErr   error
		setupRepoUser *domain.User // pre-load user for login test cases
	}{
		{
			name:        "Signup exitoso",
			action:      "signup",
			username:    "juan",
			password:    "password123",
			expectedErr: nil,
		},
		{
			name:        "Signup fallido - usuario duplicado",
			action:      "signup",
			username:    "juan",
			password:    "otrapassword",
			expectedErr: errs.ErrConflict,
		},
		{
			name:        "Login exitoso",
			action:      "login",
			username:    "pedro",
			password:    "pedropass",
			expectedErr: nil,
			setupRepoUser: func() *domain.User {
				u, _ := domain.NewUser("1", "pedro", "pedropass")
				return u
			}(),
		},
		{
			name:        "Login fallido - contraseña incorrecta",
			action:      "login",
			username:    "pedro",
			password:    "claveincorrecta",
			expectedErr: errs.ErrUnauthorized,
			setupRepoUser: func() *domain.User {
				u, _ := domain.NewUser("1", "pedro", "pedropass")
				return u
			}(),
		},
		{
			name:        "Login fallido - usuario inexistente",
			action:      "login",
			username:    "inexistente",
			password:    "algunaclave",
			expectedErr: errs.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repositories.NewMemoryUserRepository()
			uc := NewUserUsecase(repo, jwtGen)
			ctx := context.Background()

			// Pre-cargar usuario si aplica
			if tt.setupRepoUser != nil {
				_ = repo.Create(ctx, tt.setupRepoUser)
			}

			// Para el caso de duplicados en signup, pre-crear el usuario
			if tt.action == "signup" && tt.expectedErr == errs.ErrConflict {
				u, _ := domain.NewUser("existing-id", tt.username, "pwd")
				_ = repo.Create(ctx, u)
			}

			var res *ports.AuthResponse
			var err error

			if tt.action == "signup" {
				res, err = uc.Signup(ctx, ports.SignupInput{
					Username: tt.username,
					Password: tt.password,
				})
			} else {
				res, err = uc.Login(ctx, ports.LoginInput{
					Username: tt.username,
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
				assert.Equal(t, tt.username, res.User.Username)
			}
		})
	}
}
