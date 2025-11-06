package middleware

import (
	"strconv"

	"github.com/adonese/cost-of-living/pkg/metrics"
	"github.com/labstack/echo/v4"
)

// MetricsMiddleware tracks HTTP requests in Prometheus
func MetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Execute request
			err := next(c)

			// Record metric
			status := strconv.Itoa(c.Response().Status)
			metrics.HTTPRequestsTotal.WithLabelValues(
				c.Request().Method,
				c.Path(),
				status,
			).Inc()

			return err
		}
	}
}
