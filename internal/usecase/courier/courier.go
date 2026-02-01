package courier

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/cdxy1/go-courier-service/internal/model"
	repo "github.com/cdxy1/go-courier-service/internal/repository/courier"
)

type CourierUsecase struct {
	repo courierRepository
}

func NewCourierUsecase(repo courierRepository) *CourierUsecase {
	return &CourierUsecase{repo: repo}
}

func validatePhone(phone string) bool {
	re := regexp.MustCompile(`^\+?\d{10,15}$`)
	return re.MatchString(phone)
}

func (uc *CourierUsecase) GetOneById(ctx context.Context, id int) (*model.CourierModel, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	courier, err := uc.repo.GetOneById(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrCourierNotFound) {
			return nil, repo.ErrCourierNotFound
		}
		return nil, fmt.Errorf("get courier: %w", err)
	}
	return courier, nil
}

func (uc *CourierUsecase) GetAll(ctx context.Context) ([]*model.CourierModel, error) {
	return uc.repo.GetAll(ctx)
}

func (uc *CourierUsecase) Create(ctx context.Context, req *model.CourierModel) (int, error) {
	if req.Name == "" {
		return 0, ErrInvalidName
	}
	if !validatePhone(req.Phone) {
		return 0, ErrInvalidPhone
	}
	id, err := uc.repo.Create(ctx, req)
	if err != nil {
		if errors.Is(err, repo.ErrPhoneExists) {
			return 0, repo.ErrPhoneExists
		}
		return 0, fmt.Errorf("create courier: %w", err)
	}
	return id, nil
}

func (uc *CourierUsecase) Update(ctx context.Context, req *model.CourierModel) error {
	if req.ID <= 0 {
		return ErrInvalidID
	}
	if req.Name == "" {
		return ErrInvalidName
	}
	if !validatePhone(req.Phone) {
		return ErrInvalidPhone
	}
	err := uc.repo.Update(ctx, req)
	if err != nil {
		if errors.Is(err, repo.ErrCourierNotFound) {
			return repo.ErrCourierNotFound
		}
		if errors.Is(err, repo.ErrPhoneExists) {
			return repo.ErrPhoneExists
		}
		return fmt.Errorf("update courier: %w", err)
	}
	return nil
}

func (uc *CourierUsecase) AssignCourierToOrder(ctx context.Context, orderID string) (int, error) {
	couriers, err := uc.repo.GetAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("list couriers: %w", err)
	}

	if len(couriers) == 0 {
		return 0, repo.ErrCourierNotFound
	}

	var selectedCourier *model.CourierModel
	for _, courier := range couriers {
		if courier.Status == model.CourierStatusActive || courier.Status == "" {
			selectedCourier = courier
			break
		}
	}

	if selectedCourier == nil {
		selectedCourier = couriers[0]
	}

	return selectedCourier.ID, nil
}
