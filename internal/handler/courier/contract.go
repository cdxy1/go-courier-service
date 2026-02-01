package courier

import (
	"context"

	"github.com/cdxy1/go-courier-service/internal/model"
)

type courierUsecase interface {
	GetOneById(ctx context.Context, id int) (*model.CourierModel, error)
	GetAll(ctx context.Context) ([]*model.CourierModel, error)
	Create(ctx context.Context, req *model.CourierModel) (int, error)
	Update(ctx context.Context, req *model.CourierModel) error
}
