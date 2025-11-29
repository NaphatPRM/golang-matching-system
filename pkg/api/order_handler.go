package api

import (
	"encoding/json"
	"net/http"

	"github.com/NaphatPRM/golang-matching-system/internal/dispatch"
	"github.com/NaphatPRM/golang-matching-system/pkg/models"
	"github.com/NaphatPRM/golang-matching-system/pkg/services"
	"github.com/gorilla/mux"
)

// OrderHandler handles order-related API requests
type OrderHandler struct {
	service        *services.OrderService
	driverService  *services.DriverService
	dispatchEngine *dispatch.DispatchEngine
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(
	service *services.OrderService,
	driverService *services.DriverService,
	dispatchEngine *dispatch.DispatchEngine,
) *OrderHandler {
	return &OrderHandler{
		service:        service,
		driverService:  driverService,
		dispatchEngine: dispatchEngine,
	}
}

// RegisterRoutes registers order routes
func (h *OrderHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/orders", h.PlaceOrder).Methods("POST")
	r.HandleFunc("/orders", h.GetAll).Methods("GET")
	r.HandleFunc("/orders/{id}", h.Get).Methods("GET")
	r.HandleFunc("/orders/{id}/status", h.UpdateStatus).Methods("PUT")
	r.HandleFunc("/orders/{id}/dispatch", h.ManualDispatch).Methods("POST")
	r.HandleFunc("/orders/pending", h.GetPending).Methods("GET")
	r.HandleFunc("/orders/customer/{customerId}", h.GetByCustomer).Methods("GET")
	r.HandleFunc("/orders/driver/{driverId}", h.GetByDriver).Methods("GET")
}

// PlaceOrder handles placing a new order
func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req models.PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := h.service.PlaceOrder(&req)
	if err != nil {
		if err == services.ErrRestaurantNotFound || err == services.ErrCustomerNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// GetAll handles retrieving all orders
func (h *OrderHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	orders := h.service.GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// Get handles retrieving a single order
func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	order, err := h.service.Get(id)
	if err != nil {
		if err == services.ErrOrderNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// UpdateStatus handles updating an order's status
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the order first to check if we need to update driver
	order, err := h.service.Get(id)
	if err != nil {
		if err == services.ErrOrderNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update order status
	updatedOrder, err := h.service.UpdateStatus(id, &req)
	if err != nil {
		if err == services.ErrOrderNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == services.ErrInvalidOrderStatus {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If order is delivered or cancelled, free up the driver
	if (req.Status == models.OrderStatusDelivered || req.Status == models.OrderStatusCancelled) && order.DriverID != "" {
		h.driverService.CompleteOrder(order.DriverID, id)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedOrder)
}

// ManualDispatch handles manually dispatching an order to an available driver
func (h *OrderHandler) ManualDispatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.dispatchEngine.DispatchOrder(id)
	if err != nil {
		if err == services.ErrOrderNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == dispatch.ErrNoAvailableDrivers {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return the updated order
	order, err := h.service.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// GetPending handles retrieving pending orders
func (h *OrderHandler) GetPending(w http.ResponseWriter, r *http.Request) {
	orders := h.service.GetPendingOrders()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// GetByCustomer handles retrieving orders for a customer
func (h *OrderHandler) GetByCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID := vars["customerId"]

	orders := h.service.GetByCustomer(customerID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// GetByDriver handles retrieving orders for a driver
func (h *OrderHandler) GetByDriver(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	driverID := vars["driverId"]

	orders := h.service.GetByDriver(driverID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
