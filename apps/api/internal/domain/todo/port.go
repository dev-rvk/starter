package todo

import "context"

type Repository interface {
	Create(ctx context.Context, t *Todo) error
	GetByID(ctx context.Context, id string) (*Todo, error)
	Update(ctx context.Context, t *Todo) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int32) ([]*Todo, error)
}
