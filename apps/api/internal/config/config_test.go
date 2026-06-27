package config

import (
	"reflect"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear any env vars that could interfere with defaults.
	for _, key := range []string{
		"APP_ENV", "PORT", "LOG_LEVEL", "DATABASE_URL", "JWT_SECRET",
		"CORS_ORIGINS", "CLERK_SECRET_KEY", "STRIPE_SECRET_KEY",
		"STRIPE_WEBHOOK_SECRET", "SENTRY_DSN", "RESEND_TOKEN", "RESEND_FROM",
	} {
		t.Setenv(key, "")
	}

	cfg := Load()

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"Env", cfg.Env, "development"},
		{"Port", cfg.Port, "3002"},
		{"LogLevel", cfg.LogLevel, "info"},
		{"LogPretty (dev)", cfg.LogPretty, true},
		{"DatabaseURL empty", cfg.DatabaseURL, ""},
		{"JWTSecret default", cfg.JWTSecret, "starterpack-dev-secret-change-in-production"},
		{"Clerk disabled", cfg.Clerk.SecretKey, ""},
		{"Stripe disabled", cfg.Stripe.SecretKey, ""},
		{"Sentry disabled", cfg.Sentry.DSN, ""},
		{"Resend disabled", cfg.Resend.Token, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}

	// Verify default CORS origins.
	wantOrigins := []string{"http://localhost:3000", "http://localhost:3001"}
	if !reflect.DeepEqual(cfg.CORSOrigins, wantOrigins) {
		t.Errorf("CORSOrigins = %v, want %v", cfg.CORSOrigins, wantOrigins)
	}
}

func TestLoadWithEnvOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("PORT", "8080")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("DATABASE_URL", "postgres://localhost/testdb")
	t.Setenv("JWT_SECRET", "prod-secret-key")
	t.Setenv("CORS_ORIGINS", "https://example.com,https://api.example.com")
	t.Setenv("CLERK_SECRET_KEY", "sk_test_clerk")
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_stripe")
	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test")
	t.Setenv("SENTRY_DSN", "https://sentry.io/123")
	t.Setenv("RESEND_TOKEN", "re_test_token")
	t.Setenv("RESEND_FROM", "noreply@example.com")

	cfg := Load()

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"Env", cfg.Env, "production"},
		{"Port", cfg.Port, "8080"},
		{"LogLevel", cfg.LogLevel, "debug"},
		{"LogPretty (prod)", cfg.LogPretty, false},
		{"DatabaseURL", cfg.DatabaseURL, "postgres://localhost/testdb"},
		{"JWTSecret", cfg.JWTSecret, "prod-secret-key"},
		{"Clerk SecretKey", cfg.Clerk.SecretKey, "sk_test_clerk"},
		{"Stripe SecretKey", cfg.Stripe.SecretKey, "sk_test_stripe"},
		{"Stripe WebhookSecret", cfg.Stripe.WebhookSecret, "whsec_test"},
		{"Sentry DSN", cfg.Sentry.DSN, "https://sentry.io/123"},
		{"Resend Token", cfg.Resend.Token, "re_test_token"},
		{"Resend From", cfg.Resend.From, "noreply@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}

	wantOrigins := []string{"https://example.com", "https://api.example.com"}
	if !reflect.DeepEqual(cfg.CORSOrigins, wantOrigins) {
		t.Errorf("CORSOrigins = %v, want %v", cfg.CORSOrigins, wantOrigins)
	}
}

func TestLogPrettyProductionMode(t *testing.T) {
	t.Setenv("APP_ENV", "production")

	// Clear other vars so they don't interfere.
	for _, key := range []string{
		"PORT", "LOG_LEVEL", "DATABASE_URL", "JWT_SECRET", "CORS_ORIGINS",
		"CLERK_SECRET_KEY", "STRIPE_SECRET_KEY", "STRIPE_WEBHOOK_SECRET",
		"SENTRY_DSN", "RESEND_TOKEN", "RESEND_FROM",
	} {
		t.Setenv(key, "")
	}

	cfg := Load()
	if cfg.LogPretty {
		t.Error("LogPretty = true in production, want false")
	}
}

