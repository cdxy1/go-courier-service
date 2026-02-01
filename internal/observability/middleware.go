package observability

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)

	rateLimitExceededTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "Total number of rate limit exceeded responses.",
		},
	)

	gatewayRetriesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gateway_retries_total",
			Help: "Total number of gateway retry attempts.",
		},
	)

	registerMetricsOnce sync.Once
	requestLogger       = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
)

func MetricsAndLogging() echo.MiddlewareFunc {
	RegisterMetrics()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			status := c.Response().Status
			if status == 0 {
				status = http.StatusOK
			}

			path := c.Path()
			if path == "" {
				path = c.Request().URL.Path
			}

			method := c.Request().Method
			duration := time.Since(start)

			httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
			httpRequestDuration.WithLabelValues(path).Observe(duration.Seconds())

			requestLogger.Printf("method=%s path=%s status=%d duration=%s", method, path, status, duration)

			return err
		}
	}
}

func RegisterMetrics() {
	registerMetricsOnce.Do(func() {
		prometheus.MustRegister(
			httpRequestsTotal,
			httpRequestDuration,
			rateLimitExceededTotal,
			gatewayRetriesTotal,
		)
	})
}

func IncRateLimitExceeded() {
	RegisterMetrics()
	rateLimitExceededTotal.Inc()
}

func IncGatewayRetries() {
	RegisterMetrics()
	gatewayRetriesTotal.Inc()
}
