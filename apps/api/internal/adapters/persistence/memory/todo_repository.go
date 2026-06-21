// Package memory is an in-memory implementation of the todo repository port.
package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/starterpack/api/internal/domain/todo"
)

// Verify TodoRepository implements the domain port at compile time.
var _ todo.Repository = (*TodoRepository)(nil)

// TodoRepository is a goroutine-safe in-memory todo store.
type TodoRepository struct {
	mu    sync.RWMutex
	todos map[string]*todo.Todo
}

// NewTodoRepository creates an empty in-memory repository.
func NewTodoRepository() *TodoRepository {
	return &TodoRepository{todos: make(map[string]*todo.Todo)}
}

func (r *TodoRepository) Create(_ context.Context, t *todo.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.todos[t.ID] = t
	return nil
}

func (r *TodoRepository) GetByID(_ context.Context, id string) (*todo.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.todos[id]
	if !ok {
		return nil, todo.ErrNotFound
	}
	return t, nil
}

func (r *TodoRepository) Update(_ context.Context, t *todo.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.todos[t.ID]; !ok {
		return todo.ErrNotFound
	}
	r.todos[t.ID] = t
	return nil
}

func (r *TodoRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.todos, id)
	return nil
}

func (r *TodoRepository) List(_ context.Context, limit, offset int32) ([]*todo.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]*todo.Todo, 0, len(r.todos))
	for _, t := range r.todos {
		all = append(all, t)
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
