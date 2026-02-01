package routes

import (
	"github.com/labstack/echo/v4"
)

func RegisterDeliveryRoutes(e *echo.Group, h deliveryHandler) {
	delivery := e.Group("/delivery")

	delivery.POST("/assign", h.Assign)
	delivery.POST("/unassign", h.Unassign)
}
