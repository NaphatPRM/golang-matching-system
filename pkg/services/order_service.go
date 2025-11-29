package services

import (
	"errors"
	"sync"
	"time"

	"github.com/NaphatPRM/golang-matching-system/pkg/models"
	"github.com/google/uuid"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidOrderStatus = errors.New("invalid order status transition")
)

// OrderService manages order operations
type OrderService struct {
	orders            map[string]*models.Order
	mu                sync.RWMutex
	restaurantService *RestaurantService
	customerService   *CustomerService
}

// NewOrderService creates a new OrderService
func NewOrderService(restaurantService *RestaurantService, customerService *CustomerService) *OrderService {
	return &OrderService{
		orders:            make(map[string]*models.Order),
		restaurantService: restaurantService,
		customerService:   customerService,
	}
}

// PlaceOrder creates a new order
func (s *OrderService) PlaceOrder(req *models.PlaceOrderRequest) (*models.Order, error) {
	// Validate restaurant exists and is open
	restaurant, err := s.restaurantService.Get(req.RestaurantID)
	if err != nil {
		return nil, err
	}
	if !restaurant.IsOpen {
		return nil, errors.New("restaurant is closed")
	}

	// Validate customer exists
	_, err = s.customerService.Get(req.CustomerID)
	if err != nil {
		return nil, err
	}

	// Calculate total amount
	var totalAmount float64
	for _, item := range req.Items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	order := &models.Order{
		ID:                 uuid.New().String(),
		CustomerID:         req.CustomerID,
		RestaurantID:       req.RestaurantID,
		Items:              req.Items,
		Status:             models.OrderStatusPending,
		TotalAmount:        totalAmount,
		DeliveryLocation:   req.DeliveryLocation,
		RestaurantLocation: restaurant.Location,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	s.orders[order.ID] = order
	return order, nil
}

// Get retrieves an order by ID
func (s *OrderService) Get(id string) (*models.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, exists := s.orders[id]
	if !exists {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

// GetAll returns all orders
func (s *OrderService) GetAll() []*models.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]*models.Order, 0, len(s.orders))
	for _, o := range s.orders {
		orders = append(orders, o)
	}
	return orders
}

// GetByCustomer returns all orders for a customer
func (s *OrderService) GetByCustomer(customerID string) []*models.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]*models.Order, 0)
	for _, o := range s.orders {
		if o.CustomerID == customerID {
			orders = append(orders, o)
		}
	}
	return orders
}

// GetByDriver returns all orders assigned to a driver
func (s *OrderService) GetByDriver(driverID string) []*models.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]*models.Order, 0)
	for _, o := range s.orders {
		if o.DriverID == driverID {
			orders = append(orders, o)
		}
	}
	return orders
}

// GetPendingOrders returns all orders waiting for driver assignment
func (s *OrderService) GetPendingOrders() []*models.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pending := make([]*models.Order, 0)
	for _, o := range s.orders {
		if o.Status == models.OrderStatusPending {
			pending = append(pending, o)
		}
	}
	return pending
}

// AssignDriver assigns a driver to an order
func (s *OrderService) AssignDriver(orderID, driverID string) (*models.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[orderID]
	if !exists {
		return nil, ErrOrderNotFound
	}

	if order.Status != models.OrderStatusPending {
		return nil, ErrInvalidOrderStatus
	}

	now := time.Now()
	order.DriverID = driverID
	order.Status = models.OrderStatusAssigned
	order.AssignedAt = &now
	order.UpdatedAt = now
	return order, nil
}

// UpdateStatus updates the status of an order
func (s *OrderService) UpdateStatus(id string, req *models.UpdateOrderStatusRequest) (*models.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[id]
	if !exists {
		return nil, ErrOrderNotFound
	}

	// Validate status transition
	if !isValidStatusTransition(order.Status, req.Status) {
		return nil, ErrInvalidOrderStatus
	}

	now := time.Now()
	order.Status = req.Status
	order.UpdatedAt = now

	switch req.Status {
	case models.OrderStatusPickedUp:
		order.PickedUpAt = &now
	case models.OrderStatusDelivered:
		order.DeliveredAt = &now
	}

	return order, nil
}

// isValidStatusTransition checks if a status transition is valid
func isValidStatusTransition(from, to models.OrderStatus) bool {
	validTransitions := map[models.OrderStatus][]models.OrderStatus{
		models.OrderStatusPending:    {models.OrderStatusAssigned, models.OrderStatusCancelled},
		models.OrderStatusAssigned:   {models.OrderStatusPickedUp, models.OrderStatusCancelled},
		models.OrderStatusPickedUp:   {models.OrderStatusDelivering, models.OrderStatusCancelled},
		models.OrderStatusDelivering: {models.OrderStatusDelivered, models.OrderStatusCancelled},
		models.OrderStatusDelivered:  {},
		models.OrderStatusCancelled:  {},
	}

	validTargets, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, valid := range validTargets {
		if valid == to {
			return true
		}
	}
	return false
}
