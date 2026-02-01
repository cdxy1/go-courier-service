package model

import "time"

type CourierModel struct {
	ID               int
	Name             string
	Phone            string
	Status           CourierStatus
	TransportType    TransportType
	AssignmentsCount int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
