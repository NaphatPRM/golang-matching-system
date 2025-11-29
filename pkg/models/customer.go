package models

import "time"

// Customer represents a customer in the system
type Customer struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Location  Location  `json:"location"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RegisterCustomerRequest represents a request to register a new customer
type RegisterCustomerRequest struct {
	Name     string   `json:"name"`
	Location Location `json:"location"`
}

// UpdateCustomerLocationRequest represents a request to update a customer's location
type UpdateCustomerLocationRequest struct {
	Location Location `json:"location"`
}
