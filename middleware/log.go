package middleware

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var attributes = map[string]string{
	"X-Amz-Cf-Id": "correlationId",
}

func LogMiddleware(logger *zap.Logger) gin.HandlerFunc {
	config := ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        false,
		SkipPaths:  nil,
		Context: func(c *gin.Context) []zapcore.Field {
			var result []zapcore.Field
			for key, value := range attributes {
				if c.GetHeader(key) != "" {
					result = append(result, zap.String(value, c.GetHeader(key)))
				}
			}
			return result
		},
	}
	return ginzap.GinzapWithConfig(logger, &config)
}
func defaultHandleRecovery(c *gin.Context, err interface{}) {
	c.AbortWithStatus(http.StatusInternalServerError)
}

func LogRecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	var stack = true
	var recovery = defaultHandleRecovery
	config := ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		SkipPaths:  nil,
		Context: func(c *gin.Context) []zapcore.Field {
			var result []zapcore.Field
			for key, value := range attributes {
				if c.GetHeader(key) != "" {
					result = append(result, zap.String(value, c.GetHeader(key)))
				}
			}
			return result
		},
	}
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				end := time.Now()
				path := c.Request.URL.Path
				query := c.Request.URL.RawQuery
				fields := []zapcore.Field{
					zap.Int("status", 500),
					zap.String("method", c.Request.Method),
					zap.String("path", path),
					zap.String("query", query),
					zap.String("user-agent", c.Request.UserAgent()),
					zap.Any("error", err),
					zap.String("request", string(httpRequest)),
				}
				if config.TimeFormat != "" {
					fields = append(fields, zap.String("time", end.Format(config.TimeFormat)))
				}
				if config.Context != nil {
					fields = append(fields, config.Context(c)...)
				}
				if stack {
					fields = append(fields, zap.String("stack", string(debug.Stack())))
				}

				if brokenPipe {
					logger.Error("[Recovery Broken pipe]",
						fields...,
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}
				logger.Error("[Recovery from panic]",
					fields...,
				)
				recovery(c, err)
			}
		}()
		c.Next()
	}
}
