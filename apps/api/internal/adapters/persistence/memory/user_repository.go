// Package memory is an in-memory implementation of the user repository port.
// It lets the API boot and be exercised without a database (e.g. when
// DATABASE_URL is unset), and is handy for tests.
package memory

import (
	"context"
	"sort"
	"sync"

	userdomain "github.com/starterpack/api/internal/domain/user"
)

// Verify UserRepository implements the domain port at compile time.
var _ userdomain.Repository = (*UserRepository)(nil)

// UserRepository is a goroutine-safe in-memory user store.
type UserRepository struct {
	mu    sync.RWMutex
	users map[string]*userdomain.User
}

// NewUserRepository creates an empty in-memory repository.
func NewUserRepository() *UserRepository {
	return &UserRepository{users: make(map[string]*userdomain.User)}
}

func (r *UserRepository) Create(_ context.Context, u *userdomain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, ex := range r.users {
		if ex.Username == u.Username || ex.Email == u.Email {
			return userdomain.ErrAlreadyExists
		}
	}
	r.users[u.ID] = u
	return nil
}

func (r *UserRepository) GetByID(_ context.Context, id string) (*userdomain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, userdomain.ErrNotFound
	}
	return u, nil
}

func (r *UserRepository) List(_ context.Context, limit, offset int32) ([]*userdomain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]*userdomain.User, 0, len(r.users))
	for _, u := range r.users {
		all = append(all, u)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})
	start := int(offset)
	if start > len(all) {
		start = len(all)
	}
	end := start + int(limit)
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], nil
}
