package usecases

import (
	"context"
	"testing"

	"base-api/srv/users/domain"
	"base-api/srv/users/repositories"

	"github.com/stretchr/testify/assert"
)

func TestGetUseCase_GetByID(t *testing.T) {
	repo := repositories.NewMemoryUserRepository()
	uc := NewGetUseCase(repo)
	ctx := context.Background()

	u := &domain.User{ID: "test-id", Email: "test@example.com", Name: "Test User"}
	_, _ = repo.Create(ctx, u)

	t.Run("GetByID exitoso", func(t *testing.T) {
		res, err := uc.GetByID(ctx, "test-id")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "test@example.com", res.Email)
	})

	t.Run("GetByID no encontrado", func(t *testing.T) {
		res, err := uc.GetByID(ctx, "non-existent")
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}
