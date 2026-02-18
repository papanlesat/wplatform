package project

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"wplatform/backend/internal/auth"
)

type ProjectHandler struct {
	service *ProjectService
}

func NewProjectHandler(service *ProjectService) *ProjectHandler {
	return &ProjectHandler{
		service: service,
	}
}

func (h *ProjectHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/projects", h.CreateProject).Methods("POST")
	r.HandleFunc("/projects", h.ListProjects).Methods("GET")
	r.HandleFunc("/projects/{id}", h.GetProject).Methods("GET")
	r.HandleFunc("/projects/{id}", h.DeleteProject).Methods("DELETE")
	r.HandleFunc("/projects/{id}/start", h.StartProject).Methods("POST")
	r.HandleFunc("/projects/{id}/stop", h.StopProject).Methods("POST")
	r.HandleFunc("/projects/{id}/restart", h.RestartProject).Methods("POST")
	r.HandleFunc("/projects/{id}/health", h.CheckProjectHealth).Methods("GET")
	r.HandleFunc("/projects/{id}/logs", h.GetProjectLogs).Methods("GET")
}

func getUserID(r *http.Request) (uuid.UUID, bool) {
	userIDStr := auth.GetUserID(r)
	if userIDStr == "" {
		return uuid.Nil, false
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, false
	}
	return userID, true
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, ok := getUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response, err := h.service.CreateProject(r.Context(), userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.service.GetProject(r.Context(), id)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userID, ok := getUserID(r)
	if !ok || project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	projects, err := h.service.ListProjects(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (h *ProjectHandler) StartProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.service.GetProject(r.Context(), id)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userID, ok := getUserID(r)
	if !ok || project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.service.StartProject(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProjectHandler) StopProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.service.GetProject(r.Context(), id)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userID, ok := getUserID(r)
	if !ok || project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.service.StopProject(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProjectHandler) RestartProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.service.GetProject(r.Context(), id)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userID, ok := getUserID(r)
	if !ok || project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.service.RestartProject(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProjectHandler) CheckProjectHealth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.service.GetProject(r.Context(), id)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userID, ok := getUserID(r)
	if !ok || project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	status, err := h.service.CheckProjectHealth(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": string(status),
	})
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.service.GetProject(r.Context(), id)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userID, ok := getUserID(r)
	if !ok || project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.service.DeleteProject(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) GetProjectLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.service.GetProject(r.Context(), id)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userID, ok := getUserID(r)
	if !ok || project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	containerName := r.URL.Query().Get("container")
	tail := r.URL.Query().Get("tail")
	if tail == "" {
		tail = "100"
	}

	logs, err := h.service.GetProjectLogs(r.Context(), id, containerName, tail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer logs.Close()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	io.Copy(w, logs)
}
