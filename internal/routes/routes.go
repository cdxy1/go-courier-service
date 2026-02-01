package routes

import (
	"github.com/labstack/echo/v4"
)

type Routes struct {
	CourierHandler  courierHandler
	DeliveryHandler deliveryHandler
	APIMiddlewares  []echo.MiddlewareFunc
}

func NewRoutes(c courierHandler, d deliveryHandler, apiMiddlewares ...echo.MiddlewareFunc) *Routes {
	return &Routes{CourierHandler: c, DeliveryHandler: d, APIMiddlewares: apiMiddlewares}
}

func (r *Routes) Register(e *echo.Echo) {
	api := e.Group("/api/v1", r.APIMiddlewares...)

	RegisterHealthRoutes(api)
	RegisterCourierRoutes(api, r.CourierHandler)
	RegisterDeliveryRoutes(api, r.DeliveryHandler)
}
