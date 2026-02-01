package courier

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cdxy1/go-courier-service/internal/model"
	repo "github.com/cdxy1/go-courier-service/internal/repository/courier"
	usecase "github.com/cdxy1/go-courier-service/internal/usecase/courier"
	"github.com/labstack/echo/v4"
)

type mockCourierUsecase struct {
	t            *testing.T
	getOneByIDFn func(ctx context.Context, id int) (*model.CourierModel, error)
	getAllFn     func(ctx context.Context) ([]*model.CourierModel, error)
	createFn     func(ctx context.Context, req *model.CourierModel) (int, error)
	updateFn     func(ctx context.Context, req *model.CourierModel) error
}

func newMockCourierUsecase(t *testing.T) *mockCourierUsecase {
	return &mockCourierUsecase{t: t}
}

func (m *mockCourierUsecase) GetOneById(ctx context.Context, id int) (*model.CourierModel, error) {
	if m.getOneByIDFn == nil {
		m.t.Fatalf("GetOneById called unexpectedly")
	}
	return m.getOneByIDFn(ctx, id)
}

func (m *mockCourierUsecase) GetAll(ctx context.Context) ([]*model.CourierModel, error) {
	if m.getAllFn == nil {
		m.t.Fatalf("GetAll called unexpectedly")
	}
	return m.getAllFn(ctx)
}

func (m *mockCourierUsecase) Create(ctx context.Context, req *model.CourierModel) (int, error) {
	if m.createFn == nil {
		m.t.Fatalf("Create called unexpectedly")
	}
	return m.createFn(ctx, req)
}

func (m *mockCourierUsecase) Update(ctx context.Context, req *model.CourierModel) error {
	if m.updateFn == nil {
		m.t.Fatalf("Update called unexpectedly")
	}
	return m.updateFn(ctx, req)
}

