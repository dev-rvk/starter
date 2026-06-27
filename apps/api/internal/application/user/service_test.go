package userapp

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
		name     string
		username string
		email    string
		wantErr  error
	}{
		{
			name:     "success",
			username: "alice",
			email:    "alice@example.com",
			wantErr:  nil,
		},
		{
			name:     "validation_error_empty_username",
			username: "",
			email:    "alice@example.com",
			wantErr:  domain.ErrValidation,
		},
		{
			name:     "validation_error_invalid_email",
			username: "alice",
			email:    "invalid-email",
			wantErr:  domain.ErrValidation,
		},
		{
			name:     "username_too_long",
			username: "toolongusername",
			email:    "alice@example.com",
			wantErr:  domain.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewUserRepository()
			v := validator.New()
			svc := NewService(repo, v)

			u, err := svc.Create(context.Background(), CreateInput{
				Username: tt.username,
				Email:    tt.email,
			})

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				if u != nil {
					t.Errorf("expected nil user on error, got %v", u)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if u == nil {
					t.Fatal("expected non-nil user")
				}
				if u.Username != tt.username || u.Email != tt.email {
					t.Errorf("user fields mismatch: %+v", u)
				}
				if u.ID == "" {
					t.Error("expected generated ID to not be empty")
				}
			}
		})
	}
}

func TestService_Create_Duplicate(t *testing.T) {
	t.Parallel()

	repo := memory.NewUserRepository()
	v := validator.New()
	svc := NewService(repo, v)
	ctx := context.Background()

	// Create first user
	_, err := svc.Create(ctx, CreateInput{
		Username: "alice",
		Email:    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	// Try creating with duplicate username
	_, err = svc.Create(ctx, CreateInput{
		Username: "alice",
		Email:    "another@example.com",
	})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("expected duplicate username to return ErrAlreadyExists, got %v", err)
	}

	// Try creating with duplicate email
	_, err = svc.Create(ctx, CreateInput{
		Username: "bob",
		Email:    "alice@example.com",
	})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("expected duplicate email to return ErrAlreadyExists, got %v", err)
	}
}

func TestService_Get(t *testing.T) {
	t.Parallel()

	repo := memory.NewUserRepository()
	v := validator.New()
	svc := NewService(repo, v)
	ctx := context.Background()

	// Get non-existent
	_, err := svc.Get(ctx, "non-existent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for non-existent, got %v", err)
	}

	// Create and Get
	u, err := svc.Create(ctx, CreateInput{
		Username: "alice",
		Email:    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	got, err := svc.Get(ctx, u.ID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if got.ID != u.ID || got.Username != u.Username || got.Email != u.Email {
		t.Errorf("mismatch: expected %+v, got %+v", u, got)
	}
}

func TestService_List(t *testing.T) {
	t.Parallel()

	repo := memory.NewUserRepository()
	v := validator.New()
	svc := NewService(repo, v)
	ctx := context.Background()

	// List empty
	users, err := svc.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("list empty: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected empty list, got %d items", len(users))
	}

	// Populate 5 users
	names := []string{"u1", "u2", "u3", "u4", "u5"}
	for _, name := range names {
		_, err := svc.Create(ctx, CreateInput{
			Username: name,
			Email:    name + "@example.com",
		})
		if err != nil {
			t.Fatalf("failed to create user %s: %v", name, err)
		}
	}

	tests := []struct {
		name           string
		limit          int32
		offset         int32
		expectedLength int
	}{
		{
			name:           "default limit and offset",
			limit:          0,
			offset:         0,
			expectedLength: 5,
		},
		{
			name:           "limit negative",
			limit:          -5,
			offset:         0,
			expectedLength: 5,
		},
		{
			name:           "offset negative",
			limit:          2,
			offset:         -1,
			expectedLength: 2,
		},
		{
			name:           "limit too large",
			limit:          200,
			offset:         0,
			expectedLength: 5, // defaults to 20, which covers all 5
		},
		{
			name:           "offset skips",
			limit:          3,
			offset:         2,
			expectedLength: 3,
		},
		{
			name:           "offset out of range",
			limit:          3,
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
