package services

import (
	"errors"
	"sync"
	"time"

	"github.com/NaphatPRM/golang-matching-system/pkg/models"
	"github.com/google/uuid"
)

var (
	ErrDriverNotFound = errors.New("driver not found")
)

// DriverLocationUpdate represents a location update event
type DriverLocationUpdate struct {
	DriverID  string
	Location  models.Location
	Timestamp time.Time
}

// DriverService manages driver operations
type DriverService struct {
	drivers           map[string]*models.Driver
	mu                sync.RWMutex
	locationListeners []chan DriverLocationUpdate
	listenerMu        sync.RWMutex
}

// NewDriverService creates a new DriverService
func NewDriverService() *DriverService {
	return &DriverService{
		drivers:           make(map[string]*models.Driver),
		locationListeners: make([]chan DriverLocationUpdate, 0),
	}
}

// Register registers a new driver
func (s *DriverService) Register(req *models.RegisterDriverRequest) (*models.Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	maxConcurrent := req.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 3 // Default max concurrent orders
	}

	driver := &models.Driver{
		ID:            uuid.New().String(),
		Name:          req.Name,
		Location:      models.Location{},
		Status:        models.DriverStatusOffline,
		ActiveOrders:  make([]string, 0),
		MaxConcurrent: maxConcurrent,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	s.drivers[driver.ID] = driver
	return driver, nil
}

// Get retrieves a driver by ID
func (s *DriverService) Get(id string) (*models.Driver, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	driver, exists := s.drivers[id]
	if !exists {
		return nil, ErrDriverNotFound
	}
	return driver, nil
}

// GetAll returns all drivers
func (s *DriverService) GetAll() []*models.Driver {
	s.mu.RLock()
	defer s.mu.RUnlock()

	drivers := make([]*models.Driver, 0, len(s.drivers))
	for _, d := range s.drivers {
		drivers = append(drivers, d)
	}
	return drivers
}

// UpdateLocation updates a driver's location and notifies listeners
func (s *DriverService) UpdateLocation(id string, req *models.UpdateLocationRequest) (*models.Driver, error) {
	s.mu.Lock()

	driver, exists := s.drivers[id]
	if !exists {
		s.mu.Unlock()
		return nil, ErrDriverNotFound
	}

	now := time.Now()
	driver.Location = req.Location
	driver.LastLocationAt = now
	driver.UpdatedAt = now

	s.mu.Unlock()

	// Notify location listeners
	s.notifyLocationListeners(DriverLocationUpdate{
		DriverID:  id,
		Location:  req.Location,
		Timestamp: now,
	})

	return driver, nil
}

// UpdateStatus updates a driver's status
func (s *DriverService) UpdateStatus(id string, req *models.UpdateStatusRequest) (*models.Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	driver, exists := s.drivers[id]
	if !exists {
		return nil, ErrDriverNotFound
	}

	driver.Status = req.Status
	driver.UpdatedAt = time.Now()
	return driver, nil
}

// AssignOrder assigns an order to a driver
func (s *DriverService) AssignOrder(driverID, orderID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	driver, exists := s.drivers[driverID]
	if !exists {
		return ErrDriverNotFound
	}

	if !driver.IsAvailable() {
		return errors.New("driver is not available")
	}

	driver.ActiveOrders = append(driver.ActiveOrders, orderID)
	if len(driver.ActiveOrders) >= driver.MaxConcurrent {
		driver.Status = models.DriverStatusBusy
	}
	driver.UpdatedAt = time.Now()
	return nil
}

// CompleteOrder removes an order from a driver's active orders
func (s *DriverService) CompleteOrder(driverID, orderID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	driver, exists := s.drivers[driverID]
	if !exists {
		return ErrDriverNotFound
	}

	// Remove the order from active orders
	for i, id := range driver.ActiveOrders {
		if id == orderID {
			driver.ActiveOrders = append(driver.ActiveOrders[:i], driver.ActiveOrders[i+1:]...)
			break
		}
	}

	// Update status if driver now has capacity
	if driver.Status == models.DriverStatusBusy && len(driver.ActiveOrders) < driver.MaxConcurrent {
		driver.Status = models.DriverStatusAvailable
	}
	driver.UpdatedAt = time.Now()
	return nil
}

// GetAvailableDrivers returns all drivers that can accept new orders
func (s *DriverService) GetAvailableDrivers() []*models.Driver {
	s.mu.RLock()
	defer s.mu.RUnlock()

	available := make([]*models.Driver, 0)
	for _, d := range s.drivers {
		if d.IsAvailable() {
			available = append(available, d)
		}
	}
	return available
}

// SubscribeToLocationUpdates returns a channel that receives driver location updates
func (s *DriverService) SubscribeToLocationUpdates() chan DriverLocationUpdate {
	s.listenerMu.Lock()
	defer s.listenerMu.Unlock()

	ch := make(chan DriverLocationUpdate, 100)
	s.locationListeners = append(s.locationListeners, ch)
	return ch
}

// UnsubscribeFromLocationUpdates removes a listener channel
func (s *DriverService) UnsubscribeFromLocationUpdates(ch chan DriverLocationUpdate) {
	s.listenerMu.Lock()
	defer s.listenerMu.Unlock()

	for i, listener := range s.locationListeners {
		if listener == ch {
			s.locationListeners = append(s.locationListeners[:i], s.locationListeners[i+1:]...)
			close(ch)
			break
		}
	}
}

func (s *DriverService) notifyLocationListeners(update DriverLocationUpdate) {
	s.listenerMu.RLock()
	defer s.listenerMu.RUnlock()

	for _, ch := range s.locationListeners {
		select {
		case ch <- update:
		default:
			// Channel is full, skip this update
		}
	}
}
