package project

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"wplatform/backend/internal/docker"
	"wplatform/backend/internal/domain"
	"wplatform/backend/internal/models"
)

type ProjectService struct {
	repo            *ProjectRepository
	dockerClient    *docker.Client
	networkMgr      *docker.NetworkManager
	volumeMgr       *docker.VolumeManager
	containerMgr    *docker.ContainerManager
	domainValidator domain.DomainValidator
	serverIP        string
}

func NewProjectService(
	repo *ProjectRepository,
	dockerClient *docker.Client,
	networkMgr *docker.NetworkManager,
	volumeMgr *docker.VolumeManager,
	containerMgr *docker.ContainerManager,
	domainValidator domain.DomainValidator,
	serverIP string,
) *ProjectService {
	return &ProjectService{
		repo:            repo,
		dockerClient:    dockerClient,
		networkMgr:      networkMgr,
		volumeMgr:       volumeMgr,
		containerMgr:    containerMgr,
		domainValidator: domainValidator,
		serverIP:        serverIP,
	}
}

type CreateProjectRequest struct {
	Name          string  `json:"name"`
	Domain        string  `json:"domain"`
	CPULimit      float64 `json:"cpu_limit"`
	MemoryLimitMB int     `json:"memory_limit_mb"`
}

type CreateProjectResponse struct {
	Project *models.Project `json:"project"`
	Message string          `json:"message"`
}

func (s *ProjectService) CreateProject(ctx context.Context, userID uuid.UUID, req *CreateProjectRequest) (*CreateProjectResponse, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}
	if req.CPULimit <= 0 {
		return nil, fmt.Errorf("CPU limit must be positive")
	}
	if req.MemoryLimitMB <= 0 {
		return nil, fmt.Errorf("memory limit must be positive")
	}

	// Set default resource limits if not provided
	if req.CPULimit == 0 {
		req.CPULimit = 0.5 // 0.5 CPU cores
	}
	if req.MemoryLimitMB == 0 {
		req.MemoryLimitMB = 512 // 512 MB
	}

	// Verify domain ownership
	if err := s.domainValidator.VerifyDomainOwnership(ctx, req.Domain, s.serverIP); err != nil {
		return nil, fmt.Errorf("domain ownership verification failed: %w", err)
	}

	// Generate project ID
	projectID := uuid.New()

	// Generate database credentials
	dbName := fmt.Sprintf("wp_%s", hex.EncodeToString(projectID[:])[:8])
	dbUser := fmt.Sprintf("wp_user_%s", hex.EncodeToString(projectID[:])[:8])
	dbPassword, err := s.generateRandomPassword(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate database password: %w", err)
	}

	// Create project in database with status "creating"
	project := &models.Project{
		ID:            projectID,
		UserID:        userID,
		Name:          req.Name,
		Domain:        req.Domain,
		Status:        models.StatusCreating,
		CPULimit:      req.CPULimit,
		MemoryLimitMB: req.MemoryLimitMB,
		DBName:        dbName,
		DBUser:        dbUser,
		DBPassword:    dbPassword,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project in database: %w", err)
	}

	// Provision Docker resources
	if err := s.provisionProjectResources(ctx, projectID, dbName, dbUser, dbPassword, req.Domain, req.CPULimit, req.MemoryLimitMB); err != nil {
		s.cleanupProjectResources(ctx, projectID)
		s.repo.UpdateStatus(ctx, projectID, models.StatusError)
		return nil, fmt.Errorf("failed to provision project resources: %w", err)
	}

	// Update project status to running
	project.Status = models.StatusRunning
	project.WordPressContainerName = fmt.Sprintf("wordpress-%s", projectID)
	project.MySQLContainerName = fmt.Sprintf("mysql-%s", projectID)

	if err := s.repo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project status: %w", err)
	}

	return &CreateProjectResponse{
		Project: project,
		Message: "Project created successfully",
	}, nil
}

