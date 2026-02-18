package domain

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type DomainHandler struct {
	validator *DNSValidator
}

func NewDomainHandler(validator *DNSValidator) *DomainHandler {
	return &DomainHandler{
		validator: validator,
	}
}

func (h *DomainHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/dns/check", h.CheckDNS).Methods("POST")
	r.HandleFunc("/projects/{project_id}/dns/check", h.CheckProjectDNS).Methods("GET")
}

type DNSCheckRequest struct {
	Domain string `json:"domain"`
}

type DNSCheckResponse struct {
	Domain      string   `json:"domain"`
	ServerIP    string   `json:"server_ip"`
	ResolvedIPs []string `json:"resolved_ips"`
	Propagated  bool     `json:"propagated"`
	Message     string   `json:"message"`
}

func (h *DomainHandler) CheckDNS(w http.ResponseWriter, r *http.Request) {
	var req DNSCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Domain == "" {
		http.Error(w, "Domain is required", http.StatusBadRequest)
		return
	}

	serverIP, err := h.validator.GetServerIPAddress()
	if err != nil {
		serverIP = "unknown"
	}

	response := DNSCheckResponse{
		Domain:   req.Domain,
		ServerIP: serverIP,
	}

	ips, err := h.validator.resolveDNSRecords(r.Context(), req.Domain)
	if err != nil {
		response.Propagated = false
		response.Message = err.Error()
		response.ResolvedIPs = []string{}
	} else {
		response.ResolvedIPs = ips
		for _, ip := range ips {
			if ip == serverIP {
				response.Propagated = true
				response.Message = "DNS record points to server"
				break
			}
		}
		if !response.Propagated {
			response.Message = "DNS record does not point to server IP yet"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *DomainHandler) CheckProjectDNS(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.PathValue("project_id")
	if projectIDStr == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	_, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "Domain query parameter is required", http.StatusBadRequest)
		return
	}

	serverIP, err := h.validator.GetServerIPAddress()
	if err != nil {
		serverIP = "unknown"
	}

	response := DNSCheckResponse{
		Domain:   domain,
		ServerIP: serverIP,
	}

	ips, err := h.validator.resolveDNSRecords(r.Context(), domain)
	if err != nil {
		response.Propagated = false
		response.Message = err.Error()
		response.ResolvedIPs = []string{}
	} else {
		response.ResolvedIPs = ips
		for _, ip := range ips {
			if ip == serverIP {
				response.Propagated = true
				response.Message = "DNS record points to server"
				break
			}
		}
		if !response.Propagated {
			response.Message = "DNS record does not point to server IP yet"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
