package backup

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"wplatform/backend/internal/models"
)

type BackupRepository struct {
	db *sql.DB
}

func NewBackupRepository(db *sql.DB) *BackupRepository {
	return &BackupRepository{db: db}
}

func (r *BackupRepository) Create(ctx context.Context, backup *models.Backup) error {
	query := `
		INSERT INTO backups (id, project_id, file_path, size_bytes, type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		backup.ID,
		backup.ProjectID,
		backup.FilePath,
		backup.SizeBytes,
		backup.Type,
		backup.CreatedAt,
	)

	return err
}

func (r *BackupRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Backup, error) {
	query := `
		SELECT id, project_id, file_path, size_bytes, type, created_at
		FROM backups
		WHERE id = $1
	`

	backup := &models.Backup{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&backup.ID,
		&backup.ProjectID,
		&backup.FilePath,
		&backup.SizeBytes,
		&backup.Type,
		&backup.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return backup, nil
}

func (r *BackupRepository) ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.Backup, error) {
	query := `
		SELECT id, project_id, file_path, size_bytes, type, created_at
		FROM backups
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []*models.Backup
	for rows.Next() {
		backup := &models.Backup{}
		err := rows.Scan(
			&backup.ID,
			&backup.ProjectID,
			&backup.FilePath,
			&backup.SizeBytes,
			&backup.Type,
			&backup.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		backups = append(backups, backup)
	}

	return backups, nil
}

func (r *BackupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM backups WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *BackupRepository) DeleteOldBackups(ctx context.Context, projectID uuid.UUID, retentionDays int) error {
	query := `
		DELETE FROM backups
		WHERE project_id = $1
		AND created_at < NOW() - INTERVAL '1 day' * $2
	`
	_, err := r.db.ExecContext(ctx, query, projectID, retentionDays)
	return err
}

func (r *BackupRepository) GetLatestBackup(ctx context.Context, projectID uuid.UUID) (*models.Backup, error) {
	query := `
		SELECT id, project_id, file_path, size_bytes, type, created_at
		FROM backups
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	backup := &models.Backup{}
	err := r.db.QueryRowContext(ctx, query, projectID).Scan(
		&backup.ID,
		&backup.ProjectID,
		&backup.FilePath,
		&backup.SizeBytes,
		&backup.Type,
		&backup.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return backup, nil
}
