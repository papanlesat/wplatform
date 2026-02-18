package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/google/uuid"
	"wplatform/backend/internal/docker"
	"wplatform/backend/internal/models"
	"wplatform/backend/internal/project"
)

type BackupService struct {
	backupRepo       *BackupRepository
	scheduleRepo     *BackupScheduleRepository
	projectRepo      *project.ProjectRepository
	dockerClient     *docker.Client
	containerManager *docker.ContainerManager
	backupBasePath   string
}

func NewBackupService(
	backupRepo *BackupRepository,
	scheduleRepo *BackupScheduleRepository,
	projectRepo *project.ProjectRepository,
	dockerClient *docker.Client,
	containerManager *docker.ContainerManager,
	backupBasePath string,
) *BackupService {
	return &BackupService{
		backupRepo:       backupRepo,
		scheduleRepo:     scheduleRepo,
		projectRepo:      projectRepo,
		dockerClient:     dockerClient,
		containerManager: containerManager,
		backupBasePath:   backupBasePath,
	}
}

type BackupRequest struct {
	ProjectID uuid.UUID `json:"project_id"`
	Type      string    `json:"type"`
}

type BackupResponse struct {
	BackupID  uuid.UUID `json:"backup_id"`
	FilePath  string    `json:"file_path"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *BackupService) CreateBackup(ctx context.Context, req *BackupRequest) (*BackupResponse, error) {
	project, err := s.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	mysqlContainerName := fmt.Sprintf("mysql-%s", project.ID)
	wordpressContainerName := fmt.Sprintf("wordpress-%s", project.ID)

	mysqlContainerID, err := s.dockerClient.GetContainerID(ctx, mysqlContainerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get MySQL container: %w", err)
	}

	wpContainerID, err := s.dockerClient.GetContainerID(ctx, wordpressContainerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get WordPress container: %w", err)
	}

	backupDir := filepath.Join(s.backupBasePath, project.ID.String())
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	backupFilename := fmt.Sprintf("backup-%s.tar.gz", timestamp)
	backupFilePath := filepath.Join(backupDir, backupFilename)

	backupFile, err := os.Create(backupFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer backupFile.Close()

	gzWriter := gzip.NewWriter(backupFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	if err := s.backupDatabase(ctx, mysqlContainerID, project.DBName, project.DBUser, project.DBPassword, tarWriter, timestamp); err != nil {
		os.Remove(backupFilePath)
		return nil, fmt.Errorf("failed to backup database: %w", err)
	}

	if err := s.backupWpContent(ctx, wpContainerID, tarWriter); err != nil {
		os.Remove(backupFilePath)
		return nil, fmt.Errorf("failed to backup wp-content: %w", err)
	}

	fileInfo, err := os.Stat(backupFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup file size: %w", err)
	}

	backupType := models.BackupTypeManual
	if req.Type == "scheduled" {
		backupType = models.BackupTypeScheduled
	}

	backup := &models.Backup{
		ID:        uuid.New(),
		ProjectID: req.ProjectID,
		FilePath:  backupFilePath,
		SizeBytes: fileInfo.Size(),
		Type:      backupType,
		CreatedAt: time.Now(),
	}

	if err := s.backupRepo.Create(ctx, backup); err != nil {
		os.Remove(backupFilePath)
		return nil, fmt.Errorf("failed to save backup metadata: %w", err)
	}

	return &BackupResponse{
		BackupID:  backup.ID,
		FilePath:  backup.FilePath,
		SizeBytes: backup.SizeBytes,
		CreatedAt: backup.CreatedAt,
	}, nil
}

func (s *BackupService) backupDatabase(ctx context.Context, containerID, dbName, dbUser, dbPassword string, tarWriter *tar.Writer, timestamp string) error {
	cmd := []string{
		"mysqldump",
		"-u", dbUser,
		fmt.Sprintf("-p%s", dbPassword),
		dbName,
	}

	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: false,
		Tty:          false,
	}

	execID, err := s.dockerClient.ExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	hijackedResp, err := s.dockerClient.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer hijackedResp.Close()

	dbDumpName := fmt.Sprintf("database-dump-%s.sql", timestamp)
	header := &tar.Header{
		Name:    dbDumpName,
		Mode:    0644,
		Size:    -1,
		ModTime: time.Now(),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	if _, err := io.Copy(tarWriter, hijackedResp.Reader); err != nil {
		return fmt.Errorf("failed to copy mysqldump output: %w", err)
	}

	if _, err := s.dockerClient.ExecInspect(ctx, execID.ID); err != nil {
		return fmt.Errorf("failed to inspect exec: %w", err)
	}

	return nil
}

func (s *BackupService) backupWpContent(ctx context.Context, containerID string, tarWriter *tar.Writer) error {
	cmd := []string{
		"tar",
		"czf",
		"-",
		"-C",
		"/var/www/html",
		"wp-content",
	}

	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: false,
		Tty:          false,
	}

	execID, err := s.dockerClient.ExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	hijackedResp, err := s.dockerClient.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer hijackedResp.Close()

	header := &tar.Header{
		Name:    "wp-content.tar.gz",
		Mode:    0644,
		Size:    -1,
		ModTime: time.Now(),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	if _, err := io.Copy(tarWriter, hijackedResp.Reader); err != nil {
		return fmt.Errorf("failed to copy wp-content archive: %w", err)
	}

	if _, err := s.dockerClient.ExecInspect(ctx, execID.ID); err != nil {
		return fmt.Errorf("failed to inspect exec: %w", err)
	}

	return nil
}

func (s *BackupService) ListBackups(ctx context.Context, projectID uuid.UUID) ([]*models.Backup, error) {
	return s.backupRepo.ListByProjectID(ctx, projectID)
}

func (s *BackupService) GetBackup(ctx context.Context, backupID uuid.UUID) (*models.Backup, error) {
	return s.backupRepo.GetByID(ctx, backupID)
}

func (s *BackupService) DeleteBackup(ctx context.Context, backupID uuid.UUID) error {
	backup, err := s.backupRepo.GetByID(ctx, backupID)
	if err != nil {
		return fmt.Errorf("failed to get backup: %w", err)
	}

	if err := os.Remove(backup.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete backup file: %w", err)
	}

	if err := s.backupRepo.Delete(ctx, backupID); err != nil {
		return fmt.Errorf("failed to delete backup record: %w", err)
	}

	return nil
}

func (s *BackupService) ApplyRetentionPolicy(ctx context.Context, projectID uuid.UUID, retentionDays int) error {
	if err := s.backupRepo.DeleteOldBackups(ctx, projectID, retentionDays); err != nil {
		return fmt.Errorf("failed to delete old backups from database: %w", err)
	}

	backupDir := filepath.Join(s.backupBasePath, projectID.String())
	if err := os.MkdirAll(backupDir, 0755); err == nil {
		cutoff := time.Now().AddDate(0, 0, -retentionDays)
		files, err := os.ReadDir(backupDir)
		if err == nil {
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				if strings.HasPrefix(file.Name(), "backup-") && strings.HasSuffix(file.Name(), ".tar.gz") {
					info, err := file.Info()
					if err != nil {
						continue
					}
					if info.ModTime().Before(cutoff) {
						os.Remove(filepath.Join(backupDir, file.Name()))
					}
				}
			}
		}
	}

	return nil
}

func (s *BackupService) RestoreBackup(ctx context.Context, backupID uuid.UUID) error {
	backup, err := s.backupRepo.GetByID(ctx, backupID)
	if err != nil {
		return fmt.Errorf("failed to get backup: %w", err)
	}

	project, err := s.projectRepo.GetByID(ctx, backup.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	mysqlContainerName := fmt.Sprintf("mysql-%s", project.ID)
	wordpressContainerName := fmt.Sprintf("wordpress-%s", project.ID)

	mysqlContainerID, err := s.dockerClient.GetContainerID(ctx, mysqlContainerName)
	if err != nil {
		return fmt.Errorf("failed to get MySQL container: %w", err)
	}

	wpContainerID, err := s.dockerClient.GetContainerID(ctx, wordpressContainerName)
	if err != nil {
		return fmt.Errorf("failed to get WordPress container: %w", err)
	}

	if err := s.restoreDatabase(ctx, mysqlContainerID, project.DBName, project.DBUser, project.DBPassword, backup.FilePath); err != nil {
		return fmt.Errorf("failed to restore database: %w", err)
	}

	if err := s.restoreWpContent(ctx, wpContainerID, backup.FilePath); err != nil {
		return fmt.Errorf("failed to restore wp-content: %w", err)
	}

	return nil
}

func (s *BackupService) restoreDatabase(ctx context.Context, containerID, dbName, dbUser, dbPassword, backupFilePath string) error {
	cmd := []string{
		"sh",
		"-c",
		fmt.Sprintf("tar -xzf - --wildcards 'database-dump-*.sql' --to-command 'mysql -u%s -p%s %s'",
			dbUser,
			dbPassword,
			dbName,
		),
	}

	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}

	execID, err := s.dockerClient.ExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	conn, err := s.dockerClient.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return fmt.Errorf("failed to start exec: %w", err)
	}
	defer conn.Close()

	if _, err := s.dockerClient.ExecInspect(ctx, execID.ID); err != nil {
		return fmt.Errorf("failed to inspect exec: %w", err)
	}

	return nil
}

func (s *BackupService) restoreWpContent(ctx context.Context, containerID, backupFilePath string) error {
	cmd := []string{
		"sh",
		"-c",
		"tar -xzf - --wildcards 'wp-content.tar.gz' -C /var/www/html",
	}

	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}

	execID, err := s.dockerClient.ExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	conn, err := s.dockerClient.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return fmt.Errorf("failed to start exec: %w", err)
	}
	defer conn.Close()

	if _, err := s.dockerClient.ExecInspect(ctx, execID.ID); err != nil {
		return fmt.Errorf("failed to inspect exec: %w", err)
	}

	return nil
}
