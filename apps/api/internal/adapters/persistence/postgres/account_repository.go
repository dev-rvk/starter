package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/starterpack/api/internal/adapters/persistence/postgres/sqlc"
	accountdomain "github.com/starterpack/api/internal/domain/account"
)

// Compile-time interface check.
var _ accountdomain.Repository = (*AccountRepository)(nil)

// AccountRepository implements accountdomain.Repository backed by Postgres via sqlc.
type AccountRepository struct {
	q *sqlc.Queries
}

// NewAccountRepository wires the repository to a pgx pool.
func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{q: sqlc.New(pool)}
}

func (r *AccountRepository) Create(ctx context.Context, a *accountdomain.Account) error {
	_, err := r.q.CreateAccount(ctx, sqlc.CreateAccountParams{
		ID:           a.ID,
		Email:        a.Email,
		PasswordHash: a.PasswordHash,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return accountdomain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *AccountRepository) GetByEmail(ctx context.Context, email string) (*accountdomain.Account, error) {
	row, err := r.q.GetAccountByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, accountdomain.ErrNotFound
		}
		return nil, err
	}
	return toAccountDomain(row), nil
}

func toAccountDomain(row sqlc.Account) *accountdomain.Account {
	return &accountdomain.Account{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
