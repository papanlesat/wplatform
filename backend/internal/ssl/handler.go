package ssl

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type SSLHandler struct {
	service *SSLService
}

func NewSSLHandler(service *SSLService) *SSLHandler {
	return &SSLHandler{
		service: service,
	}
}

func (h *SSLHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/ssl/provision", h.ProvisionSSL).Methods("POST")
	r.HandleFunc("/projects/{id}/ssl", h.GetSSLStatus).Methods("GET")
	r.HandleFunc("/domains/{domain}/check-certificate", h.CheckCertificate).Methods("GET")
}

func (h *SSLHandler) ProvisionSSL(w http.ResponseWriter, r *http.Request) {
	var req ProvisionSSLDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.service.ProvisionSSL(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SSLHandler) GetSSLStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	projectID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	status, err := h.service.GetSSLStatus(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *SSLHandler) CheckCertificate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domain := vars["domain"]

	valid, err := h.service.CheckSSLCertificate(r.Context(), domain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"valid": valid,
	})
}
