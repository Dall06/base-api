package repositories

import (
	"context"
	"sync"
	"template-placeholder/pkg/errs"
	"template-placeholder/srv/user/domain"
	"template-placeholder/srv/user/ports"
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

func (r *memoryUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validar que no exista el username (chequeo extra)
	for _, u := range r.users {
		if u.Username == user.Username {
			return errs.ConflictError("el usuario '%s' ya existe", user.Username)
		}
	}

	r.users[user.ID] = user
	return nil
}

func (r *memoryUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, errs.NotFoundError("usuario con ID %s no encontrado", id)
	}

	return user, nil
}

func (r *memoryUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}

	return nil, errs.NotFoundError("usuario con username %s no encontrado", username)
}
