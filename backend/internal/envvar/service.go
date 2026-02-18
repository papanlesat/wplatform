package envvar

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"wplatform/backend/internal/docker"
	"wplatform/backend/internal/models"
)

type EnvironmentVariableService struct {
	repo         *EnvironmentVariableRepository
	containerMgr *docker.ContainerManager
}

func NewEnvironmentVariableService(
	repo *EnvironmentVariableRepository,
	containerMgr *docker.ContainerManager,
) *EnvironmentVariableService {
	return &EnvironmentVariableService{
		repo:         repo,
		containerMgr: containerMgr,
	}
}

type CreateEnvVarRequest struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type UpdateEnvVarRequest struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}

func (s *EnvironmentVariableService) CreateEnvVar(ctx context.Context, projectID uuid.UUID, req *CreateEnvVarRequest) (*models.EnvironmentVariable, error) {
	envVar := &models.EnvironmentVariable{
		ID:        uuid.New(),
		ProjectID: projectID,
		Key:       req.Key,
		Value:     req.Value,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, envVar); err != nil {
		return nil, fmt.Errorf("failed to create environment variable: %w", err)
	}

	return envVar, nil
}

func (s *EnvironmentVariableService) GetEnvVars(ctx context.Context, projectID uuid.UUID) ([]*models.EnvironmentVariable, error) {
	return s.repo.GetByProjectID(ctx, projectID)
}

func (s *EnvironmentVariableService) GetEnvVar(ctx context.Context, id uuid.UUID) (*models.EnvironmentVariable, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *EnvironmentVariableService) UpdateEnvVar(ctx context.Context, id uuid.UUID, req *UpdateEnvVarRequest) (*models.EnvironmentVariable, error) {
	envVar, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("environment variable not found: %w", err)
	}

	envVar.Key = req.Key
	envVar.Value = req.Value
	envVar.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, envVar); err != nil {
		return nil, fmt.Errorf("failed to update environment variable: %w", err)
	}

	return envVar, nil
}

func (s *EnvironmentVariableService) DeleteEnvVar(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete environment variable: %w", err)
	}

	return nil
}

func (s *EnvironmentVariableService) DeleteByProjectID(ctx context.Context, projectID uuid.UUID) error {
	return s.repo.DeleteByProjectID(ctx, projectID)
}

func (s *EnvironmentVariableService) GenerateDefaultWordPressSalts() map[string]string {
	return map[string]string{
		"WORDPRESS_AUTH_KEY":             generateRandomString(64),
		"WORDPRESS_SECURE_AUTH_KEY":      generateRandomString(64),
		"WORDPRESS_LOGGED_IN_KEY":        generateRandomString(64),
		"WORDPRESS_SECURE_LOGGED_IN_KEY": generateRandomString(64),
		"WORDPRESS_NONCE_KEY":            generateRandomString(64),
		"WORDPRESS_SECURE_NONCE_KEY":     generateRandomString(64),
	}
}

func (s *EnvironmentVariableService) GenerateDefaultPHPConfig() map[string]string {
	return map[string]string{
		"PHP_MEMORY_LIMIT":        "256M",
		"PHP_UPLOAD_MAX_FILESIZE": "64M",
		"PHP_POST_MAX_SIZE":       "64M",
		"PHP_MAX_EXECUTION_TIME":  "300",
	}
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			b[i] = charset[0]
			continue
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}
