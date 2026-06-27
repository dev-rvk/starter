package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/starterpack/api/internal/domain"
	"github.com/starterpack/api/internal/domain/todo"
)

// newTestTodo is a helper that builds a domain todo with a deterministic timestamp.
func newTestTodo(t *testing.T, id, title string, completed bool, createdAt time.Time) *todo.Todo {
	t.Helper()
	return todo.New(id, title, completed, createdAt)
}

func TestTodoRepository_CreateAndGetByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name      string
		id        string
		title     string
		completed bool
	}{
		{
			name:      "round-trip stores and retrieves todo",
			id:        "t-1",
			title:     "Buy milk",
			completed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := NewTodoRepository()
			now := time.Now()
			td := newTestTodo(t, tt.id, tt.title, tt.completed, now)

			if err := repo.Create(ctx, td); err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}

			got, err := repo.GetByID(ctx, tt.id)
			if err != nil {
				t.Fatalf("GetByID() unexpected error: %v", err)
			}
			if got.ID != tt.id {
				t.Errorf("ID = %q, want %q", got.ID, tt.id)
			}
			if got.Title != tt.title {
				t.Errorf("Title = %q, want %q", got.Title, tt.title)
			}
			if got.Completed != tt.completed {
				t.Errorf("Completed = %v, want %v", got.Completed, tt.completed)
			}
			if !got.CreatedAt.Equal(now) {
				t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, now)
			}
			if !got.UpdatedAt.Equal(now) {
				t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, now)
			}
		})
	}
}

func TestTodoRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name string
		id   string
	}{
		{
			name: "non-existent ID returns ErrNotFound",
			id:   "does-not-exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := NewTodoRepository()

			_, err := repo.GetByID(ctx, tt.id)
			if err == nil {
				t.Fatal("GetByID() expected error, got nil")
			}
			if !errors.Is(err, domain.ErrNotFound) {
				t.Errorf("error = %v, want errors.Is(domain.ErrNotFound)", err)
			}
		})
	}
}

func TestTodoRepository_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name      string
		seed      *todo.Todo
		update    *todo.Todo
		wantErr   bool
		wantErrIs error
		wantTitle string
		wantDone  bool
	}{
		{
			name:      "update existing persists changes",
			seed:      todo.New("t-1", "Buy milk", false, now),
			update:    todo.New("t-1", "Buy oat milk", true, now),
			wantErr:   false,
			wantTitle: "Buy oat milk",
			wantDone:  true,
		},
		{
			name:      "update non-existent returns ErrNotFound",
			seed:      nil,
			update:    todo.New("t-missing", "No such todo", false, now),
			wantErr:   true,
			wantErrIs: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := NewTodoRepository()

			if tt.seed != nil {
				if err := repo.Create(ctx, tt.seed); err != nil {
					t.Fatalf("seed Create() unexpected error: %v", err)
				}
			}

			err := repo.Update(ctx, tt.update)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Update() expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error = %v, want errors.Is(%v)", err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("Update() unexpected error: %v", err)
			}

			got, err := repo.GetByID(ctx, tt.update.ID)
			if err != nil {
				t.Fatalf("GetByID() after Update() unexpected error: %v", err)
			}
			if got.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", got.Title, tt.wantTitle)
			}
			if got.Completed != tt.wantDone {
				t.Errorf("Completed = %v, want %v", got.Completed, tt.wantDone)
			}
		})
	}
}

func TestTodoRepository_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name     string
		seed     *todo.Todo
		deleteID string
	}{
		{
			name:     "delete existing makes GetByID return ErrNotFound",
			seed:     todo.New("t-1", "Buy milk", false, now),
			deleteID: "t-1",
		},
		{
			name:     "delete non-existent returns no error",
			seed:     nil,
			deleteID: "does-not-exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := NewTodoRepository()

			if tt.seed != nil {
				if err := repo.Create(ctx, tt.seed); err != nil {
					t.Fatalf("seed Create() unexpected error: %v", err)
				}
			}

			if err := repo.Delete(ctx, tt.deleteID); err != nil {
				t.Fatalf("Delete() unexpected error: %v", err)
			}

			// If a todo was seeded, verify it is no longer retrievable.
			if tt.seed != nil && tt.seed.ID == tt.deleteID {
				_, err := repo.GetByID(ctx, tt.deleteID)
				if err == nil {
					t.Fatal("GetByID() after Delete() expected error, got nil")
				}
				if !errors.Is(err, domain.ErrNotFound) {
					t.Errorf("error = %v, want errors.Is(domain.ErrNotFound)", err)
				}
			}
		})
	}
}

func TestTodoRepository_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)

	threeTodos := []*todo.Todo{
		todo.New("t-1", "Task A", false, t1),
		todo.New("t-2", "Task B", false, t2),
		todo.New("t-3", "Task C", true, t3),
	}

	tests := []struct {
		name      string
		seed      []*todo.Todo
		limit     int32
		offset    int32
		wantCount int
		wantIDs   []string // expected IDs in desc CreatedAt order; nil to skip
	}{
		{
			name:      "empty repo returns empty non-nil slice",
			seed:      nil,
			limit:     10,
			offset:    0,
			wantCount: 0,
		},
		{
			name:      "sorted by CreatedAt desc",
			seed:      threeTodos,
			limit:     10,
			offset:    0,
			wantCount: 3,
			wantIDs:   []string{"t-3", "t-2", "t-1"},
		},
		{
			name:      "limit restricts results",
			seed:      threeTodos,
			limit:     2,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "offset skips results",
			seed:      threeTodos,
			limit:     10,
			offset:    1,
			wantCount: 2,
		},
		{
			name:      "offset beyond length returns empty slice",
			seed:      threeTodos,
			limit:     10,
			offset:    100,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := NewTodoRepository()
			for _, td := range tt.seed {
				if err := repo.Create(ctx, td); err != nil {
					t.Fatalf("seed Create(%s) unexpected error: %v", td.ID, err)
				}
			}

			got, err := repo.List(ctx, tt.limit, tt.offset)
			if err != nil {
				t.Fatalf("List() unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("List() returned nil slice, want non-nil")
			}
			if len(got) != tt.wantCount {
				t.Fatalf("List() returned %d items, want %d", len(got), tt.wantCount)
			}

			if tt.wantIDs != nil {
				for i, wantID := range tt.wantIDs {
					if got[i].ID != wantID {
						t.Errorf("List()[%d].ID = %q, want %q", i, got[i].ID, wantID)
					}
				}
			}
		})
	}
}
