package dispatch

import (
	"errors"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/NaphatPRM/golang-matching-system/pkg/models"
	"github.com/NaphatPRM/golang-matching-system/pkg/services"
)

var (
	ErrNoAvailableDrivers = errors.New("no available drivers")
)

// DispatchConfig contains configuration for the dispatch engine
type DispatchConfig struct {
	MaxSearchRadiusKm     float64       // Maximum radius to search for drivers
	DistanceWeight        float64       // Weight for distance factor (0-1)
	LoadWeight            float64       // Weight for load factor (0-1)
	AutoDispatchInterval  time.Duration // Interval for auto-dispatching pending orders
	MaxDriversToConsider  int           // Maximum number of drivers to consider for an order
}

// DefaultDispatchConfig returns a default configuration
func DefaultDispatchConfig() DispatchConfig {
	return DispatchConfig{
		MaxSearchRadiusKm:    10.0,
		DistanceWeight:       0.6,
		LoadWeight:           0.4,
		AutoDispatchInterval: 5 * time.Second,
		MaxDriversToConsider: 10,
	}
}

// DriverScore represents a driver with their calculated score
type DriverScore struct {
	Driver        *models.Driver
	Distance      float64
	LoadFactor    float64
	Score         float64
	EstimatedTime float64 // Estimated time to reach restaurant in minutes
}

// DispatchEngine handles driver assignment to orders
type DispatchEngine struct {
	config           DispatchConfig
	driverService    *services.DriverService
	orderService     *services.OrderService
	running          bool
	stopChan         chan struct{}
	mu               sync.RWMutex
	onAssignCallback func(orderID, driverID string)
}

// NewDispatchEngine creates a new dispatch engine
func NewDispatchEngine(
	config DispatchConfig,
	driverService *services.DriverService,
	orderService *services.OrderService,
) *DispatchEngine {
	return &DispatchEngine{
		config:        config,
		driverService: driverService,
		orderService:  orderService,
		stopChan:      make(chan struct{}),
	}
}

// SetOnAssignCallback sets a callback function that is called when an order is assigned
func (e *DispatchEngine) SetOnAssignCallback(callback func(orderID, driverID string)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onAssignCallback = callback
}

// Start starts the auto-dispatch loop
func (e *DispatchEngine) Start() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.stopChan = make(chan struct{})
	e.mu.Unlock()

	go e.autoDispatchLoop()
	log.Println("Dispatch engine started")
}

// Stop stops the auto-dispatch loop
func (e *DispatchEngine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return
	}

	close(e.stopChan)
	e.running = false
	log.Println("Dispatch engine stopped")
}

// autoDispatchLoop continuously tries to dispatch pending orders
func (e *DispatchEngine) autoDispatchLoop() {
	ticker := time.NewTicker(e.config.AutoDispatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopChan:
			return
		case <-ticker.C:
			e.dispatchPendingOrders()
		}
	}
}

// dispatchPendingOrders attempts to dispatch all pending orders
func (e *DispatchEngine) dispatchPendingOrders() {
	pendingOrders := e.orderService.GetPendingOrders()

	for _, order := range pendingOrders {
		err := e.DispatchOrder(order.ID)
		if err != nil {
			if !errors.Is(err, ErrNoAvailableDrivers) {
				log.Printf("Failed to dispatch order %s: %v", order.ID, err)
			}
		}
	}
}

// DispatchOrder assigns the best available driver to an order
func (e *DispatchEngine) DispatchOrder(orderID string) error {
	order, err := e.orderService.Get(orderID)
	if err != nil {
		return err
	}

	if order.Status != models.OrderStatusPending {
		return errors.New("order is not pending")
	}

	// Find the best driver
	bestDriver, err := e.FindBestDriver(order.RestaurantLocation)
	if err != nil {
		return err
	}

	// Assign the driver to the order
	_, err = e.orderService.AssignDriver(orderID, bestDriver.Driver.ID)
	if err != nil {
		return err
	}

	// Update driver's active orders
	err = e.driverService.AssignOrder(bestDriver.Driver.ID, orderID)
	if err != nil {
		return err
	}

	log.Printf("Order %s assigned to driver %s (distance: %.2f km, score: %.2f)",
		orderID, bestDriver.Driver.Name, bestDriver.Distance, bestDriver.Score)

	// Call the assignment callback if set
	e.mu.RLock()
	callback := e.onAssignCallback
	e.mu.RUnlock()

	if callback != nil {
		callback(orderID, bestDriver.Driver.ID)
	}

	return nil
}

