package usecases

import (
	"context"
	"testing"
	"time"

	"base-api/pkg/errs"
	"base-api/pkg/jwt"
	"base-api/srv/users/domain"
	"base-api/srv/users/repositories"

	"github.com/stretchr/testify/assert"
)

func TestSignupUseCase_Signup(t *testing.T) {
	jwtGen := jwt.NewGenerator(jwt.Config{
		Secret:     "test-secret",
		Expiration: 1 * time.Hour,
	})

	tests := []struct {
		name        string
		email       string
		password    string
		nameVal     string
		preCreate   bool
		expectedErr error
	}{
		{
			name:        "Signup exitoso",
			email:       "juan@example.com",
			password:    "password123",
			nameVal:     "Juan Perez",
			preCreate:   false,
			expectedErr: nil,
		},
		{
			name:        "Signup fallido - usuario duplicado",
			email:       "juan@example.com",
			password:    "otrapassword",
			nameVal:     "Juan Perez",
			preCreate:   true,
			expectedErr: errs.ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repositories.NewMemoryUserRepository()
			uc := NewSignupUseCase(repo, jwtGen)
			ctx := context.Background()

			if tt.preCreate {
				u := &domain.User{ID: "existing-id", Email: tt.email, Name: "Juan"}
				_, _ = repo.Create(ctx, u)
			}

			res, err := uc.Signup(ctx, domain.SignupRequest{
				Email:    tt.email,
				Password: tt.password,
				Name:     tt.nameVal,
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
