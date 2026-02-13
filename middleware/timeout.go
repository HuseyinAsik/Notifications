package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(c.Request.Context(), timeout*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctxWithTimeout)
		c.Next()
	}
}
