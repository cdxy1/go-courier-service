package courier

import (
	"context"
	"errors"
	"testing"

	"github.com/cdxy1/go-courier-service/internal/model"
	repoerrors "github.com/cdxy1/go-courier-service/internal/repository/courier"
)

var errBoom = errors.New("failed")

type mockCourierRepository struct {
	t            *testing.T
	createFn     func(ctx context.Context, courier *model.CourierModel) (int, error)
	updateFn     func(ctx context.Context, courier *model.CourierModel) error
	getOneByIDFn func(ctx context.Context, id int) (*model.CourierModel, error)
	getAllFn     func(ctx context.Context) ([]*model.CourierModel, error)
}

func newMockCourierRepository(t *testing.T) *mockCourierRepository {
	return &mockCourierRepository{t: t}
}

func (m *mockCourierRepository) Create(ctx context.Context, courier *model.CourierModel) (int, error) {
	if m.createFn == nil {
		m.t.Fatalf("Create called unexpectedly")
	}
	return m.createFn(ctx, courier)
}

func (m *mockCourierRepository) Update(ctx context.Context, courier *model.CourierModel) error {
	if m.updateFn == nil {
		m.t.Fatalf("Update called unexpectedly")
	}
	return m.updateFn(ctx, courier)
}

func (m *mockCourierRepository) GetOneById(ctx context.Context, id int) (*model.CourierModel, error) {
	if m.getOneByIDFn == nil {
		m.t.Fatalf("GetOneById called unexpectedly")
	}
	return m.getOneByIDFn(ctx, id)
}

func (m *mockCourierRepository) GetAll(ctx context.Context) ([]*model.CourierModel, error) {
	if m.getAllFn == nil {
		m.t.Fatalf("GetAll called unexpectedly")
	}
	return m.getAllFn(ctx)
}

