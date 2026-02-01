package delivery

import (
	"time"

	"github.com/cdxy1/go-courier-service/internal/model"
)

type assignRequest struct {
	OrderId string `json:"order_id"`
}

type assignResponse struct {
	CourierId        int                 `json:"courier_id"`
	OrderID          string              `json:"order_id"`
	TransportType    model.TransportType `json:"transport_type"`
	DeliveryDeadline time.Time           `json:"delivery_deadline"`
}

type unassignRequest struct {
	OrderId string `json:"order_id"`
}

type unassignResponse struct {
	OrderId   string `json:"order_id"`
	Status    string `json:"status"`
	CourierId int    `json:"courier_id"`
}
