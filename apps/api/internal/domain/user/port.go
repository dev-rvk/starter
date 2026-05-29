package user

import "context"

// Repository is the persistence port for users. The domain owns this interface;
// adapters (Postgres/sqlc, in-memory, etc.) implement it. Use cases depend only
// on this abstraction, never on a concrete database.
type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	List(ctx context.Context, limit, offset int32) ([]*User, error)
}