// FindBestDriver finds the best available driver for a pickup location
func (e *DispatchEngine) FindBestDriver(pickupLocation models.Location) (*DriverScore, error) {
	availableDrivers := e.driverService.GetAvailableDrivers()
	if len(availableDrivers) == 0 {
		return nil, ErrNoAvailableDrivers
	}

	// Calculate scores for all drivers
	scoredDrivers := e.scoreDrivers(availableDrivers, pickupLocation)
	if len(scoredDrivers) == 0 {
		return nil, ErrNoAvailableDrivers
	}

	// Sort by score (higher is better)
	sort.Slice(scoredDrivers, func(i, j int) bool {
		return scoredDrivers[i].Score > scoredDrivers[j].Score
	})

	return scoredDrivers[0], nil
}

// scoreDrivers calculates scores for all drivers based on distance and load
func (e *DispatchEngine) scoreDrivers(drivers []*models.Driver, pickupLocation models.Location) []*DriverScore {
	scoredDrivers := make([]*DriverScore, 0, len(drivers))

	// First pass: calculate distances and find max distance
	var maxDistance float64
	driverDistances := make(map[string]float64)

	for _, driver := range drivers {
		distance := driver.Location.DistanceTo(pickupLocation)
		if distance > e.config.MaxSearchRadiusKm {
			continue // Skip drivers outside search radius
		}
		driverDistances[driver.ID] = distance
		if distance > maxDistance {
			maxDistance = distance
		}
	}

	// Handle case where all drivers are at the same location
	if maxDistance == 0 {
		maxDistance = 1
	}

	// Second pass: calculate scores
	for _, driver := range drivers {
		distance, exists := driverDistances[driver.ID]
		if !exists {
			continue
		}

		// Normalize distance (0 = farthest, 1 = closest)
		normalizedDistance := 1 - (distance / maxDistance)

		// Normalize load (0 = fully loaded, 1 = no load)
		normalizedLoad := 1 - driver.Load()

		// Calculate weighted score
		score := (e.config.DistanceWeight * normalizedDistance) +
			(e.config.LoadWeight * normalizedLoad)

		// Estimate time to reach restaurant (assuming average speed of 30 km/h)
		estimatedTime := (distance / 30) * 60 // in minutes

		scoredDrivers = append(scoredDrivers, &DriverScore{
			Driver:        driver,
			Distance:      distance,
			LoadFactor:    driver.Load(),
			Score:         score,
			EstimatedTime: estimatedTime,
		})
	}

	// Limit the number of drivers to consider
	if len(scoredDrivers) > e.config.MaxDriversToConsider {
		// Sort and take top N
		sort.Slice(scoredDrivers, func(i, j int) bool {
			return scoredDrivers[i].Score > scoredDrivers[j].Score
		})
		scoredDrivers = scoredDrivers[:e.config.MaxDriversToConsider]
	}

	return scoredDrivers
}

// GetDriverScores returns scored drivers for a pickup location (useful for debugging/monitoring)
func (e *DispatchEngine) GetDriverScores(pickupLocation models.Location) []*DriverScore {
	availableDrivers := e.driverService.GetAvailableDrivers()
	scoredDrivers := e.scoreDrivers(availableDrivers, pickupLocation)

	sort.Slice(scoredDrivers, func(i, j int) bool {
		return scoredDrivers[i].Score > scoredDrivers[j].Score
	})

	return scoredDrivers
}
