package usecases

import (
	"context"
	"testing"
	"time"

	"base-api/pkg/crypto"
	"base-api/pkg/errs"
	"base-api/pkg/jwt"
	"base-api/srv/users/domain"
	"base-api/srv/users/repositories"

	"github.com/stretchr/testify/assert"
)

func TestLoginUseCase_Login(t *testing.T) {
	jwtGen := jwt.NewGenerator(jwt.Config{
		Secret:     "test-secret",
		Expiration: 1 * time.Hour,
	})

	tests := []struct {
		name          string
		email         string
		password      string
		expectedErr   error
		setupRepoUser *domain.User
	}{
		{
			name:        "Login exitoso",
			email:       "pedro@example.com",
			password:    "pedropass",
			expectedErr: nil,
			setupRepoUser: func() *domain.User {
				hashed, _ := crypto.HashPassword("pedropass")
				return &domain.User{ID: "1", Email: "pedro@example.com", PasswordHash: hashed, Name: "Pedro"}
			}(),
		},
		{
			name:        "Login fallido - contraseña incorrecta",
			email:       "pedro@example.com",
			password:    "claveincorrecta",
			expectedErr: errs.ErrUnauthorized,
			setupRepoUser: func() *domain.User {
				hashed, _ := crypto.HashPassword("pedropass")
				return &domain.User{ID: "1", Email: "pedro@example.com", PasswordHash: hashed, Name: "Pedro"}
			}(),
		},
		{
			name:        "Login fallido - usuario inexistente",
			email:       "inexistente@example.com",
			password:    "algunaclave",
			expectedErr: errs.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repositories.NewMemoryUserRepository()
			uc := NewLoginUseCase(repo, jwtGen)
			ctx := context.Background()

			if tt.setupRepoUser != nil {
				_, _ = repo.Create(ctx, tt.setupRepoUser)
			}

			res, err := uc.Login(ctx, domain.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			})

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
