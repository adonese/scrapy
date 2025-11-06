package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// ErrorHandler returns a middleware that handles errors consistently
func ErrorHandler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				// Check if it's already an echo.HTTPError
				if he, ok := err.(*echo.HTTPError); ok {
					code := he.Code
					message := he.Message

					// Convert message to string if needed
					var msg string
					if m, ok := message.(string); ok {
						msg = m
					} else {
						msg = "An error occurred"
					}

					return c.JSON(code, ErrorResponse{
						Error:   http.StatusText(code),
						Message: msg,
						Code:    code,
					})
				}

				// For other errors, return 500
				return c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Internal Server Error",
					Message: err.Error(),
					Code:    http.StatusInternalServerError,
				})
			}
			return nil
		}
	}
}
