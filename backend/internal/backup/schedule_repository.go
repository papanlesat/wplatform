package backup

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"wplatform/backend/internal/models"
)

type BackupScheduleRepository struct {
	db *sql.DB
}

func NewBackupScheduleRepository(db *sql.DB) *BackupScheduleRepository {
	return &BackupScheduleRepository{db: db}
}

func (r *BackupScheduleRepository) Create(ctx context.Context, schedule *models.BackupSchedule) error {
	query := `
		INSERT INTO backup_schedules (id, project_id, enabled, schedule_cron, retention_days, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		schedule.ID,
		schedule.ProjectID,
		schedule.Enabled,
		schedule.ScheduleCron,
		schedule.RetentionDays,
		schedule.CreatedAt,
		schedule.UpdatedAt,
	)

	return err
}

func (r *BackupScheduleRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*models.BackupSchedule, error) {
	query := `
		SELECT id, project_id, enabled, schedule_cron, retention_days, created_at, updated_at
		FROM backup_schedules
		WHERE project_id = $1
	`

	schedule := &models.BackupSchedule{}
	err := r.db.QueryRowContext(ctx, query, projectID).Scan(
		&schedule.ID,
		&schedule.ProjectID,
		&schedule.Enabled,
		&schedule.ScheduleCron,
		&schedule.RetentionDays,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return schedule, nil
}

func (r *BackupScheduleRepository) Update(ctx context.Context, schedule *models.BackupSchedule) error {
	query := `
		UPDATE backup_schedules
		SET enabled = $2, schedule_cron = $3, retention_days = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		schedule.ID,
		schedule.Enabled,
		schedule.ScheduleCron,
		schedule.RetentionDays,
		schedule.UpdatedAt,
	)
	return err
}

func (r *BackupScheduleRepository) ListEnabled(ctx context.Context) ([]*models.BackupSchedule, error) {
	query := `
		SELECT id, project_id, enabled, schedule_cron, retention_days, created_at, updated_at
		FROM backup_schedules
		WHERE enabled = true
		ORDER BY project_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.BackupSchedule
	for rows.Next() {
		schedule := &models.BackupSchedule{}
		err := rows.Scan(
			&schedule.ID,
			&schedule.ProjectID,
			&schedule.Enabled,
			&schedule.ScheduleCron,
			&schedule.RetentionDays,
			&schedule.CreatedAt,
			&schedule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}
