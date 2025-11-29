package models

import "testing"

func TestDriver_Load(t *testing.T) {
	tests := []struct {
		name          string
		activeOrders  []string
		maxConcurrent int
		expected      float64
	}{
		{
			name:          "No orders",
			activeOrders:  []string{},
			maxConcurrent: 3,
			expected:      0.0,
		},
		{
			name:          "Half loaded",
			activeOrders:  []string{"order1"},
			maxConcurrent: 2,
			expected:      0.5,
		},
		{
			name:          "Fully loaded",
			activeOrders:  []string{"order1", "order2", "order3"},
			maxConcurrent: 3,
			expected:      1.0,
		},
		{
			name:          "Zero max concurrent",
			activeOrders:  []string{"order1"},
			maxConcurrent: 0,
			expected:      1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := &Driver{
				ActiveOrders:  tt.activeOrders,
				MaxConcurrent: tt.maxConcurrent,
			}
			if got := driver.Load(); got != tt.expected {
				t.Errorf("Load() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestDriver_IsAvailable(t *testing.T) {
	tests := []struct {
		name          string
		status        DriverStatus
		activeOrders  []string
		maxConcurrent int
		expected      bool
	}{
		{
			name:          "Available with no orders",
			status:        DriverStatusAvailable,
			activeOrders:  []string{},
			maxConcurrent: 3,
			expected:      true,
		},
		{
			name:          "Available with some orders",
			status:        DriverStatusAvailable,
			activeOrders:  []string{"order1"},
			maxConcurrent: 3,
			expected:      true,
		},
		{
			name:          "Available but at capacity",
			status:        DriverStatusAvailable,
			activeOrders:  []string{"order1", "order2", "order3"},
			maxConcurrent: 3,
			expected:      false,
		},
		{
			name:          "Offline",
			status:        DriverStatusOffline,
			activeOrders:  []string{},
			maxConcurrent: 3,
			expected:      false,
		},
		{
			name:          "Busy",
			status:        DriverStatusBusy,
			activeOrders:  []string{},
			maxConcurrent: 3,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := &Driver{
				Status:        tt.status,
				ActiveOrders:  tt.activeOrders,
				MaxConcurrent: tt.maxConcurrent,
			}
			if got := driver.IsAvailable(); got != tt.expected {
				t.Errorf("IsAvailable() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
