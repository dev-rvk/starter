package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/starterpack/api/internal/adapters/persistence/postgres/sqlc"
	userdomain "github.com/starterpack/api/internal/domain/user"
)

// UserRepository implements userdomain.Repository backed by Postgres via sqlc.
type UserRepository struct {
	q *sqlc.Queries
}

// NewUserRepository wires the repository to a pgx pool.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{q: sqlc.New(pool)}
}

func (r *UserRepository) Create(ctx context.Context, u *userdomain.User) error {
	_, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return userdomain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*userdomain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, userdomain.ErrNotFound
		}
		return nil, err
	}
	return toDomain(row)
}

func (r *UserRepository) List(ctx context.Context, limit, offset int32) ([]*userdomain.User, error) {
	rows, err := r.q.ListUsers(ctx, sqlc.ListUsersParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	out := make([]*userdomain.User, 0, len(rows))
	for _, row := range rows {
		u, err := toDomain(row)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}

func toDomain(row sqlc.User) (*userdomain.User, error) {
	return &userdomain.User{
		ID:        row.ID,
		Username:  row.Username,
		Email:     row.Email,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
