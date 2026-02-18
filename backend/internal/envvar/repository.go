package envvar

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"wplatform/backend/internal/models"
)

type EnvironmentVariableRepository struct {
	db *sql.DB
}

func NewEnvironmentVariableRepository(db *sql.DB) *EnvironmentVariableRepository {
	return &EnvironmentVariableRepository{db: db}
}

func (r *EnvironmentVariableRepository) Create(ctx context.Context, envVar *models.EnvironmentVariable) error {
	query := `
		INSERT INTO environment_variables (id, project_id, key, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		envVar.ID,
		envVar.ProjectID,
		envVar.Key,
		envVar.Value,
		envVar.CreatedAt,
		envVar.UpdatedAt,
	)

	return err
}

func (r *EnvironmentVariableRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.EnvironmentVariable, error) {
	query := `
		SELECT id, project_id, key, value, created_at, updated_at
		FROM environment_variables
		WHERE project_id = $1
		ORDER BY key ASC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envVars []*models.EnvironmentVariable
	for rows.Next() {
		envVar := &models.EnvironmentVariable{}
		err := rows.Scan(
			&envVar.ID,
			&envVar.ProjectID,
			&envVar.Key,
			&envVar.Value,
			&envVar.CreatedAt,
			&envVar.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		envVars = append(envVars, envVar)
	}

	return envVars, nil
}

func (r *EnvironmentVariableRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.EnvironmentVariable, error) {
	query := `
		SELECT id, project_id, key, value, created_at, updated_at
		FROM environment_variables
		WHERE id = $1
	`

	envVar := &models.EnvironmentVariable{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&envVar.ID,
		&envVar.ProjectID,
		&envVar.Key,
		&envVar.Value,
		&envVar.CreatedAt,
		&envVar.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return envVar, nil
}

func (r *EnvironmentVariableRepository) Update(ctx context.Context, envVar *models.EnvironmentVariable) error {
	query := `
		UPDATE environment_variables
		SET key = $2, value = $3, updated_at = $4
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		envVar.ID,
		envVar.Key,
		envVar.Value,
		time.Now(),
	)

	return err
}

func (r *EnvironmentVariableRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM environment_variables WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *EnvironmentVariableRepository) DeleteByProjectID(ctx context.Context, projectID uuid.UUID) error {
	query := `DELETE FROM environment_variables WHERE project_id = $1`
	_, err := r.db.ExecContext(ctx, query, projectID)
	return err
}
