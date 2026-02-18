package ssl

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"wplatform/backend/internal/models"
)

type SSLRepository struct {
	db *sql.DB
}

func NewSSLRepository(db *sql.DB) *SSLRepository {
	return &SSLRepository{db: db}
}

func (r *SSLRepository) Create(ctx context.Context, domain *models.Domain) error {
	query := `
		INSERT INTO domains (id, project_id, domain_name, ssl_status, cloudflare_api_token, cloudflare_email, dns_challenge_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		domain.ID,
		domain.ProjectID,
		domain.DomainName,
		domain.SSLStatus,
		domain.CloudflareAPIToken,
		domain.CloudflareEmail,
		domain.DNSChallengeEnabled,
		domain.CreatedAt,
		domain.UpdatedAt,
	)

	return err
}

func (r *SSLRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*models.Domain, error) {
	query := `
		SELECT id, project_id, domain_name, ssl_status, cloudflare_api_token, cloudflare_email, dns_challenge_enabled, created_at, updated_at
		FROM domains
		WHERE project_id = $1
	`

	domain := &models.Domain{}
	err := r.db.QueryRowContext(ctx, query, projectID).Scan(
		&domain.ID,
		&domain.ProjectID,
		&domain.DomainName,
		&domain.SSLStatus,
		&domain.CloudflareAPIToken,
		&domain.CloudflareEmail,
		&domain.DNSChallengeEnabled,
		&domain.CreatedAt,
		&domain.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return domain, nil
}

func (r *SSLRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Domain, error) {
	query := `
		SELECT id, project_id, domain_name, ssl_status, cloudflare_api_token, cloudflare_email, dns_challenge_enabled, created_at, updated_at
		FROM domains
		WHERE id = $1
	`

	domain := &models.Domain{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&domain.ID,
		&domain.ProjectID,
		&domain.DomainName,
		&domain.SSLStatus,
		&domain.CloudflareAPIToken,
		&domain.CloudflareEmail,
		&domain.DNSChallengeEnabled,
		&domain.CreatedAt,
		&domain.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return domain, nil
}

func (r *SSLRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.SSLStatus) error {
	query := `UPDATE domains SET ssl_status = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status, time.Now())
	return err
}

func (r *SSLRepository) Update(ctx context.Context, domain *models.Domain) error {
	query := `
		UPDATE domains
		SET ssl_status = $2, cloudflare_api_token = $3, cloudflare_email = $4, dns_challenge_enabled = $5, updated_at = $6
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		domain.ID,
		domain.SSLStatus,
		domain.CloudflareAPIToken,
		domain.CloudflareEmail,
		domain.DNSChallengeEnabled,
		time.Now(),
	)

	return err
}

func (r *SSLRepository) DeleteByProjectID(ctx context.Context, projectID uuid.UUID) error {
	query := `DELETE FROM domains WHERE project_id = $1`
	_, err := r.db.ExecContext(ctx, query, projectID)
	return err
}
