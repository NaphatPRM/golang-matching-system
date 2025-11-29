package services

import (
	"testing"
	"time"

	"github.com/NaphatPRM/golang-matching-system/pkg/models"
)

func TestDriverService_Register(t *testing.T) {
	service := NewDriverService()

	req := &models.RegisterDriverRequest{
		Name:          "Test Driver",
		MaxConcurrent: 3,
	}

	driver, err := service.Register(req)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if driver.Name != req.Name {
		t.Errorf("Name = %v, expected %v", driver.Name, req.Name)
	}

	if driver.MaxConcurrent != req.MaxConcurrent {
		t.Errorf("MaxConcurrent = %v, expected %v", driver.MaxConcurrent, req.MaxConcurrent)
	}

	if driver.Status != models.DriverStatusOffline {
		t.Errorf("Status = %v, expected %v", driver.Status, models.DriverStatusOffline)
	}
}

func TestDriverService_UpdateLocation(t *testing.T) {
	service := NewDriverService()

	// Register a driver
	driver, _ := service.Register(&models.RegisterDriverRequest{Name: "Test Driver"})

	// Update location
	newLocation := models.Location{Latitude: 13.7563, Longitude: 100.5018}
	updated, err := service.UpdateLocation(driver.ID, &models.UpdateLocationRequest{Location: newLocation})

	if err != nil {
		t.Fatalf("UpdateLocation() error = %v", err)
	}

	if updated.Location != newLocation {
		t.Errorf("Location = %v, expected %v", updated.Location, newLocation)
	}
}

func TestDriverService_UpdateStatus(t *testing.T) {
	service := NewDriverService()

	// Register a driver
	driver, _ := service.Register(&models.RegisterDriverRequest{Name: "Test Driver"})

	// Update status to available
	updated, err := service.UpdateStatus(driver.ID, &models.UpdateStatusRequest{Status: models.DriverStatusAvailable})

	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	if updated.Status != models.DriverStatusAvailable {
		t.Errorf("Status = %v, expected %v", updated.Status, models.DriverStatusAvailable)
	}
}

func TestDriverService_AssignAndCompleteOrder(t *testing.T) {
	service := NewDriverService()

	// Register and set available
	driver, _ := service.Register(&models.RegisterDriverRequest{Name: "Test Driver", MaxConcurrent: 2})
	service.UpdateStatus(driver.ID, &models.UpdateStatusRequest{Status: models.DriverStatusAvailable})

	// Assign orders
	err := service.AssignOrder(driver.ID, "order1")
	if err != nil {
		t.Fatalf("AssignOrder() error = %v", err)
	}

	err = service.AssignOrder(driver.ID, "order2")
	if err != nil {
		t.Fatalf("AssignOrder() second error = %v", err)
	}

	// Check driver is now busy
	updated, _ := service.Get(driver.ID)
	if updated.Status != models.DriverStatusBusy {
		t.Errorf("Status should be busy when at capacity, got %v", updated.Status)
	}

	// Should fail to assign third order
	err = service.AssignOrder(driver.ID, "order3")
	if err == nil {
		t.Error("AssignOrder() should fail when driver is not available")
	}

	// Complete an order
	err = service.CompleteOrder(driver.ID, "order1")
	if err != nil {
		t.Fatalf("CompleteOrder() error = %v", err)
	}

	// Check driver is available again
	updated, _ = service.Get(driver.ID)
	if updated.Status != models.DriverStatusAvailable {
		t.Errorf("Status should be available after completing order, got %v", updated.Status)
	}

	if len(updated.ActiveOrders) != 1 {
		t.Errorf("ActiveOrders should have 1 order, got %d", len(updated.ActiveOrders))
	}
}

func TestDriverService_LocationUpdates(t *testing.T) {
	service := NewDriverService()

	// Register a driver
	driver, _ := service.Register(&models.RegisterDriverRequest{Name: "Test Driver"})

	// Subscribe to location updates
	updateChan := service.SubscribeToLocationUpdates()

	// Update location in a goroutine
	go func() {
		time.Sleep(100 * time.Millisecond)
		service.UpdateLocation(driver.ID, &models.UpdateLocationRequest{
			Location: models.Location{Latitude: 13.7563, Longitude: 100.5018},
		})
	}()

	// Wait for update
	select {
	case update := <-updateChan:
		if update.DriverID != driver.ID {
			t.Errorf("DriverID = %v, expected %v", update.DriverID, driver.ID)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for location update")
	}

	service.UnsubscribeFromLocationUpdates(updateChan)
}
