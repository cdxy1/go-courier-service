package routes

import (
	"github.com/labstack/echo/v4"
)

func RegisterHealthRoutes(e *echo.Group) {
	health := e.Group("")

	health.GET("/ping", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"message": "pong",
		})
	})

	health.HEAD("/healthcheck", func(c echo.Context) error {
		return c.NoContent(204)
	})
}
