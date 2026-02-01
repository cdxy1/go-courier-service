package courier

import (
	"context"

	"github.com/cdxy1/go-courier-service/internal/model"
)

type courierRepository interface {
	Create(ctx context.Context, courier *model.CourierModel) (int, error)
	Update(ctx context.Context, courier *model.CourierModel) error
	GetOneById(ctx context.Context, id int) (*model.CourierModel, error)
	GetAll(ctx context.Context) ([]*model.CourierModel, error)
}
