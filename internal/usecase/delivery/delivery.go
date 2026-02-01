package delivery

import (
	"context"
	"fmt"

	"github.com/cdxy1/go-courier-service/internal/model"
)

type DeliveryUsecase struct {
	courierRepo  courierRepository
	deliveryRepo deliveryRepository
	tm           txManager
	timeFactory  *model.DeliveryTimeFactory
	now          model.NowFunc
}

func NewDeliveryUsecase(
	courierRepo courierRepository,
	deliveryRepo deliveryRepository,
	tm txManager,
	timeFactory *model.DeliveryTimeFactory,
	now model.NowFunc,
) *DeliveryUsecase {
	return &DeliveryUsecase{
		courierRepo:  courierRepo,
		deliveryRepo: deliveryRepo,
		tm:           tm,
		timeFactory:  timeFactory,
		now:          now,
	}
}

func (uc *DeliveryUsecase) Assign(ctx context.Context, order_id string) (*model.DeliveryModel, *model.CourierModel, error) {
	var createdDelivery *model.DeliveryModel
	var assignedCourier *model.CourierModel

	if err := uc.tm.WithTx(ctx, func(ctx context.Context) error {
		courier, err := uc.courierRepo.GetAvailableLeastDelivered(ctx)
		if err != nil {
			return fmt.Errorf("get available courier: %w", err)
		}

		deadline := uc.timeFactory.ForTransport(courier.TransportType).Deadline(uc.now())
		d := &model.DeliveryModel{
			CourierId: courier.ID,
			OrderId:   order_id,
			Deadline:  deadline,
		}

		if err := uc.deliveryRepo.Create(ctx, d); err != nil {
			return fmt.Errorf("create delivery: %w", err)
		}

		if err := uc.courierRepo.MarkAssigned(ctx, courier.ID); err != nil {
			return fmt.Errorf("mark courier assigned: %w", err)
		}

		createdDelivery = d
		assignedCourier = courier
		return nil
	}); err != nil {
		return nil, nil, err
	}

	return createdDelivery, assignedCourier, nil
}

func (uc *DeliveryUsecase) Unassign(ctx context.Context, orderId string) (*model.DeliveryModel, error) {
	var courierId int
	if err := uc.tm.WithTx(ctx, func(ctx context.Context) error {
		cid, err := uc.deliveryRepo.Delete(ctx, orderId)
		if err != nil {
			return fmt.Errorf("delete delivery: %w", err)
		}
		courierId = cid

		if err := uc.courierRepo.UpdateStatus(ctx, model.CourierStatusAvailable, courierId); err != nil {
			return fmt.Errorf("update courier status: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &model.DeliveryModel{OrderId: orderId, CourierId: courierId}, nil
}

func (uc *DeliveryUsecase) Complete(ctx context.Context, orderId string) (*model.DeliveryModel, error) {
	var courierId int
	if err := uc.tm.WithTx(ctx, func(ctx context.Context) error {
		cid, err := uc.deliveryRepo.GetCourierID(ctx, orderId)
		if err != nil {
			return fmt.Errorf("get delivery courier: %w", err)
		}
		courierId = cid

		if err := uc.courierRepo.UpdateStatus(ctx, model.CourierStatusAvailable, courierId); err != nil {
			return fmt.Errorf("update courier status: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &model.DeliveryModel{OrderId: orderId, CourierId: courierId}, nil
}

func (uc *DeliveryUsecase) ProcessExpiredDeliveries(ctx context.Context) (int, error) {
	var updated int
	err := uc.tm.WithTx(ctx, func(ctx context.Context) error {
		count, err := uc.deliveryRepo.ReleaseExpiredCouriers(ctx)
		if err != nil {
			return fmt.Errorf("release expired couriers: %w", err)
		}
		updated = count
		return nil
	})
	return updated, err
}