func TestCourierUsecase_GetOneById(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		id          int
		repoSetup   func(*mockCourierRepository)
		expectErr   error
		expectModel *model.CourierModel
	}{
		{
			name:      "invalid id",
			id:        0,
			repoSetup: func(_ *mockCourierRepository) {},
			expectErr: ErrInvalidID,
		},
		{
			name: "not found",
			id:   1,
			repoSetup: func(repo *mockCourierRepository) {
				repo.getOneByIDFn = func(ctx context.Context, id int) (*model.CourierModel, error) {
					if id != 1 {
						repo.t.Fatalf("unexpected id: %d", id)
					}
					return nil, repoerrors.ErrCourierNotFound
				}
			},
			expectErr: repoerrors.ErrCourierNotFound,
		},
		{
			name: "repo error",
			id:   2,
			repoSetup: func(repo *mockCourierRepository) {
				repo.getOneByIDFn = func(ctx context.Context, id int) (*model.CourierModel, error) {
					return nil, errBoom
				}
			},
			expectErr: errBoom,
		},
		{
			name: "success",
			id:   3,
			repoSetup: func(repo *mockCourierRepository) {
				repo.getOneByIDFn = func(ctx context.Context, id int) (*model.CourierModel, error) {
					return &model.CourierModel{ID: id, Name: "John"}, nil
				}
			},
			expectModel: &model.CourierModel{ID: 3, Name: "John"},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := newMockCourierRepository(t)
			tt.repoSetup(repo)
			uc := NewCourierUsecase(repo)

			result, err := uc.GetOneById(ctx, tt.id)

			switch {
			case tt.expectErr != nil:
				if err == nil || !errors.Is(err, tt.expectErr) {
					t.Fatalf("expected error %v, got %v", tt.expectErr, err)
				}
			case tt.expectModel != nil:
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result == nil || result.ID != tt.expectModel.ID || result.Name != tt.expectModel.Name {
					t.Fatalf("unexpected result: %+v", result)
				}
			default:
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCourierUsecase_Create(t *testing.T) {
	t.Parallel()

	validCourier := &model.CourierModel{
		Name:          "Alice",
		Phone:         "+71234567890",
		Status:        model.CourierStatusAvailable,
		TransportType: model.TransportCar,
	}

	tests := []struct {
		name      string
		input     *model.CourierModel
		setupRepo func(*mockCourierRepository)
		expectErr error
		expectID  int
	}{
		{
			name:      "invalid name",
			input:     &model.CourierModel{Phone: "+71234567890"},
			setupRepo: func(_ *mockCourierRepository) {},
			expectErr: ErrInvalidName,
		},
		{
			name:      "invalid phone",
			input:     &model.CourierModel{Name: "Alice", Phone: "123"},
			setupRepo: func(_ *mockCourierRepository) {},
			expectErr: ErrInvalidPhone,
		},
		{
			name:  "phone exists",
			input: validCourier,
			setupRepo: func(repo *mockCourierRepository) {
				repo.createFn = func(ctx context.Context, courier *model.CourierModel) (int, error) {
					return 0, repoerrors.ErrPhoneExists
				}
			},
			expectErr: repoerrors.ErrPhoneExists,
		},
		{
			name:  "repository error",
			input: validCourier,
			setupRepo: func(repo *mockCourierRepository) {
				repo.createFn = func(ctx context.Context, courier *model.CourierModel) (int, error) {
					return 0, errBoom
				}
			},
			expectErr: errBoom,
		},
		{
			name:  "success",
			input: validCourier,
			setupRepo: func(repo *mockCourierRepository) {
				repo.createFn = func(ctx context.Context, courier *model.CourierModel) (int, error) {
					if courier.Name != validCourier.Name {
						repo.t.Fatalf("unexpected name: %s", courier.Name)
					}
					return 42, nil
				}
			},
			expectID: 42,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := newMockCourierRepository(t)
			tt.setupRepo(repo)
			uc := NewCourierUsecase(repo)

			id, err := uc.Create(ctx, tt.input)

			if tt.expectErr != nil {
				if err == nil || !errors.Is(err, tt.expectErr) {
					t.Fatalf("expected error %v, got %v", tt.expectErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expectID {
				t.Fatalf("expected id %d, got %d", tt.expectID, id)
			}
		})
	}
}

func TestCourierUsecase_Update(t *testing.T) {
	t.Parallel()

	validCourier := &model.CourierModel{
		ID:            1,
		Name:          "Alice",
		Phone:         "+71234567890",
		Status:        model.CourierStatusAvailable,
		TransportType: model.TransportCar,
	}

	tests := []struct {
		name      string
		input     *model.CourierModel
		setupRepo func(*mockCourierRepository)
		expectErr error
	}{
		{
			name:      "invalid id",
			input:     &model.CourierModel{},
			setupRepo: func(_ *mockCourierRepository) {},
			expectErr: ErrInvalidID,
		},
		{
			name:      "invalid name",
			input:     &model.CourierModel{ID: 1, Phone: "+71234567890"},
			setupRepo: func(_ *mockCourierRepository) {},
			expectErr: ErrInvalidName,
		},
		{
			name:      "invalid phone",
			input:     &model.CourierModel{ID: 1, Name: "Alice", Phone: "123"},
			setupRepo: func(_ *mockCourierRepository) {},
			expectErr: ErrInvalidPhone,
		},
		{
			name:  "not found",
			input: validCourier,
			setupRepo: func(repo *mockCourierRepository) {
				repo.updateFn = func(ctx context.Context, courier *model.CourierModel) error {
					return repoerrors.ErrCourierNotFound
				}
			},
			expectErr: repoerrors.ErrCourierNotFound,
		},
		{
			name:  "phone exists",
			input: validCourier,
			setupRepo: func(repo *mockCourierRepository) {
				repo.updateFn = func(ctx context.Context, courier *model.CourierModel) error {
					return repoerrors.ErrPhoneExists
				}
			},
			expectErr: repoerrors.ErrPhoneExists,
		},
		{
			name:  "repository error",
			input: validCourier,
			setupRepo: func(repo *mockCourierRepository) {
				repo.updateFn = func(ctx context.Context, courier *model.CourierModel) error {
					return errBoom
				}
			},
			expectErr: errBoom,
		},
		{
			name:  "success",
			input: validCourier,
			setupRepo: func(repo *mockCourierRepository) {
				repo.updateFn = func(ctx context.Context, courier *model.CourierModel) error {
					if courier.ID != validCourier.ID {
						repo.t.Fatalf("unexpected id: %d", courier.ID)
					}
					return nil
				}
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := newMockCourierRepository(t)
			tt.setupRepo(repo)
			uc := NewCourierUsecase(repo)

			err := uc.Update(ctx, tt.input)

			if tt.expectErr != nil {
				if err == nil || !errors.Is(err, tt.expectErr) {
					t.Fatalf("expected error %v, got %v", tt.expectErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCourierUsecase_GetAll(t *testing.T) {
	t.Parallel()

	repo := newMockCourierRepository(t)
	repo.getAllFn = func(ctx context.Context) ([]*model.CourierModel, error) {
		return []*model.CourierModel{{ID: 1}, {ID: 2}}, nil
	}

	uc := NewCourierUsecase(repo)

	result, err := uc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 couriers, got %d", len(result))
	}
}
