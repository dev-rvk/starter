package domain

import (
	"errors"
	"testing"
)

func TestNotFound(t *testing.T) {
	t.Helper()

	tests := []struct {
		name       string
		entity     string
		wantKind   ErrorKind
		wantEntity string
		wantMsg    string
		wantField  string
	}{
		{
			name:       "user not found",
			entity:     "user",
			wantKind:   KindNotFound,
			wantEntity: "user",
			wantMsg:    "user not found",
			wantField:  "",
		},
		{
			name:       "todo not found",
			entity:     "todo",
			wantKind:   KindNotFound,
			wantEntity: "todo",
			wantMsg:    "todo not found",
			wantField:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NotFound(tt.entity)

			if err.Kind != tt.wantKind {
				t.Errorf("Kind = %d, want %d", err.Kind, tt.wantKind)
			}
			if err.Entity != tt.wantEntity {
				t.Errorf("Entity = %q, want %q", err.Entity, tt.wantEntity)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", err.Message, tt.wantMsg)
			}
			if err.Field != tt.wantField {
				t.Errorf("Field = %q, want %q", err.Field, tt.wantField)
			}
			if err.err != ErrNotFound {
				t.Errorf("wrapped error = %v, want ErrNotFound", err.err)
			}
		})
	}
}

func TestAlreadyExists(t *testing.T) {
	t.Helper()

	tests := []struct {
		name       string
		entity     string
		wantKind   ErrorKind
		wantEntity string
		wantMsg    string
		wantField  string
	}{
		{
			name:       "user already exists",
			entity:     "user",
			wantKind:   KindAlreadyExists,
			wantEntity: "user",
			wantMsg:    "user already exists",
			wantField:  "",
		},
		{
			name:       "todo already exists",
			entity:     "todo",
			wantKind:   KindAlreadyExists,
			wantEntity: "todo",
			wantMsg:    "todo already exists",
			wantField:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AlreadyExists(tt.entity)

			if err.Kind != tt.wantKind {
				t.Errorf("Kind = %d, want %d", err.Kind, tt.wantKind)
			}
			if err.Entity != tt.wantEntity {
				t.Errorf("Entity = %q, want %q", err.Entity, tt.wantEntity)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", err.Message, tt.wantMsg)
			}
			if err.Field != tt.wantField {
				t.Errorf("Field = %q, want %q", err.Field, tt.wantField)
			}
			if err.err != ErrAlreadyExists {
				t.Errorf("wrapped error = %v, want ErrAlreadyExists", err.err)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	t.Helper()

	tests := []struct {
		name       string
		entity     string
		field      string
		reason     string
		wantKind   ErrorKind
		wantEntity string
		wantField  string
		wantMsg    string
	}{
		{
			name:       "user email validation",
			entity:     "user",
			field:      "email",
			reason:     "invalid email format",
			wantKind:   KindValidation,
			wantEntity: "user",
			wantField:  "email",
			wantMsg:    "invalid email format",
		},
		{
			name:       "todo title validation",
			entity:     "todo",
			field:      "title",
			reason:     "title is required",
			wantKind:   KindValidation,
			wantEntity: "todo",
			wantField:  "title",
			wantMsg:    "title is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidationError(tt.entity, tt.field, tt.reason)

			if err.Kind != tt.wantKind {
				t.Errorf("Kind = %d, want %d", err.Kind, tt.wantKind)
			}
			if err.Entity != tt.wantEntity {
				t.Errorf("Entity = %q, want %q", err.Entity, tt.wantEntity)
			}
			if err.Field != tt.wantField {
				t.Errorf("Field = %q, want %q", err.Field, tt.wantField)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", err.Message, tt.wantMsg)
			}
			if err.err != ErrValidation {
				t.Errorf("wrapped error = %v, want ErrValidation", err.err)
			}
		})
	}
}

func TestErrorsIs(t *testing.T) {
	t.Helper()

	tests := []struct {
		name     string
		err      error
		sentinel error
		want     bool
	}{
		{
			name:     "NotFound wraps ErrNotFound",
			err:      NotFound("user"),
			sentinel: ErrNotFound,
			want:     true,
		},
		{
			name:     "AlreadyExists wraps ErrAlreadyExists",
			err:      AlreadyExists("user"),
			sentinel: ErrAlreadyExists,
			want:     true,
		},
		{
			name:     "ValidationError wraps ErrValidation",
			err:      ValidationError("user", "email", "bad email"),
			sentinel: ErrValidation,
			want:     true,
		},
		{
			name:     "NotFound does not match ErrAlreadyExists",
			err:      NotFound("user"),
			sentinel: ErrAlreadyExists,
			want:     false,
		},
		{
			name:     "AlreadyExists does not match ErrValidation",
			err:      AlreadyExists("user"),
			sentinel: ErrValidation,
			want:     false,
		},
		{
			name:     "ValidationError does not match ErrNotFound",
			err:      ValidationError("user", "email", "bad"),
			sentinel: ErrNotFound,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, tt.sentinel); got != tt.want {
				t.Errorf("errors.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorsAs(t *testing.T) {
	t.Helper()

	tests := []struct {
		name       string
		err        error
		wantKind   ErrorKind
		wantEntity string
	}{
		{
			name:       "NotFound extracts as *Error",
			err:        NotFound("user"),
			wantKind:   KindNotFound,
			wantEntity: "user",
		},
		{
			name:       "AlreadyExists extracts as *Error",
			err:        AlreadyExists("todo"),
			wantKind:   KindAlreadyExists,
			wantEntity: "todo",
		},
		{
			name:       "ValidationError extracts as *Error",
			err:        ValidationError("user", "email", "bad"),
			wantKind:   KindValidation,
			wantEntity: "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var domErr *Error
			if !errors.As(tt.err, &domErr) {
				t.Fatal("errors.As(*Error) returned false")
			}
			if domErr.Kind != tt.wantKind {
				t.Errorf("Kind = %d, want %d", domErr.Kind, tt.wantKind)
			}
			if domErr.Entity != tt.wantEntity {
				t.Errorf("Entity = %q, want %q", domErr.Entity, tt.wantEntity)
			}
		})
	}
}

func TestErrorMethod(t *testing.T) {
	t.Helper()

	tests := []struct {
		name    string
		err     *Error
		wantMsg string
	}{
		{
			name:    "NotFound Error() returns message",
			err:     NotFound("user"),
			wantMsg: "user not found",
		},
		{
			name:    "AlreadyExists Error() returns message",
			err:     AlreadyExists("todo"),
			wantMsg: "todo already exists",
		},
		{
			name:    "ValidationError Error() returns reason",
			err:     ValidationError("user", "email", "invalid email format"),
			wantMsg: "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	t.Helper()

	tests := []struct {
		name         string
		err          *Error
		wantSentinel error
	}{
		{
			name:         "NotFound unwraps to ErrNotFound",
			err:          NotFound("user"),
			wantSentinel: ErrNotFound,
		},
		{
			name:         "AlreadyExists unwraps to ErrAlreadyExists",
			err:          AlreadyExists("user"),
			wantSentinel: ErrAlreadyExists,
		},
		{
			name:         "ValidationError unwraps to ErrValidation",
			err:          ValidationError("user", "email", "bad"),
			wantSentinel: ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Unwrap(); got != tt.wantSentinel {
				t.Errorf("Unwrap() = %v, want %v", got, tt.wantSentinel)
			}
		})
	}
}

func TestKindUnknownIsZeroValue(t *testing.T) {
	var k ErrorKind
	if k != KindUnknown {
		t.Errorf("zero value of ErrorKind = %d, want KindUnknown (%d)", k, KindUnknown)
	}
	if KindUnknown != 0 {
		t.Errorf("KindUnknown = %d, want 0", KindUnknown)
	}
}
