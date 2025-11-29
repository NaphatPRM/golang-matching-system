package models

import "time"

// DriverStatus represents the current status of a driver
type DriverStatus string

const (
	DriverStatusOffline   DriverStatus = "offline"
	DriverStatusAvailable DriverStatus = "available"
	DriverStatusBusy      DriverStatus = "busy"
)

// Driver represents a delivery driver in the system
type Driver struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Location       Location     `json:"location"`
	Status         DriverStatus `json:"status"`
	ActiveOrders   []string     `json:"active_orders"` // Order IDs currently being delivered
	MaxConcurrent  int          `json:"max_concurrent"` // Maximum concurrent orders a driver can handle
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	LastLocationAt time.Time    `json:"last_location_at"`
}

// RegisterDriverRequest represents a request to register a new driver
type RegisterDriverRequest struct {
	Name          string `json:"name"`
	MaxConcurrent int    `json:"max_concurrent"`
}

// UpdateLocationRequest represents a request to update a driver's location
type UpdateLocationRequest struct {
	Location Location `json:"location"`
}

// UpdateStatusRequest represents a request to update a driver's status
type UpdateStatusRequest struct {
	Status DriverStatus `json:"status"`
}

// Load returns the current load factor of the driver (active orders / max concurrent)
func (d *Driver) Load() float64 {
	if d.MaxConcurrent == 0 {
		return 1.0
	}
	return float64(len(d.ActiveOrders)) / float64(d.MaxConcurrent)
}

// IsAvailable returns true if the driver can accept more orders
func (d *Driver) IsAvailable() bool {
	return d.Status == DriverStatusAvailable && len(d.ActiveOrders) < d.MaxConcurrent
}
