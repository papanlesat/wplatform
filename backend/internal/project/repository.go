package project

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"wplatform/backend/internal/models"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project *models.Project) error {
	query := `
		INSERT INTO projects (id, user_id, name, domain, status, cpu_limit, memory_limit_mb, wordpress_container_name, mysql_container_name, db_name, db_user, db_password, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		project.ID,
		project.UserID,
		project.Name,
		project.Domain,
		project.Status,
		project.CPULimit,
		project.MemoryLimitMB,
		project.WordPressContainerName,
		project.MySQLContainerName,
		project.DBName,
		project.DBUser,
		project.DBPassword,
		project.CreatedAt,
	)

	return err
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	query := `
		SELECT id, user_id, name, domain, status, cpu_limit, memory_limit_mb, wordpress_container_name, mysql_container_name, db_name, db_user, db_password, created_at
		FROM projects
		WHERE id = $1
	`

	project := &models.Project{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.UserID,
		&project.Name,
		&project.Domain,
		&project.Status,
		&project.CPULimit,
		&project.MemoryLimitMB,
		&project.WordPressContainerName,
		&project.MySQLContainerName,
		&project.DBName,
		&project.DBUser,
		&project.DBPassword,
		&project.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return project, nil
}

func (r *ProjectRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Project, error) {
	query := `
		SELECT id, user_id, name, domain, status, cpu_limit, memory_limit_mb, wordpress_container_name, mysql_container_name, db_name, db_user, db_password, created_at
		FROM projects
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		project := &models.Project{}
		err := rows.Scan(
			&project.ID,
			&project.UserID,
			&project.Name,
			&project.Domain,
			&project.Status,
			&project.CPULimit,
			&project.MemoryLimitMB,
			&project.WordPressContainerName,
			&project.MySQLContainerName,
			&project.DBName,
			&project.DBUser,
			&project.DBPassword,
			&project.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func (r *ProjectRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.ProjectStatus) error {
	query := `UPDATE projects SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *ProjectRepository) Update(ctx context.Context, project *models.Project) error {
	query := `
		UPDATE projects
		SET status = $2, wordpress_container_name = $3, mysql_container_name = $4
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		project.ID,
		project.Status,
		project.WordPressContainerName,
		project.MySQLContainerName,
	)
	return err
}

func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
