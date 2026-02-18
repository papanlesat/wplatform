package auth

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"wplatform/backend/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uuid.UUID) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
}

type SQLUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &SQLUserRepository{db: db}
}

func (r *SQLUserRepository) Create(user *models.User) error {
	query := `INSERT INTO users (id, email, password_hash, role) VALUES ($1, $2, $3, $4)`

	_, err := r.db.Exec(query, user.ID, user.Email, user.PasswordHash, user.Role)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *SQLUserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	query := `SELECT id, email, password_hash, role, created_at FROM users WHERE id = $1`

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

func (r *SQLUserRepository) FindByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`

	var user models.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return &user, nil
}
