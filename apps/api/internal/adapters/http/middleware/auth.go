package middleware

import (
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gin-gonic/gin"
)

// ContextUserIDKey is the Gin context key holding the authenticated Clerk user id.
const ContextUserIDKey = "userID"

// ClerkAuth verifies the Clerk session JWT in the Authorization header and
// stores the user id in the request context. Mount only on protected routes,
// and only after clerk.SetKey has been called (see composition root).
func ClerkAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c)
		if token == "" {
			abortUnauthorized(c, "missing bearer token")
			return
		}
		claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{Token: token})
		if err != nil {
			abortUnauthorized(c, "invalid or expired token")
			return
		}
		c.Set(ContextUserIDKey, claims.Subject)
		c.Next()
	}
}

func bearerToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

func abortUnauthorized(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
}
