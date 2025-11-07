package render

import (
    "net/http"

    "github.com/a-h/templ"
    "github.com/labstack/echo/v4"
)

// Component writes a templ component to the Echo response writer.
func Component(c echo.Context, status int, comp templ.Component) error {
    c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
    c.Response().WriteHeader(status)
    if err := comp.Render(c.Request().Context(), c.Response().Writer); err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }
    return nil
}
