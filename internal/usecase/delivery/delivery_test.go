package delivery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cdxy1/go-courier-service/internal/model"
)

var errBoom = errors.New("failed")

type mockCourierRepository struct {
	t              *testing.T
	getAvailableFn func(ctx context.Context) (*model.CourierModel, error)
	updateStatusFn func(ctx context.Context, status model.CourierStatus, id int) error
	markAssignedFn func(ctx context.Context, id int) error
}

func newMockCourierRepository(t *testing.T) *mockCourierRepository {
	return &mockCourierRepository{t: t}
}

func (m *mockCourierRepository) Create(ctx context.Context, courier *model.CourierModel) (int, error) {
	m.t.Fatalf("Create called unexpectedly")
	return 0, nil
}

func (m *mockCourierRepository) Update(ctx context.Context, courier *model.CourierModel) error {
	m.t.Fatalf("Update called unexpectedly")
	return nil
}

func (m *mockCourierRepository) GetOneById(ctx context.Context, id int) (*model.CourierModel, error) {
	m.t.Fatalf("GetOneById called unexpectedly")
	return nil, nil
}

func (m *mockCourierRepository) GetAll(ctx context.Context) ([]*model.CourierModel, error) {
	m.t.Fatalf("GetAll called unexpectedly")
	return nil, nil
}

func (m *mockCourierRepository) GetByStatus(ctx context.Context, status model.CourierStatus) (*model.CourierModel, error) {
	m.t.Fatalf("GetByStatus called unexpectedly")
	return nil, nil
}

func (m *mockCourierRepository) UpdateStatus(ctx context.Context, status model.CourierStatus, id int) error {
	if m.updateStatusFn == nil {
		m.t.Fatalf("UpdateStatus called unexpectedly")
	}
	return m.updateStatusFn(ctx, status, id)
}

func (m *mockCourierRepository) MarkAssigned(ctx context.Context, id int) error {
	if m.markAssignedFn == nil {
		m.t.Fatalf("MarkAssigned called unexpectedly")
	}
	return m.markAssignedFn(ctx, id)
}

func (m *mockCourierRepository) GetAvailableLeastDelivered(ctx context.Context) (*model.CourierModel, error) {
	if m.getAvailableFn == nil {
		m.t.Fatalf("GetAvailableLeastDelivered called unexpectedly")
	}
	return m.getAvailableFn(ctx)
}

type mockDeliveryRepository struct {
	t                *testing.T
	createFn         func(ctx context.Context, delivery *model.DeliveryModel) error
	deleteFn         func(ctx context.Context, orderId string) (int, error)
	getCourierIDFn   func(ctx context.Context, orderId string) (int, error)
	releaseExpiredFn func(ctx context.Context) (int, error)
}

func newMockDeliveryRepository(t *testing.T) *mockDeliveryRepository {
	return &mockDeliveryRepository{t: t}
}

func (m *mockDeliveryRepository) Create(ctx context.Context, delivery *model.DeliveryModel) error {
	if m.createFn == nil {
		m.t.Fatalf("Create called unexpectedly")
	}
	return m.createFn(ctx, delivery)
}

func (m *mockDeliveryRepository) Delete(ctx context.Context, orderId string) (int, error) {
	if m.deleteFn == nil {
		m.t.Fatalf("Delete called unexpectedly")
	}
	return m.deleteFn(ctx, orderId)
}

func (m *mockDeliveryRepository) GetCourierID(ctx context.Context, orderId string) (int, error) {
	if m.getCourierIDFn == nil {
		m.t.Fatalf("GetCourierID called unexpectedly")
	}
	return m.getCourierIDFn(ctx, orderId)
}

func (m *mockDeliveryRepository) ReleaseExpiredCouriers(ctx context.Context) (int, error) {
	if m.releaseExpiredFn == nil {
		m.t.Fatalf("ReleaseExpiredCouriers called unexpectedly")
	}
	return m.releaseExpiredFn(ctx)
}