func (s *ProjectService) provisionProjectResources(ctx context.Context, projectID uuid.UUID, dbName, dbUser, dbPassword, domain string, cpuLimit float64, memoryLimitMB int) error {
	// Create project network
	projectNetworkID, err := s.networkMgr.CreateProjectNetwork(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to create project network: %w", err)
	}

	// Create volumes
	wpVolumeName, dbVolumeName, err := s.volumeMgr.CreateProjectVolumes(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to create project volumes: %w", err)
	}

	// Get public network ID
	publicNetworkID, err := s.networkMgr.GetNetworkID(ctx, docker.PublicNetworkName)
	if err != nil {
		return fmt.Errorf("failed to get public network ID: %w", err)
	}

	// Create MySQL container
	mysqlContainerID, err := s.containerMgr.CreateMySQLContainer(ctx, projectID, dbName, dbUser, dbPassword, dbVolumeName, projectNetworkID)
	if err != nil {
		return fmt.Errorf("failed to create MySQL container: %w", err)
	}

	// Start MySQL container
	if err := s.containerMgr.StartContainer(ctx, mysqlContainerID); err != nil {
		return fmt.Errorf("failed to start MySQL container: %w", err)
	}

	// Wait for MySQL to be ready
	time.Sleep(10 * time.Second)

	// Pull WordPress image
	if err := s.containerMgr.PullImage(ctx, s.containerMgr.GetWordPressImage()); err != nil {
		return fmt.Errorf("failed to pull WordPress image: %w", err)
	}

	// Create WordPress container with default environment variables
	envVars := map[string]string{
		"WORDPRESS_DEBUG":        "0",
		"WORDPRESS_CONFIG_EXTRA": "define('WP_HOME', 'https://" + domain + "');\ndefine('WP_SITEURL', 'https://" + domain + "');",
	}

	wpContainerID, err := s.containerMgr.CreateWordPressContainer(
		ctx,
		projectID,
		domain,
		dbName,
		dbUser,
		dbPassword,
		wpVolumeName,
		projectNetworkID,
		publicNetworkID,
		cpuLimit,
		memoryLimitMB,
		envVars,
	)
	if err != nil {
		return fmt.Errorf("failed to create WordPress container: %w", err)
	}

	// Start WordPress container
	if err := s.containerMgr.StartContainer(ctx, wpContainerID); err != nil {
		return fmt.Errorf("failed to start WordPress container: %w", err)
	}

	return nil
}

func (s *ProjectService) GetProject(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProjectService) ListProjects(ctx context.Context, userID uuid.UUID) ([]*models.Project, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *ProjectService) StartProject(ctx context.Context, id uuid.UUID) error {
	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	mysqlContainerName := fmt.Sprintf("mysql-%s", id)
	wordpressContainerName := fmt.Sprintf("wordpress-%s", id)

	mysqlContainerID, err := s.dockerClient.GetContainerID(ctx, mysqlContainerName)
	if err != nil {
		return fmt.Errorf("MySQL container not found: %w", err)
	}

	wordpressContainerID, err := s.dockerClient.GetContainerID(ctx, wordpressContainerName)
	if err != nil {
		return fmt.Errorf("WordPress container not found: %w", err)
	}

	if err := s.containerMgr.StartContainer(ctx, mysqlContainerID); err != nil {
		return fmt.Errorf("failed to start MySQL container: %w", err)
	}

	time.Sleep(5 * time.Second)

	if err := s.containerMgr.StartContainer(ctx, wordpressContainerID); err != nil {
		return fmt.Errorf("failed to start WordPress container: %w", err)
	}

	project.Status = models.StatusRunning
	if err := s.repo.UpdateStatus(ctx, id, project.Status); err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}

	return nil
}

func (s *ProjectService) StopProject(ctx context.Context, id uuid.UUID) error {
	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	mysqlContainerName := fmt.Sprintf("mysql-%s", id)
	wordpressContainerName := fmt.Sprintf("wordpress-%s", id)

	wordpressContainerID, err := s.dockerClient.GetContainerID(ctx, wordpressContainerName)
	if err != nil {
		return fmt.Errorf("WordPress container not found: %w", err)
	}

	if err := s.containerMgr.StopContainer(ctx, wordpressContainerID); err != nil {
		return fmt.Errorf("failed to stop WordPress container: %w", err)
	}

	mysqlContainerID, err := s.dockerClient.GetContainerID(ctx, mysqlContainerName)
	if err != nil {
		return fmt.Errorf("MySQL container not found: %w", err)
	}

	if err := s.containerMgr.StopContainer(ctx, mysqlContainerID); err != nil {
		return fmt.Errorf("failed to stop MySQL container: %w", err)
	}

	project.Status = models.StatusStopped
	if err := s.repo.UpdateStatus(ctx, id, project.Status); err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}

	return nil
}

