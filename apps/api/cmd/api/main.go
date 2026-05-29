// Command api is the composition root for the starterpack backend. It loads
// config, builds adapters, wires the use cases and starts the HTTP server.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	clerk "github.com/clerk/clerk-sdk-go/v2"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	_ "github.com/starterpack/api/docs" // generated OpenAPI spec (swag init)
	httpadapter "github.com/starterpack/api/internal/adapters/http"
	"github.com/starterpack/api/internal/adapters/persistence/memory"
	"github.com/starterpack/api/internal/adapters/persistence/postgres"
	userapp "github.com/starterpack/api/internal/application/user"
	"github.com/starterpack/api/internal/config"
	userdomain "github.com/starterpack/api/internal/domain/user"
	"github.com/starterpack/api/internal/platform/logger"
)

// @title                      Starterpack API
// @version                    0.1.0
// @description                Hexagonal Go backend for the starterpack monorepo.
// @BasePath                   /api/v1
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
func main() {
	_ = godotenv.Load(".env.local", ".env")

	cfg := config.Load()
	log := logger.New(cfg.LogLevel, cfg.LogPretty)

	if cfg.Clerk.Enabled() {
		clerk.SetKey(cfg.Clerk.SecretKey)
		log.Info().Msg("feature: Clerk auth ENABLED")
	} else {
		log.Warn().Msg("feature: Clerk auth DISABLED (set CLERK_SECRET_KEY)")
	}
	logFeature(log, "Stripe", cfg.Stripe.Enabled())
	logFeature(log, "Sentry", cfg.Sentry.Enabled())
	logFeature(log, "Resend", cfg.Resend.Enabled())

	// Select the persistence adapter behind the domain Repository port.
	var userRepo userdomain.Repository
	if cfg.HasDatabase() {
		pool, err := postgres.NewPool(context.Background(), cfg.DatabaseURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to database")
		}
		defer pool.Close()
		userRepo = postgres.NewUserRepository(pool)
		log.Info().Msg("persistence: PostgreSQL (pgx + sqlc)")
	} else {
		userRepo = memory.NewUserRepository()
		log.Warn().Msg("persistence: in-memory (set DATABASE_URL to use PostgreSQL)")
	}

	userService := userapp.NewService(userRepo)
	router := httpadapter.NewRouter(httpadapter.ServerDeps{
		Config:      cfg,
		Logger:      log,
		UserHandler: httpadapter.NewUserHandler(userService),
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Str("env", cfg.Env).Msg("API listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutting down…")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("forced shutdown")
	}
}

func logFeature(log zerolog.Logger, name string, enabled bool) {
	if enabled {
		log.Info().Msgf("feature: %s ENABLED", name)
	} else {
		log.Info().Msgf("feature: %s disabled", name)
	}
}
