package dispatch

import (
	"testing"
	"time"

	"github.com/NaphatPRM/golang-matching-system/pkg/models"
	"github.com/NaphatPRM/golang-matching-system/pkg/services"
)

func setupTestEnvironment() (*DispatchEngine, *services.DriverService, *services.OrderService, *services.RestaurantService, *services.CustomerService) {
	restaurantService := services.NewRestaurantService()
	customerService := services.NewCustomerService()
	driverService := services.NewDriverService()
	orderService := services.NewOrderService(restaurantService, customerService)

	config := DefaultDispatchConfig()
	config.AutoDispatchInterval = 100 * time.Millisecond
	dispatchEngine := NewDispatchEngine(config, driverService, orderService)

	return dispatchEngine, driverService, orderService, restaurantService, customerService
}

func TestDispatchEngine_FindBestDriver(t *testing.T) {
	engine, driverService, _, _, _ := setupTestEnvironment()

	// Register drivers at different locations
	driver1, _ := driverService.Register(&models.RegisterDriverRequest{Name: "Driver 1", MaxConcurrent: 3})
	driver2, _ := driverService.Register(&models.RegisterDriverRequest{Name: "Driver 2", MaxConcurrent: 3})
	driver3, _ := driverService.Register(&models.RegisterDriverRequest{Name: "Driver 3", MaxConcurrent: 3})

	// Set all drivers as available
	driverService.UpdateStatus(driver1.ID, &models.UpdateStatusRequest{Status: models.DriverStatusAvailable})
	driverService.UpdateStatus(driver2.ID, &models.UpdateStatusRequest{Status: models.DriverStatusAvailable})
	driverService.UpdateStatus(driver3.ID, &models.UpdateStatusRequest{Status: models.DriverStatusAvailable})

	// Set driver locations (driver1 closest, driver3 farthest)
	driverService.UpdateLocation(driver1.ID, &models.UpdateLocationRequest{
		Location: models.Location{Latitude: 13.7563, Longitude: 100.5018},
	})
	driverService.UpdateLocation(driver2.ID, &models.UpdateLocationRequest{
		Location: models.Location{Latitude: 13.7663, Longitude: 100.5018}, // ~1km away
	})
	driverService.UpdateLocation(driver3.ID, &models.UpdateLocationRequest{
		Location: models.Location{Latitude: 13.7863, Longitude: 100.5018}, // ~3km away
	})

	// Find best driver for location near driver1
	pickupLocation := models.Location{Latitude: 13.7563, Longitude: 100.5018}
	bestDriver, err := engine.FindBestDriver(pickupLocation)

	if err != nil {
		t.Fatalf("FindBestDriver() error = %v", err)
	}

	if bestDriver.Driver.ID != driver1.ID {
		t.Errorf("Expected driver1 (closest), got driver %s", bestDriver.Driver.Name)
	}
}

func TestDispatchEngine_FindBestDriver_ConsidersLoad(t *testing.T) {
	engine, driverService, _, _, _ := setupTestEnvironment()

	// Register two drivers at similar locations
	driver1, _ := driverService.Register(&models.RegisterDriverRequest{Name: "Driver 1", MaxConcurrent: 3})
	driver2, _ := driverService.Register(&models.RegisterDriverRequest{Name: "Driver 2", MaxConcurrent: 3})

	// Set both drivers as available
	driverService.UpdateStatus(driver1.ID, &models.UpdateStatusRequest{Status: models.DriverStatusAvailable})
	driverService.UpdateStatus(driver2.ID, &models.UpdateStatusRequest{Status: models.DriverStatusAvailable})

	// Set similar locations
	driverService.UpdateLocation(driver1.ID, &models.UpdateLocationRequest{
		Location: models.Location{Latitude: 13.7563, Longitude: 100.5018},
	})
	driverService.UpdateLocation(driver2.ID, &models.UpdateLocationRequest{
		Location: models.Location{Latitude: 13.7563, Longitude: 100.5018},
	})

	// Give driver1 some load
	driverService.AssignOrder(driver1.ID, "order1")
	driverService.AssignOrder(driver1.ID, "order2")

	// Find best driver - should prefer driver2 (lower load)
	pickupLocation := models.Location{Latitude: 13.7563, Longitude: 100.5018}
	bestDriver, err := engine.FindBestDriver(pickupLocation)

	if err != nil {
		t.Fatalf("FindBestDriver() error = %v", err)
	}

	if bestDriver.Driver.ID != driver2.ID {
		t.Errorf("Expected driver2 (lower load), got driver %s", bestDriver.Driver.Name)
	}
}

func TestDispatchEngine_NoAvailableDrivers(t *testing.T) {
	engine, driverService, _, _, _ := setupTestEnvironment()

	// Register a driver but don't set as available
	driverService.Register(&models.RegisterDriverRequest{Name: "Driver 1"})

	pickupLocation := models.Location{Latitude: 13.7563, Longitude: 100.5018}
	_, err := engine.FindBestDriver(pickupLocation)

	if err != ErrNoAvailableDrivers {
		t.Errorf("Expected ErrNoAvailableDrivers, got %v", err)
	}
}

func TestDispatchEngine_DispatchOrder(t *testing.T) {
	engine, driverService, orderService, restaurantService, customerService := setupTestEnvironment()

	// Register a restaurant
	restaurant, _ := restaurantService.Register(&models.RegisterRestaurantRequest{
		Name:     "Test Restaurant",
		Location: models.Location{Latitude: 13.7563, Longitude: 100.5018},
		Menu:     []models.MenuItem{{Name: "Burger", Price: 10.99}},
	})

	// Register a customer
	customer, _ := customerService.Register(&models.RegisterCustomerRequest{
		Name:     "Test Customer",
		Location: models.Location{Latitude: 13.7700, Longitude: 100.5100},
	})

	// Register and set up a driver
	driver, _ := driverService.Register(&models.RegisterDriverRequest{Name: "Test Driver", MaxConcurrent: 3})
	driverService.UpdateStatus(driver.ID, &models.UpdateStatusRequest{Status: models.DriverStatusAvailable})
	driverService.UpdateLocation(driver.ID, &models.UpdateLocationRequest{
		Location: models.Location{Latitude: 13.7563, Longitude: 100.5018},
	})

	// Place an order
	order, _ := orderService.PlaceOrder(&models.PlaceOrderRequest{
		CustomerID:   customer.ID,
		RestaurantID: restaurant.ID,
		Items: []models.OrderItem{
			{Name: "Burger", Price: 10.99, Quantity: 1},
		},
		DeliveryLocation: customer.Location,
	})

	// Dispatch the order
	err := engine.DispatchOrder(order.ID)
	if err != nil {
		t.Fatalf("DispatchOrder() error = %v", err)
	}

	// Verify order is assigned
	updatedOrder, _ := orderService.Get(order.ID)
	if updatedOrder.Status != models.OrderStatusAssigned {
		t.Errorf("Order status = %v, expected %v", updatedOrder.Status, models.OrderStatusAssigned)
	}

	if updatedOrder.DriverID != driver.ID {
		t.Errorf("DriverID = %v, expected %v", updatedOrder.DriverID, driver.ID)
	}

	// Verify driver has the order
	updatedDriver, _ := driverService.Get(driver.ID)
	found := false
	for _, oid := range updatedDriver.ActiveOrders {
		if oid == order.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Order not found in driver's active orders")
	}
}
