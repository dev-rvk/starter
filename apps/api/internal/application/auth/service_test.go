package authapp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/starterpack/api/internal/adapters/persistence/memory"
	"github.com/starterpack/api/internal/domain"
	accountdomain "github.com/starterpack/api/internal/domain/account"
	"github.com/starterpack/api/internal/platform/jwtutil"
	"github.com/starterpack/api/internal/platform/validator"
)

func TestService_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		username string
		email    string
		password string
		wantErr  error
	}{
		{
			name:     "success",
			username: "alice",
			email:    "alice@example.com",
			password: "password123",
			wantErr:  nil,
		},
		{
			name:     "validation_error_username",
			username: "",
			email:    "alice@example.com",
			password: "password123",
			wantErr:  domain.ErrValidation,
		},
		{
			name:     "validation_error_email",
			username: "alice",
			email:    "invalid-email",
			password: "password123",
			wantErr:  domain.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts := memory.NewAccountRepository()
			users := memory.NewUserRepository()
			jwt := jwtutil.New("test-secret-key-that-is-long-enough", 24*time.Hour)
			v := validator.New()
			svc := NewService(accounts, users, jwt, v)

			res, err := svc.Register(context.Background(), RegisterInput{
				Username: tt.username,
				Email:    tt.email,
				Password: tt.password,
			})

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				if res != nil {
					t.Errorf("expected nil result on error, got %v", res)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if res == nil {
					t.Fatal("expected non-nil result")
				}
				if res.Token == "" {
					t.Error("expected non-empty token")
				}
				if res.User == nil {
					t.Fatal("expected non-nil user")
				}
				if res.User.Username != tt.username || res.User.Email != tt.email {
					t.Errorf("mismatch: %+v", res.User)
				}

				// Verify token works
				claims, err := jwt.Verify(res.Token)
				if err != nil {
					t.Fatalf("failed to verify token: %v", err)
				}
				if claims.AccountID != res.User.ID {
					t.Errorf("token account ID mismatch: expected %s, got %s", res.User.ID, claims.AccountID)
				}
			}
		})
	}
}

func TestService_Register_DuplicateEmail(t *testing.T) {
	t.Parallel()

	accounts := memory.NewAccountRepository()
	users := memory.NewUserRepository()
	jwt := jwtutil.New("test-secret-key-that-is-long-enough", 24*time.Hour)
	v := validator.New()
	svc := NewService(accounts, users, jwt, v)
	ctx := context.Background()

	_, err := svc.Register(ctx, RegisterInput{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("failed to register first user: %v", err)
	}

	_, err = svc.Register(ctx, RegisterInput{
		Username: "bob",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("expected duplicate email to return ErrAlreadyExists, got %v", err)
	}
}

func TestService_Login(t *testing.T) {
	t.Parallel()

	accounts := memory.NewAccountRepository()
	users := memory.NewUserRepository()
	jwt := jwtutil.New("test-secret-key-that-is-long-enough", 24*time.Hour)
	v := validator.New()
	svc := NewService(accounts, users, jwt, v)
	ctx := context.Background()

	// Register a user
	_, err := svc.Register(ctx, RegisterInput{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  error
	}{
		{
			name:     "success",
			email:    "alice@example.com",
			password: "password123",
			wantErr:  nil,
		},
		{
			name:     "wrong_password",
			email:    "alice@example.com",
			password: "wrongpassword",
			wantErr:  accountdomain.ErrNotFound, // returns ErrNotFound to avoid leaking email existence
		},
		{
			name:     "non_existent_email",
			email:    "nonexistent@example.com",
			password: "password123",
			wantErr:  accountdomain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := svc.Login(ctx, LoginInput{
				Email:    tt.email,
				Password: tt.password,
			})

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				if res != nil {
					t.Errorf("expected nil result, got %+v", res)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if res == nil {
					t.Fatal("expected non-nil result")
				}
				if res.Token == "" {
					t.Error("expected non-empty token")
				}
				if res.User == nil {
					t.Fatal("expected non-nil user")
				}
			}
		})
	}
}

func TestService_Me(t *testing.T) {
	t.Parallel()

	accounts := memory.NewAccountRepository()
	users := memory.NewUserRepository()
	jwt := jwtutil.New("test-secret-key-that-is-long-enough", 24*time.Hour)
	v := validator.New()
	svc := NewService(accounts, users, jwt, v)
	ctx := context.Background()

	// Me on non-existent
	_, err := svc.Me(ctx, "non-existent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for non-existent, got %v", err)
	}

	// Register and call Me
	res, err := svc.Register(ctx, RegisterInput{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	got, err := svc.Me(ctx, res.User.ID)
	if err != nil {
		t.Fatalf("failed to get me: %v", err)
	}
	if got.ID != res.User.ID || got.Username != res.User.Username {
		t.Errorf("mismatch: expected %+v, got %+v", res.User, got)
	}
}