type mockTxManager struct {
	t        *testing.T
	withTxFn func(ctx context.Context, fn func(context.Context) error) error
}

func newMockTxManager(t *testing.T) *mockTxManager {
	return &mockTxManager{t: t}
}

func (m *mockTxManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	return fn(ctx)
}

func TestDeliveryUsecase_Assign(t *testing.T) {
	t.Parallel()

	orderID := "order-1"
	now := time.Date(2025, time.December, 22, 10, 0, 0, 0, time.UTC)
	factory := model.NewDeliveryTimeFactory(time.Minute*30, time.Minute*15, time.Minute*5)

	tests := []struct {
		name      string
		setup     func(*mockCourierRepository, *mockDeliveryRepository, *mockTxManager)
		expectErr error
	}{
		{
			name: "success",
			setup: func(cRepo *mockCourierRepository, dRepo *mockDeliveryRepository, tm *mockTxManager) {
				courier := &model.CourierModel{ID: 7, TransportType: model.TransportCar}
				cRepo.getAvailableFn = func(ctx context.Context) (*model.CourierModel, error) {
					return courier, nil
				}
				dRepo.createFn = func(ctx context.Context, delivery *model.DeliveryModel) error {
					if delivery.CourierId != courier.ID {
						t.Fatalf("unexpected courier id: %d", delivery.CourierId)
					}
					if delivery.OrderId != orderID {
						t.Fatalf("unexpected order id: %s", delivery.OrderId)
					}
					if delivery.Deadline.IsZero() {
						t.Fatalf("deadline must be set")
					}
					return nil
				}
				cRepo.markAssignedFn = func(ctx context.Context, id int) error {
					if id != courier.ID {
						t.Fatalf("unexpected id: %d", id)
					}
					return nil
				}
			},
		},
		{
			name: "get courier error",
			setup: func(cRepo *mockCourierRepository, dRepo *mockDeliveryRepository, tm *mockTxManager) {
				cRepo.getAvailableFn = func(ctx context.Context) (*model.CourierModel, error) {
					return nil, errBoom
				}
			},
			expectErr: errBoom,
		},
		{
			name: "create delivery error",
			setup: func(cRepo *mockCourierRepository, dRepo *mockDeliveryRepository, tm *mockTxManager) {
				cRepo.getAvailableFn = func(ctx context.Context) (*model.CourierModel, error) {
					return &model.CourierModel{ID: 1}, nil
				}
				dRepo.createFn = func(ctx context.Context, delivery *model.DeliveryModel) error {
					return errBoom
				}
			},
			expectErr: errBoom,
		},
		{
			name: "mark assigned error",
			setup: func(cRepo *mockCourierRepository, dRepo *mockDeliveryRepository, tm *mockTxManager) {
				cRepo.getAvailableFn = func(ctx context.Context) (*model.CourierModel, error) {
					return &model.CourierModel{ID: 1}, nil
				}
				dRepo.createFn = func(ctx context.Context, delivery *model.DeliveryModel) error {
					return nil
				}
				cRepo.markAssignedFn = func(ctx context.Context, id int) error {
					return errBoom
				}
			},
			expectErr: errBoom,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cRepo := newMockCourierRepository(t)
			dRepo := newMockDeliveryRepository(t)
			tm := newMockTxManager(t)

			tt.setup(cRepo, dRepo, tm)

			uc := NewDeliveryUsecase(cRepo, dRepo, tm, factory, func() time.Time { return now })
			delivery, courier, err := uc.Assign(context.Background(), orderID)

			if tt.expectErr != nil {
				if err == nil || !errors.Is(err, tt.expectErr) {
					t.Fatalf("expected error %v, got %v", tt.expectErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if delivery == nil || courier == nil {
				t.Fatalf("expected non-nil delivery and courier")
			}
			if delivery.OrderId != orderID {
				t.Fatalf("unexpected order id: %s", delivery.OrderId)
			}
			if delivery.CourierId != courier.ID {
				t.Fatalf("courier id mismatch")
			}
			if want := now.Add(time.Minute * 5); !delivery.Deadline.Equal(want) {
				t.Fatalf("unexpected deadline: %s", delivery.Deadline)
			}
		})
	}
}

func TestDeliveryUsecase_Unassign(t *testing.T) {
	t.Parallel()

	orderID := "order-77"

	tests := []struct {
		name      string
		setup     func(*mockCourierRepository, *mockDeliveryRepository, *mockTxManager)
		expectErr error
	}{
		{
			name: "success",
			setup: func(cRepo *mockCourierRepository, dRepo *mockDeliveryRepository, tm *mockTxManager) {
				dRepo.deleteFn = func(ctx context.Context, orderId string) (int, error) {
					if orderId != orderID {
						t.Fatalf("unexpected order id: %s", orderId)
					}
					return 5, nil
				}
				cRepo.updateStatusFn = func(ctx context.Context, status model.CourierStatus, id int) error {
					if status != model.CourierStatusAvailable {
						t.Fatalf("unexpected status: %s", status)
					}
					if id != 5 {
						t.Fatalf("unexpected id: %d", id)
					}
					return nil
				}
			},
		},
		{
			name: "delete error",
			setup: func(cRepo *mockCourierRepository, dRepo *mockDeliveryRepository, tm *mockTxManager) {
				dRepo.deleteFn = func(ctx context.Context, orderId string) (int, error) {
					return 0, errBoom
				}
			},
			expectErr: errBoom,
		},
		{
			name: "update status error",
			setup: func(cRepo *mockCourierRepository, dRepo *mockDeliveryRepository, tm *mockTxManager) {
				dRepo.deleteFn = func(ctx context.Context, orderId string) (int, error) {
					return 3, nil
				}
				cRepo.updateStatusFn = func(ctx context.Context, status model.CourierStatus, id int) error {
					return errBoom
				}
			},
			expectErr: errBoom,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cRepo := newMockCourierRepository(t)
			dRepo := newMockDeliveryRepository(t)
			tm := newMockTxManager(t)

			tt.setup(cRepo, dRepo, tm)

			uc := NewDeliveryUsecase(cRepo, dRepo, tm, model.NewDeliveryTimeFactory(time.Minute, time.Minute, time.Minute), time.Now)
			result, err := uc.Unassign(context.Background(), orderID)

			if tt.expectErr != nil {
				if err == nil || !errors.Is(err, tt.expectErr) {
					t.Fatalf("expected error %v, got %v", tt.expectErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatalf("result must not be nil")
			}
			if result.OrderId != orderID {
				t.Fatalf("unexpected order id: %s", result.OrderId)
			}
			if result.CourierId != 5 {
				t.Fatalf("unexpected courier id: %d", result.CourierId)
			}
		})
	}
}

func TestDeliveryUsecase_ProcessExpiredDeliveries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		releaseFn   func(ctx context.Context) (int, error)
		expectCount int
		expectErr   error
	}{
		{
			name: "success",
			releaseFn: func(ctx context.Context) (int, error) {
				return 3, nil
			},
			expectCount: 3,
		},
		{
			name: "release error",
			releaseFn: func(ctx context.Context) (int, error) {
				return 0, errBoom
			},
			expectErr: errBoom,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cRepo := newMockCourierRepository(t)
			dRepo := newMockDeliveryRepository(t)
			m := newMockTxManager(t)

			dRepo.releaseExpiredFn = tt.releaseFn

			uc := NewDeliveryUsecase(cRepo, dRepo, m, model.NewDeliveryTimeFactory(time.Minute, time.Minute, time.Minute), time.Now)
			count, err := uc.ProcessExpiredDeliveries(context.Background())

			if tt.expectErr != nil {
				if err == nil || !errors.Is(err, tt.expectErr) {
					t.Fatalf("expected error %v, got %v", tt.expectErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tt.expectCount {
				t.Fatalf("expected count %d, got %d", tt.expectCount, count)
			}
		})
	}
}
