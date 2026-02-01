package model

import "time"

type DeliveryModel struct {
	ID         int
	CourierId  int
	OrderId    string
	AssignedAt time.Time
	Deadline   time.Time
}
