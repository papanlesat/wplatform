package envvar

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type EnvironmentVariableHandler struct {
	service *EnvironmentVariableService
}

func NewEnvironmentVariableHandler(service *EnvironmentVariableService) *EnvironmentVariableHandler {
	return &EnvironmentVariableHandler{
		service: service,
	}
}

func (h *EnvironmentVariableHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/projects/{project_id}/envvars", h.ListEnvVars).Methods("GET")
	r.HandleFunc("/projects/{project_id}/envvars", h.CreateEnvVar).Methods("POST")
	r.HandleFunc("/envvars/{id}", h.GetEnvVar).Methods("GET")
	r.HandleFunc("/envvars/{id}", h.UpdateEnvVar).Methods("PUT", "PATCH")
	r.HandleFunc("/envvars/{id}", h.DeleteEnvVar).Methods("DELETE")
	r.HandleFunc("/projects/{project_id}/envvars/generate-salts", h.GenerateSalts).Methods("POST")
}

func (h *EnvironmentVariableHandler) CreateEnvVar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectIDStr := vars["project_id"]

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var req CreateEnvVarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	envVar, err := h.service.CreateEnvVar(r.Context(), projectID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(envVar)
}

func (h *EnvironmentVariableHandler) ListEnvVars(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectIDStr := vars["project_id"]

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	envVars, err := h.service.GetEnvVars(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(envVars)
}

func (h *EnvironmentVariableHandler) GetEnvVar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid environment variable ID", http.StatusBadRequest)
		return
	}

	envVar, err := h.service.GetEnvVar(r.Context(), id)
	if err != nil {
		http.Error(w, "Environment variable not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(envVar)
}

func (h *EnvironmentVariableHandler) UpdateEnvVar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid environment variable ID", http.StatusBadRequest)
		return
	}

	var req UpdateEnvVarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	envVar, err := h.service.UpdateEnvVar(r.Context(), id, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(envVar)
}

func (h *EnvironmentVariableHandler) DeleteEnvVar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid environment variable ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteEnvVar(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *EnvironmentVariableHandler) GenerateSalts(w http.ResponseWriter, r *http.Request) {
	salts := h.service.GenerateDefaultWordPressSalts()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(salts)
}
