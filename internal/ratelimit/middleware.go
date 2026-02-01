package ratelimit

import (
	"log"
	"net/http"
	"os"

	"github.com/cdxy1/go-courier-service/internal/observability"
	"github.com/labstack/echo/v4"
)

func Middleware(limiter Limiter, logger *log.Logger) echo.MiddlewareFunc {
	if limiter == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}
	if logger == nil {
		logger = log.New(os.Stdout, "[WARN] ", log.LstdFlags)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.RealIP()
			if key == "" {
				key = "unknown"
			}
			if !limiter.Allow(key) {
				observability.IncRateLimitExceeded()
				logger.Printf("rate limit exceeded ip=%s method=%s path=%s", key, c.Request().Method, c.Path())
				return c.NoContent(http.StatusTooManyRequests)
			}
			return next(c)
		}
	}
}
