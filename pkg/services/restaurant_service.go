package services

import (
	"errors"
	"sync"
	"time"

	"github.com/NaphatPRM/golang-matching-system/pkg/models"
	"github.com/google/uuid"
)

var (
	ErrRestaurantNotFound = errors.New("restaurant not found")
	ErrMenuItemNotFound   = errors.New("menu item not found")
)

// RestaurantService manages restaurant operations
type RestaurantService struct {
	restaurants map[string]*models.Restaurant
	mu          sync.RWMutex
}

// NewRestaurantService creates a new RestaurantService
func NewRestaurantService() *RestaurantService {
	return &RestaurantService{
		restaurants: make(map[string]*models.Restaurant),
	}
}

// Register registers a new restaurant
func (s *RestaurantService) Register(req *models.RegisterRestaurantRequest) (*models.Restaurant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	restaurant := &models.Restaurant{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Location:  req.Location,
		Menu:      req.Menu,
		IsOpen:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Assign IDs to menu items if not provided
	for i := range restaurant.Menu {
		if restaurant.Menu[i].ID == "" {
			restaurant.Menu[i].ID = uuid.New().String()
		}
		restaurant.Menu[i].Available = true
	}

	s.restaurants[restaurant.ID] = restaurant
	return restaurant, nil
}

// Get retrieves a restaurant by ID
func (s *RestaurantService) Get(id string) (*models.Restaurant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	restaurant, exists := s.restaurants[id]
	if !exists {
		return nil, ErrRestaurantNotFound
	}
	return restaurant, nil
}

// GetAll returns all restaurants
func (s *RestaurantService) GetAll() []*models.Restaurant {
	s.mu.RLock()
	defer s.mu.RUnlock()

	restaurants := make([]*models.Restaurant, 0, len(s.restaurants))
	for _, r := range s.restaurants {
		restaurants = append(restaurants, r)
	}
	return restaurants
}

// UpdateMenu updates a restaurant's menu
func (s *RestaurantService) UpdateMenu(id string, req *models.UpdateMenuRequest) (*models.Restaurant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	restaurant, exists := s.restaurants[id]
	if !exists {
		return nil, ErrRestaurantNotFound
	}

	// Assign IDs to new menu items
	for i := range req.Menu {
		if req.Menu[i].ID == "" {
			req.Menu[i].ID = uuid.New().String()
		}
	}

	restaurant.Menu = req.Menu
	restaurant.UpdatedAt = time.Now()
	return restaurant, nil
}

// SetOpenStatus sets the open/closed status of a restaurant
func (s *RestaurantService) SetOpenStatus(id string, isOpen bool) (*models.Restaurant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	restaurant, exists := s.restaurants[id]
	if !exists {
		return nil, ErrRestaurantNotFound
	}

	restaurant.IsOpen = isOpen
	restaurant.UpdatedAt = time.Now()
	return restaurant, nil
}

// GetMenuItem gets a specific menu item from a restaurant
func (s *RestaurantService) GetMenuItem(restaurantID, menuItemID string) (*models.MenuItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	restaurant, exists := s.restaurants[restaurantID]
	if !exists {
		return nil, ErrRestaurantNotFound
	}

	for _, item := range restaurant.Menu {
		if item.ID == menuItemID {
			return &item, nil
		}
	}
	return nil, ErrMenuItemNotFound
}

// GetNearbyRestaurants returns restaurants within a certain distance (km) from a location
func (s *RestaurantService) GetNearbyRestaurants(location models.Location, maxDistanceKm float64) []*models.Restaurant {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nearby := make([]*models.Restaurant, 0)
	for _, r := range s.restaurants {
		if r.IsOpen && location.DistanceTo(r.Location) <= maxDistanceKm {
			nearby = append(nearby, r)
		}
	}
	return nearby
}
