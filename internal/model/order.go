package model

import "time"

type Order struct {
	ID                string          `json:"id"`
	UserID            string          `json:"user_id"`
	OrderNumber       string          `json:"order_number,omitempty"`
	FIO               string          `json:"fio,omitempty"`
	RestaurantID      string          `json:"restaurant_id"`
	Items             []Item          `json:"items"`
	TotalPrice        int64           `json:"total_price"`
	Address           DeliveryAddress `json:"address"`
	Status            string          `json:"status"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	EstimatedDelivery time.Time       `json:"estimated_delivery"`
}

type Item struct {
	FoodID   string `json:"food_id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    int    `json:"price"`
}

type DeliveryAddress struct {
	Street    string `json:"street,omitempty"`
	House     string `json:"house,omitempty"`
	Apartment string `json:"apartment,omitempty"`
	Floor     string `json:"floor,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

type AssignCourierRequest struct {
	OrderID   string `json:"order_id"`
	CourierID int    `json:"courier_id"`
}
