package todoapp

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/starterpack/api/internal/domain/todo"
	"github.com/starterpack/api/internal/platform/validator"
)

// TodoService is the inbound port consumed by the HTTP adapter.
type TodoService interface {
	Create(ctx context.Context, title string) (*todo.Todo, error)
	Get(ctx context.Context, id string) (*todo.Todo, error)
	Update(ctx context.Context, id string, title string, completed bool) (*todo.Todo, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int32) ([]*todo.Todo, error)
}

// Compile-time check: Service must satisfy TodoService.
var _ TodoService = (*Service)(nil)

// Service implements the todo use cases.
type Service struct {
	repo todo.Repository
	v    *validator.Validator
}

// NewService wires a use-case service to a repository port and a validator.
func NewService(repo todo.Repository, v *validator.Validator) *Service {
	return &Service{repo: repo, v: v}
}

// Create validates and persists a new todo.
func (s *Service) Create(ctx context.Context, title string) (*todo.Todo, error) {
	task := todo.New(uuid.NewString(), title, false, time.Now().UTC())
	if err := s.v.ValidateAndMap("todo", task); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

// Get returns a single todo by id.
func (s *Service) Get(ctx context.Context, id string) (*todo.Todo, error) {
	return s.repo.GetByID(ctx, id)
}

// Update modifies a todo's title and completion state.
func (s *Service) Update(ctx context.Context, id, title string, completed bool) (*todo.Todo, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	task.Title = title
	task.Completed = completed
	task.UpdatedAt = time.Now().UTC()
	if err := s.v.ValidateAndMap("todo", task); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

// Delete removes a todo by id.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// List returns a page of todos.
func (s *Service) List(ctx context.Context, limit, offset int32) ([]*todo.Todo, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}
