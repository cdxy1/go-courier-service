package order_event

import (
	"context"
	"fmt"

	"github.com/cdxy1/go-courier-service/internal/model"
)

const (
	statusCreated   = "created"
	statusCancelled = "cancelled"
	statusCanceled  = "canceled"
	statusCompleted = "completed"
	statusDelivered = "delivered"
)

type deliveryUsecase interface {
	Assign(ctx context.Context, orderID string) (*model.DeliveryModel, *model.CourierModel, error)
	Unassign(ctx context.Context, orderID string) (*model.DeliveryModel, error)
	Complete(ctx context.Context, orderID string) (*model.DeliveryModel, error)
}

type createdHandler struct {
	uc deliveryUsecase
}

func (h *createdHandler) Handle(ctx context.Context, event model.OrderStatusEvent) error {
	_, _, err := h.uc.Assign(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("assign courier: %w", err)
	}
	return nil
}

type cancelledHandler struct {
	uc deliveryUsecase
}

func (h *cancelledHandler) Handle(ctx context.Context, event model.OrderStatusEvent) error {
	_, err := h.uc.Unassign(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("unassign courier: %w", err)
	}
	return nil
}

type completedHandler struct {
	uc deliveryUsecase
}

func (h *completedHandler) Handle(ctx context.Context, event model.OrderStatusEvent) error {
	_, err := h.uc.Complete(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("complete delivery: %w", err)
	}
	return nil
}
