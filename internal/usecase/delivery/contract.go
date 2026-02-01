package delivery

import (
	"context"

	"github.com/cdxy1/go-courier-service/internal/model"
)

type courierRepository interface {
	Create(ctx context.Context, courier *model.CourierModel) (int, error)
	Update(ctx context.Context, courier *model.CourierModel) error
	GetOneById(ctx context.Context, id int) (*model.CourierModel, error)
	GetAll(ctx context.Context) ([]*model.CourierModel, error)
	GetByStatus(ctx context.Context, status model.CourierStatus) (*model.CourierModel, error)
	UpdateStatus(ctx context.Context, status model.CourierStatus, id int) error
	MarkAssigned(ctx context.Context, id int) error
	GetAvailableLeastDelivered(ctx context.Context) (*model.CourierModel, error)
}

type deliveryRepository interface {
	Create(ctx context.Context, delivery *model.DeliveryModel) error
	Delete(ctx context.Context, orderId string) (int, error)
	GetCourierID(ctx context.Context, orderId string) (int, error)
	ReleaseExpiredCouriers(ctx context.Context) (int, error)
}

type txManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
