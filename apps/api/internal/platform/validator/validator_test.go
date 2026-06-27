package validator

import (
	"errors"
	"testing"

	"github.com/starterpack/api/internal/domain"
)

// testUser is a local struct used to avoid importing domain/user.
type testUser struct {
	ID       string `validate:"required"`
	Username string `validate:"required,username"`
	Email    string `validate:"required,email"`
}

func validTestUser() testUser {
	return testUser{
		ID:       "usr_123",
		Username: "alice",
		Email:    "alice@example.com",
	}
}

func TestValidateAndMap_ValidStruct(t *testing.T) {
	v := New()
	err := v.ValidateAndMap("user", validTestUser())
	if err != nil {
		t.Errorf("ValidateAndMap() = %v, want nil", err)
	}
}

func TestValidateAndMap_RequiredFields(t *testing.T) {
	t.Helper()

	tests := []struct {
		name  string
		input testUser
	}{
		{
			name:  "missing ID",
			input: testUser{Username: "alice", Email: "alice@example.com"},
		},
		{
			name:  "missing Username",
			input: testUser{ID: "usr_1", Email: "alice@example.com"},
		},
		{
			name:  "missing Email",
			input: testUser{ID: "usr_1", Username: "alice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			err := v.ValidateAndMap("user", tt.input)
			if err == nil {
				t.Fatal("ValidateAndMap() = nil, want error")
			}

			assertDomainValidationError(t, err)
		})
	}
}

func TestValidateAndMap_InvalidEmail(t *testing.T) {
	v := New()
	u := validTestUser()
	u.Email = "not-an-email"

	err := v.ValidateAndMap("user", u)
	if err == nil {
		t.Fatal("ValidateAndMap() = nil, want error for invalid email")
	}

	assertDomainValidationError(t, err)
}

func TestValidateAndMap_UsernameInvalid(t *testing.T) {
	t.Helper()

	tests := []struct {
		name     string
		username string
	}{
		{
			name:     "too short (1 char)",
			username: "a",
		},
		{
			name:     "too long (7 chars)",
			username: "abcdefg",
		},
		{
			name:     "special characters (dash)",
			username: "ab-cd",
		},
		{
			name:     "special characters (dot)",
			username: "ab.cd",
		},
		{
			name:     "special characters (at sign)",
			username: "ab@cd",
		},
		{
			name:     "space in username",
			username: "ab cd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			u := validTestUser()
			u.Username = tt.username

			err := v.ValidateAndMap("user", u)
			if err == nil {
				t.Fatalf("ValidateAndMap() = nil, want error for username %q", tt.username)
			}

			assertDomainValidationError(t, err)
		})
	}
}

func TestValidateAndMap_UsernameValid(t *testing.T) {
	t.Helper()

	tests := []struct {
		name     string
		username string
	}{
		{
			name:     "2 chars minimum",
			username: "ab",
		},
		{
			name:     "6 chars maximum",
			username: "abcdef",
		},
		{
			name:     "with underscore",
			username: "a_b",
		},
		{
			name:     "all digits",
			username: "1234",
		},
		{
			name:     "mixed alphanumeric and underscore",
			username: "a1_b2",
		},
		{
			name:     "uppercase letters",
			username: "AbCdE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			u := validTestUser()
			u.Username = tt.username

			err := v.ValidateAndMap("user", u)
			if err != nil {
				t.Errorf("ValidateAndMap() = %v, want nil for username %q", err, tt.username)
			}
		})
	}
}

func TestValidateAndMap_ErrorsIsSentinel(t *testing.T) {
	v := New()
	u := validTestUser()
	u.Email = "bad"

	err := v.ValidateAndMap("user", u)
	if err == nil {
		t.Fatal("ValidateAndMap() = nil, want error")
	}

	if !errors.Is(err, domain.ErrValidation) {
		t.Error("errors.Is(err, domain.ErrValidation) = false, want true")
	}
}

func TestValidateAndMap_ErrorsAs(t *testing.T) {
	v := New()
	u := validTestUser()
	u.Username = "x" // too short

	err := v.ValidateAndMap("user", u)
	if err == nil {
		t.Fatal("ValidateAndMap() = nil, want error")
	}

	var domErr *domain.Error
	if !errors.As(err, &domErr) {
		t.Fatal("errors.As(*domain.Error) = false, want true")
	}

	if domErr.Kind != domain.KindValidation {
		t.Errorf("Kind = %d, want %d (KindValidation)", domErr.Kind, domain.KindValidation)
	}
	if domErr.Entity != "user" {
		t.Errorf("Entity = %q, want %q", domErr.Entity, "user")
	}
	if domErr.Field == "" {
		t.Error("Field is empty, want non-empty field name")
	}
}

// assertDomainValidationError is a test helper that verifies an error is a
// domain.Error with KindValidation that wraps domain.ErrValidation.
func assertDomainValidationError(t *testing.T, err error) {
	t.Helper()

	var domErr *domain.Error
	if !errors.As(err, &domErr) {
		t.Fatal("errors.As(*domain.Error) = false, want true")
	}
	if domErr.Kind != domain.KindValidation {
		t.Errorf("Kind = %d, want %d (KindValidation)", domErr.Kind, domain.KindValidation)
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Error("errors.Is(err, domain.ErrValidation) = false, want true")
	}
}
