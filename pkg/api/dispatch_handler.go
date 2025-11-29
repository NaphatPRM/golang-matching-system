package api

import (
	"encoding/json"
	"net/http"

	"github.com/NaphatPRM/golang-matching-system/internal/dispatch"
	"github.com/NaphatPRM/golang-matching-system/pkg/models"
	"github.com/gorilla/mux"
)

// DispatchHandler handles dispatch-related API requests
type DispatchHandler struct {
	engine *dispatch.DispatchEngine
}

// NewDispatchHandler creates a new DispatchHandler
func NewDispatchHandler(engine *dispatch.DispatchEngine) *DispatchHandler {
	return &DispatchHandler{engine: engine}
}

// RegisterRoutes registers dispatch routes
func (h *DispatchHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/dispatch/scores", h.GetDriverScores).Methods("POST")
}

// GetDriverScores returns scored drivers for a given pickup location
func (h *DispatchHandler) GetDriverScores(w http.ResponseWriter, r *http.Request) {
	var location models.Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	scores := h.engine.GetDriverScores(location)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}
