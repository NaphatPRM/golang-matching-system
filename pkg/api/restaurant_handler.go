package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/NaphatPRM/golang-matching-system/pkg/models"
	"github.com/NaphatPRM/golang-matching-system/pkg/services"
	"github.com/gorilla/mux"
)

// RestaurantHandler handles restaurant-related API requests
type RestaurantHandler struct {
	service *services.RestaurantService
}

// NewRestaurantHandler creates a new RestaurantHandler
func NewRestaurantHandler(service *services.RestaurantService) *RestaurantHandler {
	return &RestaurantHandler{service: service}
}

// RegisterRoutes registers restaurant routes
func (h *RestaurantHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/restaurants", h.Register).Methods("POST")
	r.HandleFunc("/restaurants", h.GetAll).Methods("GET")
	r.HandleFunc("/restaurants/{id}", h.Get).Methods("GET")
	r.HandleFunc("/restaurants/{id}/menu", h.UpdateMenu).Methods("PUT")
	r.HandleFunc("/restaurants/{id}/status", h.SetStatus).Methods("PUT")
	r.HandleFunc("/restaurants/nearby", h.GetNearby).Methods("GET")
}

// Register handles restaurant registration
func (h *RestaurantHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRestaurantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	restaurant, err := h.service.Register(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(restaurant)
}

// GetAll handles retrieving all restaurants
func (h *RestaurantHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	restaurants := h.service.GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurants)
}

// Get handles retrieving a single restaurant
func (h *RestaurantHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	restaurant, err := h.service.Get(id)
	if err != nil {
		if err == services.ErrRestaurantNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurant)
}

// UpdateMenu handles updating a restaurant's menu
func (h *RestaurantHandler) UpdateMenu(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	restaurant, err := h.service.UpdateMenu(id, &req)
	if err != nil {
		if err == services.ErrRestaurantNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurant)
}

// SetStatus handles setting restaurant open/closed status
func (h *RestaurantHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		IsOpen bool `json:"is_open"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	restaurant, err := h.service.SetOpenStatus(id, req.IsOpen)
	if err != nil {
		if err == services.ErrRestaurantNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurant)
}

// GetNearby handles retrieving nearby restaurants
func (h *RestaurantHandler) GetNearby(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	lat := query.Get("lat")
	lon := query.Get("lon")
	radius := query.Get("radius")

	var location models.Location
	var maxRadius float64 = 5.0 // Default 5km

	if lat != "" && lon != "" {
		latVal, err := strconv.ParseFloat(lat, 64)
		if err != nil {
			http.Error(w, "invalid latitude", http.StatusBadRequest)
			return
		}
		lonVal, err := strconv.ParseFloat(lon, 64)
		if err != nil {
			http.Error(w, "invalid longitude", http.StatusBadRequest)
			return
		}
		location.Latitude = latVal
		location.Longitude = lonVal
	}

	if radius != "" {
		radiusVal, err := strconv.ParseFloat(radius, 64)
		if err != nil {
			http.Error(w, "invalid radius", http.StatusBadRequest)
			return
		}
		maxRadius = radiusVal
	}

	restaurants := h.service.GetNearbyRestaurants(location, maxRadius)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurants)
}
