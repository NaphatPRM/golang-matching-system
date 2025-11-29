package models

import "time"

// MenuItem represents an item in a restaurant's menu
type MenuItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Available   bool    `json:"available"`
}

// Restaurant represents a restaurant registered in the system
type Restaurant struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Location  Location   `json:"location"`
	Menu      []MenuItem `json:"menu"`
	IsOpen    bool       `json:"is_open"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// RegisterRestaurantRequest represents a request to register a new restaurant
type RegisterRestaurantRequest struct {
	Name     string     `json:"name"`
	Location Location   `json:"location"`
	Menu     []MenuItem `json:"menu"`
}

// UpdateMenuRequest represents a request to update a restaurant's menu
type UpdateMenuRequest struct {
	Menu []MenuItem `json:"menu"`
}
