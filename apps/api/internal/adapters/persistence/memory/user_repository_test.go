package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/starterpack/api/internal/domain"
	userdomain "github.com/starterpack/api/internal/domain/user"
)

// newTestUser is a helper that builds a domain user with a deterministic timestamp.
func newTestUser(t *testing.T, id, username, email string, createdAt time.Time) *userdomain.User {
	t.Helper()
	return userdomain.New(id, username, email, createdAt)
}

func TestUserRepository_CreateAndGetByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name     string
		id       string
		username string
		email    string
	}{
		{
			name:     "round-trip stores and retrieves user",
			id:       "u-1",
			username: "alice",
			email:    "alice@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := NewUserRepository()
			now := time.Now()
			u := newTestUser(t, tt.id, tt.username, tt.email, now)

			if err := repo.Create(ctx, u); err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}

			got, err := repo.GetByID(ctx, tt.id)
			if err != nil {
				t.Fatalf("GetByID() unexpected error: %v", err)
			}
			if got.ID != tt.id {
				t.Errorf("ID = %q, want %q", got.ID, tt.id)
			}
			if got.Username != tt.username {
				t.Errorf("Username = %q, want %q", got.Username, tt.username)
			}
			if got.Email != tt.email {
				t.Errorf("Email = %q, want %q", got.Email, tt.email)
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

func TestUserRepository_CreateDuplicate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name   string
		first  *userdomain.User
		second *userdomain.User
	}{
		{
			name:   "duplicate username returns ErrAlreadyExists",
			first:  userdomain.New("u-1", "alice", "alice@example.com", now),
			second: userdomain.New("u-2", "alice", "other@example.com", now),
		},
		{
			name:   "duplicate email returns ErrAlreadyExists",
			first:  userdomain.New("u-1", "alice", "alice@example.com", now),
			second: userdomain.New("u-2", "bob", "alice@example.com", now),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := NewUserRepository()

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

func TestUserRepository_GetByID_NotFound(t *testing.T) {
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
			repo := NewUserRepository()

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

func TestUserRepository_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Three users with increasing CreatedAt so we can verify desc sort.
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)

	threeUsers := []*userdomain.User{
		userdomain.New("u-1", "alice", "alice@example.com", t1),
		userdomain.New("u-2", "bob", "bob@example.com", t2),
		userdomain.New("u-3", "carol", "carol@example.com", t3),
	}

	fiveUsers := []*userdomain.User{
		userdomain.New("u-1", "alice", "alice@example.com", t1),
		userdomain.New("u-2", "bob", "bob@example.com", t2),
		userdomain.New("u-3", "carol", "carol@example.com", t3),
		userdomain.New("u-4", "dave", "dave@example.com", t3.Add(time.Hour)),
		userdomain.New("u-5", "eve", "eve@example.com", t3.Add(2*time.Hour)),
	}

	tests := []struct {
		name      string
		seed      []*userdomain.User
		limit     int32
		offset    int32
		wantCount int
		wantIDs   []string // expected IDs in order (desc by CreatedAt); nil to skip
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
			seed:      threeUsers,
			limit:     10,
			offset:    0,
			wantCount: 3,
			wantIDs:   []string{"u-3", "u-2", "u-1"},
		},
		{
			name:      "limit restricts results",
			seed:      fiveUsers,
			limit:     2,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "offset skips results",
			seed:      threeUsers,
			limit:     10,
			offset:    1,
			wantCount: 2,
		},
		{
			name:      "offset beyond length returns empty slice",
			seed:      threeUsers,
			limit:     10,
			offset:    100,
			wantCount: 0,
		},
		{
			name:      "limit zero returns empty slice",
			seed:      threeUsers,
			limit:     0,
			offset:    0,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := NewUserRepository()
			for _, u := range tt.seed {
				if err := repo.Create(ctx, u); err != nil {
					t.Fatalf("seed Create(%s) unexpected error: %v", u.ID, err)
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