func TestFeatureToggleEnabled(t *testing.T) {
	t.Helper()

	tests := []struct {
		name    string
		enabled func() bool
		want    bool
	}{
		{"ClerkConfig empty", func() bool { return ClerkConfig{SecretKey: ""}.Enabled() }, false},
		{"ClerkConfig set", func() bool { return ClerkConfig{SecretKey: "sk_key"}.Enabled() }, true},
		{"StripeConfig empty", func() bool { return StripeConfig{SecretKey: ""}.Enabled() }, false},
		{"StripeConfig set", func() bool { return StripeConfig{SecretKey: "sk_stripe"}.Enabled() }, true},
		{"SentryConfig empty", func() bool { return SentryConfig{DSN: ""}.Enabled() }, false},
		{"SentryConfig set", func() bool { return SentryConfig{DSN: "https://sentry"}.Enabled() }, true},
		{"ResendConfig empty", func() bool { return ResendConfig{Token: ""}.Enabled() }, false},
		{"ResendConfig set", func() bool { return ResendConfig{Token: "re_tok"}.Enabled() }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.enabled(); got != tt.want {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasDatabase(t *testing.T) {
	t.Helper()

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"empty URL", "", false},
		{"set URL", "postgres://localhost/db", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{DatabaseURL: tt.url}
			if got := cfg.HasDatabase(); got != tt.want {
				t.Errorf("HasDatabase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCORSOriginsParsing(t *testing.T) {
	t.Helper()

	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{
			name: "single origin",
			raw:  "https://example.com",
			want: []string{"https://example.com"},
		},
		{
			name: "multiple origins",
			raw:  "https://a.com,https://b.com,https://c.com",
			want: []string{"https://a.com", "https://b.com", "https://c.com"},
		},
		{
			name: "with spaces around commas",
			raw:  "https://a.com , https://b.com , https://c.com",
			want: []string{"https://a.com", "https://b.com", "https://c.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env and load to test full pipeline.
			t.Setenv("CORS_ORIGINS", tt.raw)
			// Clear APP_ENV so other defaults don't interfere.
			t.Setenv("APP_ENV", "development")

			cfg := Load()
			if !reflect.DeepEqual(cfg.CORSOrigins, tt.want) {
				t.Errorf("CORSOrigins = %v, want %v", cfg.CORSOrigins, tt.want)
			}
		})
	}
}

func TestSplitCSV(t *testing.T) {
	t.Helper()

	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{
			name: "empty string",
			raw:  "",
			want: []string{},
		},
		{
			name: "single value",
			raw:  "hello",
			want: []string{"hello"},
		},
		{
			name: "multiple values",
			raw:  "a,b,c",
			want: []string{"a", "b", "c"},
		},
		{
			name: "spaces around values",
			raw:  " a , b , c ",
			want: []string{"a", "b", "c"},
		},
		{
			name: "trailing comma",
			raw:  "a,b,",
			want: []string{"a", "b"},
		},
		{
			name: "leading comma",
			raw:  ",a,b",
			want: []string{"a", "b"},
		},
		{
			name: "only commas",
			raw:  ",,,",
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitCSV(tt.raw)
			// Normalize nil vs empty slice for comparison.
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitCSV(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestGetenv(t *testing.T) {
	t.Helper()

	tests := []struct {
		name     string
		key      string
		envVal   string
		fallback string
		want     string
	}{
		{
			name:     "returns fallback when env empty",
			key:      "TEST_GETENV_EMPTY",
			envVal:   "",
			fallback: "default_val",
			want:     "default_val",
		},
		{
			name:     "returns env value when set",
			key:      "TEST_GETENV_SET",
			envVal:   "custom_val",
			fallback: "default_val",
			want:     "custom_val",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.key, tt.envVal)

			got := getenv(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("getenv(%q, %q) = %q, want %q", tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}
