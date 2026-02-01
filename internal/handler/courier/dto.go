package courier

import "github.com/cdxy1/go-courier-service/internal/model"

type createCourierRequest struct {
	Name          string              `json:"name"`
	Phone         string              `json:"phone"`
	Status        model.CourierStatus `json:"status"`
	TransportType model.TransportType `json:"transport_type"`
}

type updateCourierRequest struct {
	ID            int                 `json:"id"`
	Name          string              `json:"name"`
	Phone         string              `json:"phone"`
	Status        model.CourierStatus `json:"status"`
	TransportType model.TransportType `json:"transport_type"`
}

type courierResponse struct {
	ID            int                 `json:"id"`
	Name          string              `json:"name"`
	Phone         string              `json:"phone"`
	Status        model.CourierStatus `json:"status"`
	TransportType model.TransportType `json:"transport_type"`
}
