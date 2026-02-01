package routes

import (
	"github.com/labstack/echo/v4"
)

func RegisterCourierRoutes(e *echo.Group, h courierHandler) {
	couriers := e.Group("/couriers")

	couriers.GET("/:id", h.GetByID)
	couriers.GET("", h.GetAll)
	couriers.POST("", h.Create)
	couriers.PUT("", h.Update)
}
