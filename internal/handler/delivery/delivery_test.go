package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	handlerErrors "github.com/cdxy1/go-courier-service/internal/handler/errors"
	"github.com/cdxy1/go-courier-service/internal/model"
	"github.com/labstack/echo/v4"
)

var errBoom = errors.New("failed")

type mockDeliveryUsecase struct {
	t          *testing.T
	assignFn   func(ctx context.Context, orderID string) (*model.DeliveryModel, *model.CourierModel, error)
	unassignFn func(ctx context.Context, orderID string) (*model.DeliveryModel, error)
}

func newMockDeliveryUsecase(t *testing.T) *mockDeliveryUsecase {
	return &mockDeliveryUsecase{t: t}
}

func (m *mockDeliveryUsecase) Assign(ctx context.Context, orderID string) (*model.DeliveryModel, *model.CourierModel, error) {
	if m.assignFn == nil {
		m.t.Fatalf("Assign called unexpectedly")
	}
	return m.assignFn(ctx, orderID)
}

func (m *mockDeliveryUsecase) Unassign(ctx context.Context, orderID string) (*model.DeliveryModel, error) {
	if m.unassignFn == nil {
		m.t.Fatalf("Unassign called unexpectedly")
	}
	return m.unassignFn(ctx, orderID)
}

func TestDeliveryHandler_Assign(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		setup      func(*mockDeliveryUsecase)
		wantStatus int
		wantErr    string
		wantResp   *assignResponse
	}{
		{
			name:       "invalid body",
			body:       "{",
			setup:      func(_ *mockDeliveryUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    handlerErrors.ErrBadRequest.Error(),
		},
		{
			name:       "empty order id",
			body:       `{"order_id":""}`,
			setup:      func(_ *mockDeliveryUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    handlerErrors.ErrBadRequest.Error(),
		},
		{
			name: "usecase error",
			body: `{"order_id":"order-1"}`,
			setup: func(uc *mockDeliveryUsecase) {
				uc.assignFn = func(ctx context.Context, orderID string) (*model.DeliveryModel, *model.CourierModel, error) {
					if orderID != "order-1" {
						uc.t.Fatalf("unexpected orderID: %s", orderID)
					}
					return nil, nil, errBoom
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "internal server error",
		},
		{
			name: "success",
			body: `{"order_id":"order-1"}`,
			setup: func(uc *mockDeliveryUsecase) {
				uc.assignFn = func(ctx context.Context, orderID string) (*model.DeliveryModel, *model.CourierModel, error) {
					return &model.DeliveryModel{OrderId: orderID, CourierId: 11}, &model.CourierModel{ID: 11, TransportType: model.TransportCar}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantResp:   &assignResponse{CourierId: 11, OrderID: "order-1", TransportType: model.TransportCar},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/delivery/assign", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			uc := newMockDeliveryUsecase(t)
			tt.setup(uc)
			handler := NewDeliveryHandler(uc)

			if err := handler.Assign(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if tt.wantResp != nil {
				var resp assignResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.CourierId != tt.wantResp.CourierId || resp.OrderID != tt.wantResp.OrderID || resp.TransportType != tt.wantResp.TransportType {
					t.Fatalf("unexpected response: %+v", resp)
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

func TestDeliveryHandler_Unassign(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		setup      func(*mockDeliveryUsecase)
		wantStatus int
		wantErr    string
		wantResp   *unassignResponse
	}{
		{
			name:       "invalid body",
			body:       "{",
			setup:      func(_ *mockDeliveryUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    handlerErrors.ErrBadRequest.Error(),
		},
		{
			name:       "empty order id",
			body:       `{"order_id":""}`,
			setup:      func(_ *mockDeliveryUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    handlerErrors.ErrBadRequest.Error(),
		},
		{
			name: "usecase error",
			body: `{"order_id":"order-2"}`,
			setup: func(uc *mockDeliveryUsecase) {
				uc.unassignFn = func(ctx context.Context, orderID string) (*model.DeliveryModel, error) {
					return nil, errBoom
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "internal server error",
		},
		{
			name: "success",
			body: `{"order_id":"order-2"}`,
			setup: func(uc *mockDeliveryUsecase) {
				uc.unassignFn = func(ctx context.Context, orderID string) (*model.DeliveryModel, error) {
					return &model.DeliveryModel{OrderId: orderID, CourierId: 7}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantResp:   &unassignResponse{OrderId: "order-2", Status: "unassigned", CourierId: 7},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/delivery/unassign", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			uc := newMockDeliveryUsecase(t)
			tt.setup(uc)
			handler := NewDeliveryHandler(uc)

			if err := handler.Unassign(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if tt.wantResp != nil {
				var resp unassignResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp != *tt.wantResp {
					t.Fatalf("unexpected response: %+v", resp)
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
