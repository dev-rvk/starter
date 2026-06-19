// Package middleware holds Gin middleware adapters.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Logger logs each request using zerolog.
func Logger(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		evt := log.Info()
		if c.Writer.Status() >= 500 {
			evt = log.Error()
		} else if c.Writer.Status() >= 400 {
			evt = log.Warn()
		}
		evt.
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Str("ip", c.ClientIP()).
			Msg("request")
	}
}
