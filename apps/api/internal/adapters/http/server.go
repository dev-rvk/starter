// Package http is the inbound HTTP adapter (Gin). Handlers translate requests
// into use-case calls and back; no business logic lives here.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	"github.com/starterpack/api/internal/adapters/http/middleware"
	"github.com/starterpack/api/internal/config"
)

// ServerDeps are the dependencies needed to build the router.
type ServerDeps struct {
	Config      config.Config
	Logger      zerolog.Logger
	UserHandler *UserHandler
}

// NewRouter builds the Gin engine with middleware and routes.
func NewRouter(deps ServerDeps) *gin.Engine {
	if deps.Config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(deps.Logger))
	r.Use(middleware.CORS(deps.Config.CORSOrigins))

	// Liveness probe (always public).
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// OpenAPI UI at /swagger/index.html (spec registered by the generated docs
	// package imported in the composition root).
	r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	api := r.Group("/api/v1")
	if deps.Config.Clerk.Enabled() {
		api.Use(middleware.ClerkAuth())
	} else {
		deps.Logger.Warn().
			Msg("Clerk disabled (CLERK_SECRET_KEY unset): /api/v1 routes are UNPROTECTED")
	}

	deps.UserHandler.register(api)

	return r
}
