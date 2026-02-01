package routes

import "github.com/labstack/echo/v4"

type courierHandler interface {
	GetByID(c echo.Context) error
	GetAll(c echo.Context) error
	Create(c echo.Context) error
	Update(c echo.Context) error
}

type deliveryHandler interface {
	Assign(c echo.Context) error
	Unassign(c echo.Context) error
}
