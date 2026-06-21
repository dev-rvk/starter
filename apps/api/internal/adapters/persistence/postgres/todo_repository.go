package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/starterpack/api/internal/adapters/persistence/postgres/sqlc"
	"github.com/starterpack/api/internal/domain/todo"
)

var _ todo.Repository = (*TodoRepository)(nil)

// struct
type TodoRepository struct {
	q *sqlc.Queries
}

// constructor
func NewTodoRepository(pool *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{q: sqlc.New(pool)}
}

// implementation of ports
func (r *TodoRepository) Create(ctx context.Context, todo *todo.Todo) error {
	if _, err := r.q.CreateTodo(ctx, sqlc.CreateTodoParams{
		ID:        todo.ID,
		Title:     todo.Title,
		Completed: pgtype.Bool{Bool: todo.Completed, Valid: true},
		CreatedAt: todo.CreatedAt,
		UpdatedAt: todo.UpdatedAt,
	}); err != nil {
		return err
	}
	return nil
}

func (r *TodoRepository) GetByID(ctx context.Context, id string) (*todo.Todo, error) {
	row, err := r.q.GetTodoByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, todo.ErrNotFound
		}
		return nil, err
	}

	return toTodoDomain(row), nil
}

func (r *TodoRepository) Update(ctx context.Context, todo *todo.Todo) error {
	if _, err := r.q.UpdateTodo(ctx, sqlc.UpdateTodoParams{
		ID:        todo.ID,
		Title:     todo.Title,
		Completed: pgtype.Bool{Bool: todo.Completed, Valid: true},
		UpdatedAt: todo.UpdatedAt,
	}); err != nil {
		return err
	}
	return nil
}

func (r *TodoRepository) Delete(ctx context.Context, id string) error {
	if err := r.q.DeleteTodo(ctx, id); err != nil {
		return err
	}
	return nil
}

func (r *TodoRepository) List(ctx context.Context, limit, offset int32) ([]*todo.Todo, error) {
	rows, err := r.q.ListTodos(ctx, sqlc.ListTodosParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	todoList := make([]*todo.Todo, 0, len(rows))
	for _, row := range rows {
		todoList = append(todoList, toTodoDomain(row))
	}
	return todoList, nil
}

func toTodoDomain(row sqlc.Todo) *todo.Todo {
	return &todo.Todo{
		ID:        row.ID,
		Title:     row.Title,
		Completed: row.Completed.Bool,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
