package ssl

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"wplatform/backend/internal/docker"
	"wplatform/backend/internal/models"
)

type SSLService struct {
	repo      *SSLRepository
	dockerMgr *docker.ContainerManager
}

func NewSSLService(repo *SSLRepository, dockerMgr *docker.ContainerManager) *SSLService {
	return &SSLService{
		repo:      repo,
		dockerMgr: dockerMgr,
	}
}

type ProvisionSSLDomainRequest struct {
	ProjectID          uuid.UUID `json:"project_id"`
	DomainName         string    `json:"domain_name"`
	UseDNSChallenge    bool      `json:"use_dns_challenge"`
	UseCustomSSL       bool      `json:"use_custom_ssl"`
	CertificateData    string    `json:"certificate_data,omitempty"`
	CertificateKey     string    `json:"certificate_key,omitempty"`
	CloudflareAPIToken string    `json:"cloudflare_api_token,omitempty"`
	CloudflareEmail    string    `json:"cloudflare_email,omitempty"`
}

type SSLProvisionResponse struct {
	Message string           `json:"message"`
	Status  models.SSLStatus `json:"status"`
}

func (s *SSLService) ProvisionSSL(ctx context.Context, req *ProvisionSSLDomainRequest) (*SSLProvisionResponse, error) {
	domain := &models.Domain{
		ID:                  uuid.New(),
		ProjectID:           req.ProjectID,
		DomainName:          req.DomainName,
		SSLStatus:           models.SSLStatusPending,
		CloudflareAPIToken:  req.CloudflareAPIToken,
		CloudflareEmail:     req.CloudflareEmail,
		DNSChallengeEnabled: req.UseDNSChallenge,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if req.UseCustomSSL && req.CertificateData == "" {
		return nil, fmt.Errorf("certificate data is required for custom SSL")
	}

	if err := s.repo.Create(ctx, domain); err != nil {
		return nil, fmt.Errorf("failed to create domain record: %w", err)
	}

	if req.UseCustomSSL {
		domain.SSLStatus = models.SSLStatusActive
	} else {
		if err := s.restartWordPressForSSL(ctx, req.ProjectID); err != nil {
			return nil, fmt.Errorf("failed to restart WordPress for SSL provisioning: %w", err)
		}

		go s.waitForSSLCertificate(ctx, domain.ID)
	}

	if err := s.repo.Update(ctx, domain); err != nil {
		return nil, fmt.Errorf("failed to update domain status: %w", err)
	}

	return &SSLProvisionResponse{
		Message: "SSL provisioning initiated",
		Status:  domain.SSLStatus,
	}, nil
}

func (s *SSLService) restartWordPressForSSL(ctx context.Context, projectID uuid.UUID) error {
	wordpressContainerName := fmt.Sprintf("wordpress-%s", projectID)

	containerID, err := s.dockerMgr.GetContainerID(ctx, wordpressContainerName)
	if err != nil {
		return fmt.Errorf("WordPress container not found: %w", err)
	}

	if err := s.dockerMgr.RestartContainer(ctx, containerID); err != nil {
		return fmt.Errorf("failed to restart WordPress: %w", err)
	}

	return nil
}

func (s *SSLService) waitForSSLCertificate(ctx context.Context, domainID uuid.UUID) {
	const maxRetries = 5
	const retryInterval = 30 * time.Second

	for i := 0; i < maxRetries; i++ {
		time.Sleep(retryInterval)

		domain, err := s.repo.GetByID(ctx, domainID)
		if err != nil {
			return
		}

		if domain.SSLStatus == models.SSLStatusActive {
			return
		}

		if domain.SSLStatus == models.SSLStatusFailed {
			return
		}
	}
}

func (s *SSLService) GetSSLStatus(ctx context.Context, projectID uuid.UUID) (*models.Domain, error) {
	return s.repo.GetByProjectID(ctx, projectID)
}

func (s *SSLService) UpdateSSLStatus(ctx context.Context, domainID uuid.UUID, status models.SSLStatus) error {
	return s.repo.UpdateStatus(ctx, domainID, status)
}

func (s *SSLService) CheckSSLCertificate(ctx context.Context, domain string) (bool, error) {
	wordpressContainerName := "wordpress-" + domain

	containerID, err := s.dockerMgr.GetContainerID(ctx, wordpressContainerName)
	if err != nil {
		return false, fmt.Errorf("WordPress container not found: %w", err)
	}

	container, err := s.dockerMgr.InspectContainer(ctx, containerID)
	if err != nil {
		return false, fmt.Errorf("failed to inspect WordPress container: %w", err)
	}

	labels := container.Config.Labels
	if labels == nil {
		return false, nil
	}

	enabled, exists := labels["traefik.enable"]
	if !exists {
		return false, nil
	}

	return enabled == "true", nil
}
