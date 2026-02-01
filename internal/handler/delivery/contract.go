package delivery

import (
	"context"

	"github.com/cdxy1/go-courier-service/internal/model"
)

type deliveryUsecase interface {
	Assign(ctx context.Context, orderID string) (*model.DeliveryModel, *model.CourierModel, error)
	Unassign(ctx context.Context, orderID string) (*model.DeliveryModel, error)
}
