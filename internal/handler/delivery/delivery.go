package delivery

import (
	"errors"
	"net/http"

	handlerErrors "github.com/cdxy1/go-courier-service/internal/handler/errors"
	courierRepo "github.com/cdxy1/go-courier-service/internal/repository/courier"
	deliveryRepo "github.com/cdxy1/go-courier-service/internal/repository/delivery"
	"github.com/labstack/echo/v4"
)

type DeliveryHandler struct {
	uc deliveryUsecase
}

func NewDeliveryHandler(uc deliveryUsecase) *DeliveryHandler {
	return &DeliveryHandler{uc: uc}
}

func (h *DeliveryHandler) Assign(c echo.Context) error {
	var orderIdRequest assignRequest

	if err := c.Bind(&orderIdRequest); err != nil || orderIdRequest.OrderId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": handlerErrors.ErrBadRequest.Error()})
	}

	delivery, courier, err := h.uc.Assign(c.Request().Context(), orderIdRequest.OrderId)
	if err != nil {
		if errors.Is(err, courierRepo.ErrCourierNotFound) || errors.Is(err, deliveryRepo.ErrDeliveryNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	assignResponse := &assignResponse{
		CourierId:        courier.ID,
		OrderID:          delivery.OrderId,
		TransportType:    courier.TransportType,
		DeliveryDeadline: delivery.Deadline,
	}

	return c.JSON(http.StatusOK, assignResponse)
}

func (h *DeliveryHandler) Unassign(c echo.Context) error {
	var orderIdRequest unassignRequest

	if err := c.Bind(&orderIdRequest); err != nil || orderIdRequest.OrderId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": handlerErrors.ErrBadRequest.Error()})
	}

	unassignResult, err := h.uc.Unassign(c.Request().Context(), orderIdRequest.OrderId)
	if err != nil {
		if errors.Is(err, deliveryRepo.ErrDeliveryNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	unassignResponse := &unassignResponse{
		OrderId:   unassignResult.OrderId,
		Status:    "unassigned",
		CourierId: unassignResult.CourierId,
	}

	return c.JSON(http.StatusOK, unassignResponse)
}