func TestCourierHandler_GetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		param      string
		setup      func(*mockCourierUsecase)
		wantStatus int
		wantErr    string
		wantResp   *courierResponse
	}{
		{
			name:       "invalid path param",
			param:      "abc",
			setup:      func(_ *mockCourierUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid id",
		},
		{
			name:  "invalid id from usecase",
			param: "0",
			setup: func(m *mockCourierUsecase) {
				m.getOneByIDFn = func(ctx context.Context, id int) (*model.CourierModel, error) {
					if id != 0 {
						m.t.Fatalf("unexpected id: %d", id)
					}
					return nil, usecase.ErrInvalidID
				}
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    usecase.ErrInvalidID.Error(),
		},
		{
			name:  "not found",
			param: "1",
			setup: func(m *mockCourierUsecase) {
				m.getOneByIDFn = func(ctx context.Context, id int) (*model.CourierModel, error) {
					return nil, repo.ErrCourierNotFound
				}
			},
			wantStatus: http.StatusNotFound,
			wantErr:    repo.ErrCourierNotFound.Error(),
		},
		{
			name:  "internal error",
			param: "2",
			setup: func(m *mockCourierUsecase) {
				m.getOneByIDFn = func(ctx context.Context, id int) (*model.CourierModel, error) {
					return nil, errors.New("boom")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "internal server error",
		},
		{
			name:  "success",
			param: "3",
			setup: func(m *mockCourierUsecase) {
				m.getOneByIDFn = func(ctx context.Context, id int) (*model.CourierModel, error) {
					return &model.CourierModel{ID: id, Name: "Alice", Phone: "+79991234567", Status: model.CourierStatusAvailable, TransportType: model.TransportCar}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantResp:   &courierResponse{ID: 3, Name: "Alice", Phone: "+79991234567", Status: model.CourierStatusAvailable, TransportType: model.TransportCar},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/couriers/"+tc.param, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.param)

			uc := newMockCourierUsecase(t)
			tc.setup(uc)
			handler := NewCourierHandler(uc)

			if err := handler.GetByID(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantResp != nil {
				var resp courierResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp != *tc.wantResp {
					t.Fatalf("unexpected response: %+v", resp)
				}
			} else {
				var resp map[string]string
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp["error"] != tc.wantErr {
					t.Fatalf("expected error %q, got %q", tc.wantErr, resp["error"])
				}
			}
		})
	}
}

func TestCourierHandler_GetAll(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/couriers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	uc := newMockCourierUsecase(t)
	uc.getAllFn = func(ctx context.Context) ([]*model.CourierModel, error) {
		return []*model.CourierModel{{ID: 1}, {ID: 2}}, nil
	}

	handler := NewCourierHandler(uc)
	if err := handler.GetAll(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp []courierResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("expected 2 couriers, got %d", len(resp))
	}
}

func TestCourierHandler_GetAll_Error(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/couriers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	uc := newMockCourierUsecase(t)
	uc.getAllFn = func(ctx context.Context) ([]*model.CourierModel, error) {
		return nil, errors.New("boom")
	}

	handler := NewCourierHandler(uc)
	if err := handler.GetAll(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}

func TestCourierHandler_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		setup      func(*mockCourierUsecase)
		wantStatus int
		wantErr    string
		wantID     int
	}{
		{
			name:       "invalid body",
			body:       "{",
			setup:      func(_ *mockCourierUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid request body",
		},
		{
			name: "invalid data",
			body: `{"name":"","phone":"+79991234567","status":"available","transport_type":"car"}`,
			setup: func(m *mockCourierUsecase) {
				m.createFn = func(ctx context.Context, req *model.CourierModel) (int, error) {
					return 0, usecase.ErrInvalidName
				}
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    usecase.ErrInvalidName.Error(),
		},
		{
			name: "phone exists",
			body: `{"name":"Alice","phone":"+79991234567","status":"available","transport_type":"car"}`,
			setup: func(m *mockCourierUsecase) {
				m.createFn = func(ctx context.Context, req *model.CourierModel) (int, error) {
					return 0, repo.ErrPhoneExists
				}
			},
			wantStatus: http.StatusConflict,
			wantErr:    repo.ErrPhoneExists.Error(),
		},
		{
			name: "internal error",
			body: `{"name":"Alice","phone":"+79991234567","status":"available","transport_type":"car"}`,
			setup: func(m *mockCourierUsecase) {
				m.createFn = func(ctx context.Context, req *model.CourierModel) (int, error) {
					return 0, errors.New("boom")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "internal server error",
		},
		{
			name: "success",
			body: `{"name":"Alice","phone":"+79991234567","status":"available","transport_type":"car"}`,
			setup: func(m *mockCourierUsecase) {
				m.createFn = func(ctx context.Context, req *model.CourierModel) (int, error) {
					return 7, nil
				}
			},
			wantStatus: http.StatusCreated,
			wantID:     7,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/couriers", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			uc := newMockCourierUsecase(t)
			tt.setup(uc)

			handler := NewCourierHandler(uc)
			if err := handler.Create(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if tt.wantID != 0 {
				var resp map[string]int
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp["id"] != tt.wantID {
					t.Fatalf("expected id %d, got %d", tt.wantID, resp["id"])
				}
			} else {
				var resp map[string]string
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp["error"] != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, resp["error"])
				}
			}
		})
	}
}

func TestCourierHandler_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		setup      func(*mockCourierUsecase)
		wantStatus int
		wantErr    string
	}{
		{
			name:       "invalid body",
			body:       "{",
			setup:      func(_ *mockCourierUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid request body",
		},
		{
			name: "invalid data",
			body: `{"id":1,"name":"","phone":"+79991234567","status":"available","transport_type":"car"}`,
			setup: func(m *mockCourierUsecase) {
				m.updateFn = func(ctx context.Context, req *model.CourierModel) error {
					return usecase.ErrInvalidName
				}
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    usecase.ErrInvalidName.Error(),
		},
		{
			name: "not found",
			body: `{"id":1,"name":"Alice","phone":"+79991234567","status":"available","transport_type":"car"}`,
			setup: func(m *mockCourierUsecase) {
				m.updateFn = func(ctx context.Context, req *model.CourierModel) error {
					return repo.ErrCourierNotFound
				}
			},
			wantStatus: http.StatusNotFound,
			wantErr:    repo.ErrCourierNotFound.Error(),
		},
		{
			name: "phone exists",
			body: `{"id":1,"name":"Alice","phone":"+79991234567","status":"available","transport_type":"car"}`,
			setup: func(m *mockCourierUsecase) {
				m.updateFn = func(ctx context.Context, req *model.CourierModel) error {
					return repo.ErrPhoneExists
				}
			},
			wantStatus: http.StatusConflict,
			wantErr:    repo.ErrPhoneExists.Error(),
		},
		{
			name: "internal error",
			body: `{"id":1,"name":"Alice","phone":"+79991234567","status":"available","transport_type":"car"}`,
			setup: func(m *mockCourierUsecase) {
				m.updateFn = func(ctx context.Context, req *model.CourierModel) error {
					return errors.New("boom")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "internal server error",
		},
		{
			name: "success",
			body: `{"id":1,"name":"Alice","phone":"+79991234567","status":"available","transport_type":"car"}`,
			setup: func(m *mockCourierUsecase) {
				m.updateFn = func(ctx context.Context, req *model.CourierModel) error {
					if req.ID != 1 {
						m.t.Fatalf("unexpected id: %d", req.ID)
					}
					return nil
				}
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/couriers", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			uc := newMockCourierUsecase(t)
			tt.setup(uc)
			handler := NewCourierHandler(uc)

			if err := handler.Update(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if tt.wantErr != "" {
				var resp map[string]string
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp["error"] != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, resp["error"])
				}
			} else {
				var resp map[string]string
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp["status"] != "updated" {
					t.Fatalf("expected status 'updated', got %q", resp["status"])
				}
			}
		})
	}
}
