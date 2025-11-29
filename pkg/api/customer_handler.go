package api

import (
	"encoding/json"
	"net/http"

	"github.com/NaphatPRM/golang-matching-system/pkg/models"
	"github.com/NaphatPRM/golang-matching-system/pkg/services"
	"github.com/gorilla/mux"
)

// CustomerHandler handles customer-related API requests
type CustomerHandler struct {
	service *services.CustomerService
}

// NewCustomerHandler creates a new CustomerHandler
func NewCustomerHandler(service *services.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

// RegisterRoutes registers customer routes
func (h *CustomerHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/customers", h.Register).Methods("POST")
	r.HandleFunc("/customers", h.GetAll).Methods("GET")
	r.HandleFunc("/customers/{id}", h.Get).Methods("GET")
	r.HandleFunc("/customers/{id}/location", h.UpdateLocation).Methods("PUT")
}

// Register handles customer registration
func (h *CustomerHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	customer, err := h.service.Register(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(customer)
}

// GetAll handles retrieving all customers
func (h *CustomerHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	customers := h.service.GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}

// Get handles retrieving a single customer
func (h *CustomerHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	customer, err := h.service.Get(id)
	if err != nil {
		if err == services.ErrCustomerNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

// UpdateLocation handles updating a customer's location
func (h *CustomerHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateCustomerLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	customer, err := h.service.UpdateLocation(id, &req)
	if err != nil {
		if err == services.ErrCustomerNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}
