package jwtutil

import (
	"testing"
	"time"
)

func TestSign(t *testing.T) {
	t.Helper()

	tests := []struct {
		name      string
		secret    string
		duration  time.Duration
		accountID string
	}{
		{
			name:      "produces non-empty token",
			secret:    "test-secret",
			duration:  time.Hour,
			accountID: "acc_123",
		},
		{
			name:      "different account ID",
			secret:    "another-secret",
			duration:  30 * time.Minute,
			accountID: "acc_456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := New(tt.secret, tt.duration)
			token, err := mgr.Sign(tt.accountID)
			if err != nil {
				t.Fatalf("Sign() error = %v", err)
			}
			if token == "" {
				t.Error("Sign() returned empty token")
			}
		})
	}
}

func TestSignVerifyRoundTrip(t *testing.T) {
	t.Helper()

	tests := []struct {
		name      string
		secret    string
		duration  time.Duration
		accountID string
	}{
		{
			name:      "standard round-trip",
			secret:    "my-hmac-secret",
			duration:  time.Hour,
			accountID: "user_abc",
		},
		{
			name:      "uuid account ID",
			secret:    "secret-key-256",
			duration:  24 * time.Hour,
			accountID: "550e8400-e29b-41d4-a716-446655440000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := New(tt.secret, tt.duration)

			token, err := mgr.Sign(tt.accountID)
			if err != nil {
				t.Fatalf("Sign() error = %v", err)
			}

			claims, err := mgr.Verify(token)
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}

			if claims.AccountID != tt.accountID {
				t.Errorf("AccountID = %q, want %q", claims.AccountID, tt.accountID)
			}
			if claims.Subject != tt.accountID {
				t.Errorf("Subject = %q, want %q", claims.Subject, tt.accountID)
			}
		})
	}
}

func TestVerifyWrongSecret(t *testing.T) {
	signer := New("correct-secret", time.Hour)
	verifier := New("wrong-secret", time.Hour)

	token, err := signer.Sign("acc_123")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	_, err = verifier.Verify(token)
	if err == nil {
		t.Error("Verify() with wrong secret should return error")
	}
}

func TestVerifyExpiredToken(t *testing.T) {
	mgr := New("test-secret", time.Millisecond)

	token, err := mgr.Sign("acc_123")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	_, err = mgr.Verify(token)
	if err == nil {
		t.Error("Verify() with expired token should return error")
	}
}

func TestVerifyInvalidTokens(t *testing.T) {
	t.Helper()

	mgr := New("test-secret", time.Hour)

	tests := []struct {
		name     string
		tokenStr string
	}{
		{
			name:     "garbage string",
			tokenStr: "not-a-jwt-token-at-all",
		},
		{
			name:     "empty string",
			tokenStr: "",
		},
		{
			name:     "partial JWT format",
			tokenStr: "eyJhbGciOiJIUzI1NiJ9.garbage.garbage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mgr.Verify(tt.tokenStr)
			if err == nil {
				t.Error("Verify() should return error for invalid token")
			}
		})
	}
}
