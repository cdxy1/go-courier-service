package model

type CourierStatus string

const (
	CourierStatusAvailable CourierStatus = "available"
	CourierStatusBusy      CourierStatus = "busy"
	CourierStatusPaused    CourierStatus = "paused"
	CourierStatusActive    CourierStatus = "active"
)

type TransportType string

const (
	TransportOnFoot  TransportType = "on_foot"
	TransportScooter TransportType = "scooter"
	TransportCar     TransportType = "car"
)
