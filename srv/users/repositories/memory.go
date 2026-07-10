package repositories

import (
	"context"
	"sync"

	"base-api/pkg/errs"
	"base-api/srv/users/domain"
	"base-api/srv/users/ports"

	"github.com/google/uuid"
)

type memoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

func NewMemoryUserRepository() ports.UserRepository {
	return &memoryUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (r *memoryUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	for _, u := range r.users {
		if u.Email == user.Email {
			return nil, errs.ConflictError("el usuario con email '%s' ya existe", user.Email)
		}
	}

	r.users[user.ID] = user
	return user, nil
}

func (r *memoryUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, errs.NotFoundError("usuario no encontrado")
	}
	return user, nil
}

func (r *memoryUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errs.NotFoundError("usuario no encontrado")
}
