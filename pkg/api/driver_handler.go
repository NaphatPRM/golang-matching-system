package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/NaphatPRM/golang-matching-system/pkg/models"
	"github.com/NaphatPRM/golang-matching-system/pkg/services"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// DriverHandler handles driver-related API requests
type DriverHandler struct {
	service  *services.DriverService
	upgrader websocket.Upgrader
}

// NewDriverHandler creates a new DriverHandler
func NewDriverHandler(service *services.DriverService) *DriverHandler {
	return &DriverHandler{
		service: service,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}
}

// RegisterRoutes registers driver routes
func (h *DriverHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/drivers", h.Register).Methods("POST")
	r.HandleFunc("/drivers", h.GetAll).Methods("GET")
	r.HandleFunc("/drivers/{id}", h.Get).Methods("GET")
	r.HandleFunc("/drivers/{id}/location", h.UpdateLocation).Methods("PUT")
	r.HandleFunc("/drivers/{id}/status", h.UpdateStatus).Methods("PUT")
	r.HandleFunc("/drivers/available", h.GetAvailable).Methods("GET")
	r.HandleFunc("/drivers/locations/ws", h.WebSocketLocationUpdates).Methods("GET")
}

// Register handles driver registration
func (h *DriverHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterDriverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	driver, err := h.service.Register(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(driver)
}

// GetAll handles retrieving all drivers
func (h *DriverHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	drivers := h.service.GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(drivers)
}

// Get handles retrieving a single driver
func (h *DriverHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	driver, err := h.service.Get(id)
	if err != nil {
		if err == services.ErrDriverNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(driver)
}

// UpdateLocation handles updating a driver's location
func (h *DriverHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	driver, err := h.service.UpdateLocation(id, &req)
	if err != nil {
		if err == services.ErrDriverNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(driver)
}

// UpdateStatus handles updating a driver's status
func (h *DriverHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	driver, err := h.service.UpdateStatus(id, &req)
	if err != nil {
		if err == services.ErrDriverNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(driver)
}

// GetAvailable handles retrieving available drivers
func (h *DriverHandler) GetAvailable(w http.ResponseWriter, r *http.Request) {
	drivers := h.service.GetAvailableDrivers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(drivers)
}

// WebSocketLocationUpdates handles WebSocket connections for real-time location updates
func (h *DriverHandler) WebSocketLocationUpdates(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Subscribe to location updates
	updateChan := h.service.SubscribeToLocationUpdates()
	defer h.service.UnsubscribeFromLocationUpdates(updateChan)

	// Send location updates to the WebSocket client
	for update := range updateChan {
		err := conn.WriteJSON(update)
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			return
		}
	}
}
