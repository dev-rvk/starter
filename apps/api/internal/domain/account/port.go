package account

import "context"

// Repository is the persistence port for accounts. The domain owns this
// interface; adapters (Postgres/sqlc, in-memory) implement it.
type Repository interface {
	Create(ctx context.Context, a *Account) error
	GetByEmail(ctx context.Context, email string) (*Account, error)
}
