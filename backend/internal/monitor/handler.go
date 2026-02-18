package monitor

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type MonitorHandler struct {
	service *MonitorService
}

func NewMonitorHandler(service *MonitorService) *MonitorHandler {
	return &MonitorHandler{
		service: service,
	}
}

func (h *MonitorHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/projects/{id}/stats", h.GetProjectStats).Methods("GET")
}

func (h *MonitorHandler) GetProjectStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	projectID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	stats, err := h.service.GetProjectStats(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
