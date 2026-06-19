// Package config loads runtime configuration from the environment. Every
// optional integration is a feature toggle: it is considered enabled only when
// its key(s) are present, so the service boots cleanly with them absent.
package config

import (
	"os"
	"strings"
)

// Config is the fully-resolved application configuration.
type Config struct {
	Env         string // "development" | "production"
	Port        string
	LogLevel    string
	LogPretty   bool
	DatabaseURL string
	CORSOrigins []string

	Clerk  ClerkConfig
	Stripe StripeConfig
	Sentry SentryConfig
	Resend ResendConfig
}

// ClerkConfig holds authentication settings. Auth is REQUIRED in this project,
// but the server still boots without it (endpoints requiring auth will reject).
type ClerkConfig struct {
	SecretKey string
}

func (c ClerkConfig) Enabled() bool { return c.SecretKey != "" }

// StripeConfig holds payment settings (off until keys are provided).
type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
}

func (c StripeConfig) Enabled() bool { return c.SecretKey != "" }

// SentryConfig holds error-tracking settings (off until DSN is provided).
type SentryConfig struct {
	DSN string
}

func (c SentryConfig) Enabled() bool { return c.DSN != "" }

// ResendConfig holds transactional-email settings (off until token is provided).
type ResendConfig struct {
	Token string
	From  string
}

func (c ResendConfig) Enabled() bool { return c.Token != "" }

// HasDatabase reports whether a database connection should be attempted.
func (c Config) HasDatabase() bool { return c.DatabaseURL != "" }

// Load reads configuration from environment variables, applying defaults.
func Load() Config {
	cfg := Config{
		Env:         getenv("APP_ENV", "development"),
		Port:        getenv("PORT", "3002"),
		LogLevel:    getenv("LOG_LEVEL", "info"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		CORSOrigins: splitCSV(getenv("CORS_ORIGINS", "http://localhost:3000,http://localhost:3001")),
		Clerk:       ClerkConfig{SecretKey: os.Getenv("CLERK_SECRET_KEY")},
		Stripe: StripeConfig{
			SecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
			WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		},
		Sentry: SentryConfig{DSN: os.Getenv("SENTRY_DSN")},
		Resend: ResendConfig{
			Token: os.Getenv("RESEND_TOKEN"),
			From:  os.Getenv("RESEND_FROM"),
		},
	}
	cfg.LogPretty = cfg.Env == "development"
	return cfg
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}