func (s *ProjectService) RestartProject(ctx context.Context, id uuid.UUID) error {
	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	mysqlContainerName := fmt.Sprintf("mysql-%s", id)
	wordpressContainerName := fmt.Sprintf("wordpress-%s", id)

	wordpressContainerID, err := s.dockerClient.GetContainerID(ctx, wordpressContainerName)
	if err != nil {
		return fmt.Errorf("WordPress container not found: %w", err)
	}

	if err := s.containerMgr.RestartContainer(ctx, wordpressContainerID); err != nil {
		return fmt.Errorf("failed to restart WordPress container: %w", err)
	}

	mysqlContainerID, err := s.dockerClient.GetContainerID(ctx, mysqlContainerName)
	if err != nil {
		return fmt.Errorf("MySQL container not found: %w", err)
	}

	if err := s.containerMgr.RestartContainer(ctx, mysqlContainerID); err != nil {
		return fmt.Errorf("failed to restart MySQL container: %w", err)
	}

	project.Status = models.StatusRunning
	if err := s.repo.UpdateStatus(ctx, id, project.Status); err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}

	return nil
}

func (s *ProjectService) CheckProjectHealth(ctx context.Context, id uuid.UUID) (models.ProjectStatus, error) {
	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("project not found: %w", err)
	}

	mysqlContainerName := fmt.Sprintf("mysql-%s", id)
	wordpressContainerName := fmt.Sprintf("wordpress-%s", id)

	mysqlContainerID, err := s.dockerClient.GetContainerID(ctx, mysqlContainerName)
	if err != nil {
		return models.StatusError, nil
	}

	wordpressContainerID, err := s.dockerClient.GetContainerID(ctx, wordpressContainerName)
	if err != nil {
		return models.StatusError, nil
	}

	mysqlContainer, err := s.dockerClient.InspectContainer(ctx, mysqlContainerID)
	if err != nil {
		return models.StatusError, nil
	}

	wordpressContainer, err := s.dockerClient.InspectContainer(ctx, wordpressContainerID)
	if err != nil {
		return models.StatusError, nil
	}

	mysqlRunning := mysqlContainer.State.Running
	wordpressRunning := wordpressContainer.State.Running

	if mysqlRunning && wordpressRunning {
		if project.Status != models.StatusRunning {
			project.Status = models.StatusRunning
			s.repo.UpdateStatus(ctx, id, project.Status)
		}
		return models.StatusRunning, nil
	}

	if !mysqlRunning && !wordpressRunning {
		if project.Status != models.StatusStopped {
			project.Status = models.StatusStopped
			s.repo.UpdateStatus(ctx, id, project.Status)
		}
		return models.StatusStopped, nil
	}

	project.Status = models.StatusError
	s.repo.UpdateStatus(ctx, id, project.Status)
	return models.StatusError, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	if err := s.containerMgr.RemoveProjectContainers(ctx, id); err != nil {
		return fmt.Errorf("failed to remove containers: %w", err)
	}

	if err := s.volumeMgr.RemoveProjectVolumes(ctx, id); err != nil {
		return fmt.Errorf("failed to remove volumes: %w", err)
	}

	if err := s.networkMgr.RemoveProjectNetwork(ctx, id); err != nil {
		return fmt.Errorf("failed to remove network: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete project from database: %w", err)
	}

	return nil
}

func (s *ProjectService) GetProjectLogs(ctx context.Context, id uuid.UUID, containerName string, tail string) (io.ReadCloser, error) {
	if containerName == "" {
		containerName = "wordpress"
	}

	targetContainerName := fmt.Sprintf("%s-%s", containerName, id)

	logs, err := s.dockerClient.GetContainerLogs(ctx, targetContainerName, tail)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	return logs, nil
}

func (s *ProjectService) generateRandomPassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

func (s *ProjectService) cleanupProjectResources(ctx context.Context, projectID uuid.UUID) {
	s.containerMgr.RemoveProjectContainers(ctx, projectID)
	s.volumeMgr.RemoveProjectVolumes(ctx, projectID)
	s.networkMgr.RemoveProjectNetwork(ctx, projectID)
}
