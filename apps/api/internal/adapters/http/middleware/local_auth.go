package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/starterpack/api/internal/platform/jwtutil"
)

// LocalAuth verifies local JWT tokens in the Authorization header and stores
// the account ID in the request context. Mount only on protected routes when
// Clerk is not configured.
func LocalAuth(jwt *jwtutil.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c)
		if token == "" {
			abortUnauthorized(c, "missing bearer token")
			return
		}
		claims, err := jwt.Verify(token)
		if err != nil {
			abortUnauthorized(c, "invalid or expired token")
			return
		}
		c.Set(ContextUserIDKey, claims.AccountID)
		c.Next()
	}
}
