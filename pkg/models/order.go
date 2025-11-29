package models

import "time"

// OrderStatus represents the current status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"    // Order placed, awaiting driver assignment
	OrderStatusAssigned   OrderStatus = "assigned"   // Driver assigned, on the way to restaurant
	OrderStatusPickedUp   OrderStatus = "picked_up"  // Driver picked up the order
	OrderStatusDelivering OrderStatus = "delivering" // Driver on the way to customer
	OrderStatusDelivered  OrderStatus = "delivered"  // Order delivered successfully
	OrderStatusCancelled  OrderStatus = "cancelled"  // Order cancelled
)

// OrderItem represents an item in an order
type OrderItem struct {
	MenuItemID string  `json:"menu_item_id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	Quantity   int     `json:"quantity"`
}

// Order represents a customer order
type Order struct {
	ID                 string      `json:"id"`
	CustomerID         string      `json:"customer_id"`
	RestaurantID       string      `json:"restaurant_id"`
	DriverID           string      `json:"driver_id,omitempty"`
	Items              []OrderItem `json:"items"`
	Status             OrderStatus `json:"status"`
	TotalAmount        float64     `json:"total_amount"`
	DeliveryLocation   Location    `json:"delivery_location"`
	RestaurantLocation Location    `json:"restaurant_location"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
	AssignedAt         *time.Time  `json:"assigned_at,omitempty"`
	PickedUpAt         *time.Time  `json:"picked_up_at,omitempty"`
	DeliveredAt        *time.Time  `json:"delivered_at,omitempty"`
}

// PlaceOrderRequest represents a request to place a new order
type PlaceOrderRequest struct {
	CustomerID       string      `json:"customer_id"`
	RestaurantID     string      `json:"restaurant_id"`
	Items            []OrderItem `json:"items"`
	DeliveryLocation Location    `json:"delivery_location"`
}

// UpdateOrderStatusRequest represents a request to update an order's status
type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status"`
}
