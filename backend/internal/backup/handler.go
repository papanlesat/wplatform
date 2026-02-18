package backup

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"wplatform/backend/internal/project"
)

type BackupHandler struct {
	service      *BackupService
	scheduleRepo *BackupScheduleRepository
	projectRepo  *project.ProjectRepository
}

func NewBackupHandler(
	service *BackupService,
	scheduleRepo *BackupScheduleRepository,
	projectRepo *project.ProjectRepository,
) *BackupHandler {
	return &BackupHandler{
		service:      service,
		scheduleRepo: scheduleRepo,
		projectRepo:  projectRepo,
	}
}

func (h *BackupHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/backups", h.CreateBackup).Methods("POST")
	r.HandleFunc("/projects/{id}/backups", h.ListBackups).Methods("GET")
	r.HandleFunc("/backups/{id}", h.GetBackup).Methods("GET")
	r.HandleFunc("/backups/{id}", h.DeleteBackup).Methods("DELETE")
	r.HandleFunc("/backups/{id}/download", h.DownloadBackup).Methods("GET")
	r.HandleFunc("/backups/{id}/restore", h.RestoreBackup).Methods("POST")
	r.HandleFunc("/projects/{id}/schedule", h.UpdateBackupSchedule).Methods("PUT")
	r.HandleFunc("/projects/{id}/schedule", h.GetBackupSchedule).Methods("GET")
}

func (h *BackupHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	var req BackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), req.ProjectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userIDStr := r.Context().Value("user_id")
	if userIDStr == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, _ := uuid.Parse(userIDStr.(string))
	if project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	response, err := h.service.CreateBackup(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *BackupHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	projectID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), projectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userIDStr := r.Context().Value("user_id")
	if userIDStr == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, _ := uuid.Parse(userIDStr.(string))
	if project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	backups, err := h.service.ListBackups(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(backups)
}

func (h *BackupHandler) GetBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	backupID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid backup ID", http.StatusBadRequest)
		return
	}

	backup, err := h.service.GetBackup(r.Context(), backupID)
	if err != nil {
		http.Error(w, "Backup not found", http.StatusNotFound)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), backup.ProjectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userIDStr := r.Context().Value("user_id")
	if userIDStr == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, _ := uuid.Parse(userIDStr.(string))
	if project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(backup)
}

func (h *BackupHandler) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	backupID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid backup ID", http.StatusBadRequest)
		return
	}

	backup, err := h.service.GetBackup(r.Context(), backupID)
	if err != nil {
		http.Error(w, "Backup not found", http.StatusNotFound)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), backup.ProjectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userIDStr := r.Context().Value("user_id")
	if userIDStr == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, _ := uuid.Parse(userIDStr.(string))
	if project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.service.DeleteBackup(r.Context(), backupID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BackupHandler) DownloadBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	backupID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid backup ID", http.StatusBadRequest)
		return
	}

	backup, err := h.service.GetBackup(r.Context(), backupID)
	if err != nil {
		http.Error(w, "Backup not found", http.StatusNotFound)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), backup.ProjectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userIDStr := r.Context().Value("user_id")
	if userIDStr == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, _ := uuid.Parse(userIDStr.(string))
	if project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if _, err := os.Stat(backup.FilePath); os.IsNotExist(err) {
		http.Error(w, "Backup file not found", http.StatusNotFound)
		return
	}

	file, err := os.Open(backup.FilePath)
	if err != nil {
		http.Error(w, "Failed to open backup file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(backup.FilePath)))
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", backup.SizeBytes))

	http.ServeContent(w, r, filepath.Base(backup.FilePath), backup.CreatedAt, file)
}

type ScheduleRequest struct {
	Enabled       bool   `json:"enabled"`
	ScheduleCron  string `json:"schedule_cron"`
	RetentionDays int    `json:"retention_days"`
}

func (h *BackupHandler) UpdateBackupSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	projectID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), projectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userIDStr := r.Context().Value("user_id")
	if userIDStr == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, _ := uuid.Parse(userIDStr.(string))
	if project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	schedule, err := h.scheduleRepo.GetByProjectID(r.Context(), projectID)
	if err != nil {
		http.Error(w, "Failed to get backup schedule", http.StatusInternalServerError)
		return
	}

	schedule.Enabled = req.Enabled
	schedule.ScheduleCron = req.ScheduleCron
	schedule.RetentionDays = req.RetentionDays

	if err := h.scheduleRepo.Update(r.Context(), schedule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

func (h *BackupHandler) GetBackupSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	projectID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), projectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userIDStr := r.Context().Value("user_id")
	if userIDStr == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, _ := uuid.Parse(userIDStr.(string))
	if project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	schedule, err := h.scheduleRepo.GetByProjectID(r.Context(), projectID)
	if err != nil {
		http.Error(w, "Failed to get backup schedule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

func (h *BackupHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	backupID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid backup ID", http.StatusBadRequest)
		return
	}

	backup, err := h.service.GetBackup(r.Context(), backupID)
	if err != nil {
		http.Error(w, "Backup not found", http.StatusNotFound)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), backup.ProjectID)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	userIDStr := r.Context().Value("user_id")
	if userIDStr == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, _ := uuid.Parse(userIDStr.(string))
	if project.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.service.RestoreBackup(r.Context(), backupID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "restored",
	})
}
