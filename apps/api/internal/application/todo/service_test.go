package todoapp

import (
	"context"
	"errors"
	"testing"

	"github.com/starterpack/api/internal/adapters/persistence/memory"
	"github.com/starterpack/api/internal/domain"
	"github.com/starterpack/api/internal/platform/validator"
)

func TestService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		title   string
		wantErr error
	}{
		{
			name:    "success",
			title:   "Buy groceries",
			wantErr: nil,
		},
		{
			name:    "validation_error_empty_title",
			title:   "",
			wantErr: domain.ErrValidation,
		},
		{
			name:    "validation_error_title_too_long",
			title:   "this title is definitely more than fifty characters long and should trigger a validation error",
			wantErr: domain.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewTodoRepository()
			v := validator.New()
			svc := NewService(repo, v)

			todo, err := svc.Create(context.Background(), tt.title)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				if todo != nil {
					t.Errorf("expected nil todo on error, got %v", todo)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if todo == nil {
					t.Fatal("expected non-nil todo")
				}
				if todo.Title != tt.title {
					t.Errorf("expected title %q, got %q", tt.title, todo.Title)
				}
				if todo.Completed {
					t.Error("expected new todo to be uncompleted")
				}
				if todo.ID == "" {
					t.Error("expected generated ID to not be empty")
				}
			}
		})
	}
}

func TestService_Get(t *testing.T) {
	t.Parallel()

	repo := memory.NewTodoRepository()
	v := validator.New()
	svc := NewService(repo, v)
	ctx := context.Background()

	// Get non-existent
	_, err := svc.Get(ctx, "non-existent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// Create and Get
	todo, err := svc.Create(ctx, "Buy groceries")
	if err != nil {
		t.Fatalf("failed to create todo: %v", err)
	}

	got, err := svc.Get(ctx, todo.ID)
	if err != nil {
		t.Fatalf("failed to get todo: %v", err)
	}
	if got.ID != todo.ID || got.Title != todo.Title || got.Completed != todo.Completed {
		t.Errorf("mismatch: expected %+v, got %+v", todo, got)
	}
}

func TestService_Update(t *testing.T) {
	t.Parallel()

	repo := memory.NewTodoRepository()
	v := validator.New()
	svc := NewService(repo, v)
	ctx := context.Background()

	// Update non-existent
	_, err := svc.Update(ctx, "non-existent", "Title", true)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for non-existent, got %v", err)
	}

	// Create and Update
	todo, err := svc.Create(ctx, "Buy groceries")
	if err != nil {
		t.Fatalf("failed to create todo: %v", err)
	}

	updated, err := svc.Update(ctx, todo.ID, "Clean kitchen", true)
	if err != nil {
		t.Fatalf("failed to update todo: %v", err)
	}
	if updated.Title != "Clean kitchen" || !updated.Completed {
		t.Errorf("expected updated values, got %+v", updated)
	}

	// Update with invalid title (empty)
	_, err = svc.Update(ctx, todo.ID, "", false)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("expected ErrValidation for empty title, got %v", err)
	}
}

func TestService_Delete(t *testing.T) {
	t.Parallel()

	repo := memory.NewTodoRepository()
	v := validator.New()
	svc := NewService(repo, v)
	ctx := context.Background()

	// Delete non-existent (should succeed silently)
	err := svc.Delete(ctx, "non-existent")
	if err != nil {
		t.Errorf("unexpected error deleting non-existent: %v", err)
	}

	// Create, Delete, and verify missing
	todo, err := svc.Create(ctx, "Buy groceries")
	if err != nil {
		t.Fatalf("failed to create todo: %v", err)
	}

	err = svc.Delete(ctx, todo.ID)
	if err != nil {
		t.Fatalf("failed to delete todo: %v", err)
	}

	_, err = svc.Get(ctx, todo.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestService_List(t *testing.T) {
	t.Parallel()

	repo := memory.NewTodoRepository()
	v := validator.New()
	svc := NewService(repo, v)
	ctx := context.Background()

	// List empty
	todos, err := svc.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("list empty failed: %v", err)
	}
	if len(todos) != 0 {
		t.Errorf("expected empty list, got %d items", len(todos))
	}

	// Create some todos
	for i := 1; i <= 5; i++ {
		_, err := svc.Create(ctx, "Todo item")
		if err != nil {
			t.Fatalf("failed to create todo: %v", err)
		}
	}

	tests := []struct {
		name           string
		limit          int32
		offset         int32
		expectedLength int
	}{
		{
			name:           "default parameters",
			limit:          0,
			offset:         0,
			expectedLength: 5,
		},
		{
			name:           "offset skips",
			limit:          2,
			offset:         2,
			expectedLength: 2,
		},
		{
			name:           "limit out of bounds too large",
			limit:          150,
			offset:         0,
			expectedLength: 5,
		},
		{
			name:           "offset out of range",
			limit:          5,
			offset:         10,
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.List(ctx, tt.limit, tt.offset)
			if err != nil {
				t.Fatalf("list failed: %v", err)
			}
			if len(got) != tt.expectedLength {
				t.Errorf("expected %d items, got %d", tt.expectedLength, len(got))
			}
		})
	}
}
