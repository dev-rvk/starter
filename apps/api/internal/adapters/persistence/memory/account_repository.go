package memory

import (
	"context"
	"sync"

	accountdomain "github.com/starterpack/api/internal/domain/account"
)

// Verify AccountRepository implements the domain port at compile time.
var _ accountdomain.Repository = (*AccountRepository)(nil)

// AccountRepository is a goroutine-safe in-memory account store.
type AccountRepository struct {
	mu       sync.RWMutex
	accounts map[string]*accountdomain.Account // keyed by ID
}

// NewAccountRepository creates an empty in-memory account repository.
func NewAccountRepository() *AccountRepository {
	return &AccountRepository{accounts: make(map[string]*accountdomain.Account)}
}

func (r *AccountRepository) Create(_ context.Context, a *accountdomain.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, ex := range r.accounts {
		if ex.Email == a.Email {
			return accountdomain.ErrAlreadyExists
		}
	}
	r.accounts[a.ID] = a
	return nil
}

func (r *AccountRepository) GetByEmail(_ context.Context, email string) (*accountdomain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.accounts {
		if a.Email == email {
			return a, nil
		}
	}
	return nil, accountdomain.ErrNotFound
}
