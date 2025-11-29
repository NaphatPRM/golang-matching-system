package services

import (
	"testing"

	"github.com/NaphatPRM/golang-matching-system/pkg/models"
)

func TestRestaurantService_Register(t *testing.T) {
	service := NewRestaurantService()

	req := &models.RegisterRestaurantRequest{
		Name: "Test Restaurant",
		Location: models.Location{
			Latitude:  13.7563,
			Longitude: 100.5018,
		},
		Menu: []models.MenuItem{
			{Name: "Burger", Price: 10.99},
			{Name: "Fries", Price: 3.99},
		},
	}

	restaurant, err := service.Register(req)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if restaurant.Name != req.Name {
		t.Errorf("Name = %v, expected %v", restaurant.Name, req.Name)
	}

	if restaurant.Location != req.Location {
		t.Errorf("Location = %v, expected %v", restaurant.Location, req.Location)
	}

	if len(restaurant.Menu) != len(req.Menu) {
		t.Errorf("Menu length = %v, expected %v", len(restaurant.Menu), len(req.Menu))
	}

	// Verify menu items have IDs assigned
	for _, item := range restaurant.Menu {
		if item.ID == "" {
			t.Error("Menu item ID should not be empty")
		}
	}
}

func TestRestaurantService_Get(t *testing.T) {
	service := NewRestaurantService()

	// Register a restaurant
	req := &models.RegisterRestaurantRequest{
		Name:     "Test Restaurant",
		Location: models.Location{Latitude: 13.7563, Longitude: 100.5018},
	}
	registered, _ := service.Register(req)

	// Test Get
	restaurant, err := service.Get(registered.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if restaurant.ID != registered.ID {
		t.Errorf("ID = %v, expected %v", restaurant.ID, registered.ID)
	}

	// Test Get with non-existent ID
	_, err = service.Get("non-existent-id")
	if err != ErrRestaurantNotFound {
		t.Errorf("Expected ErrRestaurantNotFound, got %v", err)
	}
}

func TestRestaurantService_GetNearbyRestaurants(t *testing.T) {
	service := NewRestaurantService()

	// Register some restaurants at different locations
	locations := []models.Location{
		{Latitude: 13.7563, Longitude: 100.5018}, // Center
		{Latitude: 13.7663, Longitude: 100.5018}, // ~1km away
		{Latitude: 13.8563, Longitude: 100.5018}, // ~11km away
	}

	for i, loc := range locations {
		service.Register(&models.RegisterRestaurantRequest{
			Name:     "Restaurant " + string(rune('A'+i)),
			Location: loc,
		})
	}

	// Search within 5km radius
	center := models.Location{Latitude: 13.7563, Longitude: 100.5018}
	nearby := service.GetNearbyRestaurants(center, 5.0)

	if len(nearby) != 2 {
		t.Errorf("GetNearbyRestaurants() returned %d restaurants, expected 2", len(nearby))
	}
}
