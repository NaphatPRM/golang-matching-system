package services

import (
	"errors"
	"sync"
	"time"

	"github.com/NaphatPRM/golang-matching-system/pkg/models"
	"github.com/google/uuid"
)

var (
	ErrCustomerNotFound = errors.New("customer not found")
)

// CustomerService manages customer operations
type CustomerService struct {
	customers map[string]*models.Customer
	mu        sync.RWMutex
}

// NewCustomerService creates a new CustomerService
func NewCustomerService() *CustomerService {
	return &CustomerService{
		customers: make(map[string]*models.Customer),
	}
}

// Register registers a new customer
func (s *CustomerService) Register(req *models.RegisterCustomerRequest) (*models.Customer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	customer := &models.Customer{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Location:  req.Location,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.customers[customer.ID] = customer
	return customer, nil
}

// Get retrieves a customer by ID
func (s *CustomerService) Get(id string) (*models.Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	customer, exists := s.customers[id]
	if !exists {
		return nil, ErrCustomerNotFound
	}
	return customer, nil
}

// GetAll returns all customers
func (s *CustomerService) GetAll() []*models.Customer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	customers := make([]*models.Customer, 0, len(s.customers))
	for _, c := range s.customers {
		customers = append(customers, c)
	}
	return customers
}

// UpdateLocation updates a customer's location
func (s *CustomerService) UpdateLocation(id string, req *models.UpdateCustomerLocationRequest) (*models.Customer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	customer, exists := s.customers[id]
	if !exists {
		return nil, ErrCustomerNotFound
	}

	customer.Location = req.Location
	customer.UpdatedAt = time.Now()
	return customer, nil
}
