package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/starterpack/api/internal/domain"
	accountdomain "github.com/starterpack/api/internal/domain/account"
)

// newTestAccount is a helper that builds a domain account with a deterministic timestamp.
func newTestAccount(t *testing.T, id, email, passwordHash string, createdAt time.Time) *accountdomain.Account {
	t.Helper()
	return accountdomain.New(id, email, passwordHash, createdAt)
}

func TestAccountRepository_CreateAndGetByEmail(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name         string
		id           string
		email        string
		passwordHash string
	}{
		{
			name:         "round-trip stores and retrieves account",
			id:           "a-1",
			email:        "alice@example.com",
			passwordHash: "$2a$10$hashedpassword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := NewAccountRepository()
			now := time.Now()
			acct := newTestAccount(t, tt.id, tt.email, tt.passwordHash, now)

			if err := repo.Create(ctx, acct); err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}

			got, err := repo.GetByEmail(ctx, tt.email)
			if err != nil {
				t.Fatalf("GetByEmail() unexpected error: %v", err)
			}
			if got.ID != tt.id {
				t.Errorf("ID = %q, want %q", got.ID, tt.id)
			}
			if got.Email != tt.email {
				t.Errorf("Email = %q, want %q", got.Email, tt.email)
			}
			if got.PasswordHash != tt.passwordHash {
				t.Errorf("PasswordHash = %q, want %q", got.PasswordHash, tt.passwordHash)
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

func TestAccountRepository_CreateDuplicateEmail(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name   string
		first  *accountdomain.Account
		second *accountdomain.Account
	}{
		{
			name:   "duplicate email returns ErrAlreadyExists",
			first:  accountdomain.New("a-1", "alice@example.com", "$2a$10$hash1", now),
			second: accountdomain.New("a-2", "alice@example.com", "$2a$10$hash2", now),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := NewAccountRepository()

			if err := repo.Create(ctx, tt.first); err != nil {
				t.Fatalf("Create(first) unexpected error: %v", err)
			}

			err := repo.Create(ctx, tt.second)
			if err == nil {
				t.Fatal("Create(second) expected error, got nil")
			}
			if !errors.Is(err, domain.ErrAlreadyExists) {
				t.Errorf("error = %v, want errors.Is(domain.ErrAlreadyExists)", err)
			}
		})
	}
}

func TestAccountRepository_GetByEmail_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "non-existent email returns ErrNotFound",
			email: "nobody@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := NewAccountRepository()

			_, err := repo.GetByEmail(ctx, tt.email)
			if err == nil {
				t.Fatal("GetByEmail() expected error, got nil")
			}
			if !errors.Is(err, domain.ErrNotFound) {
				t.Errorf("error = %v, want errors.Is(domain.ErrNotFound)", err)
			}
		})
	}
}
