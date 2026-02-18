package auth

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"wplatform/backend/internal/models"
	"wplatform/backend/internal/validation"
)

func init() {
	validation.Init()
}

func setupTestHandler(t *testing.T) (*AuthHandler, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}

	userRepo := NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret-key", 24*time.Hour)
	authHandler := NewAuthHandler(authService)

	return authHandler, mock
}

func TestRegister_Success(t *testing.T) {
	handler, mock := setupTestHandler(t)
	defer mock.ExpectationsWereMet()

	email := "test@example.com"
	password := "password123"

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), email, sqlmock.AnyArg(), "user").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := map[string]string{
		"email":    email,
		"password": password,
		"role":     "user",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var response AuthResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response.Token == "" {
		t.Error("Expected token to be returned")
	}

	if response.User.Email != email {
		t.Errorf("Expected email %s, got %s", email, response.User.Email)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	handler, _ := setupTestHandler(t)

	body := map[string]string{
		"email":    "invalid-email",
		"password": "password123",
		"role":     "user",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	handler, _ := setupTestHandler(t)

	body := map[string]string{
		"email":    "test@example.com",
		"password": "short",
		"role":     "user",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestRegister_InvalidRole(t *testing.T) {
	handler, _ := setupTestHandler(t)

	body := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"role":     "superadmin",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	handler, mock := setupTestHandler(t)
	defer mock.ExpectationsWereMet()

	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := HashPassword(password)
	userID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "role", "created_at"}).
		AddRow(userID, email, hashedPassword, "user", time.Now())

	mock.ExpectQuery(`SELECT id, email, password_hash, role, created_at FROM users WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(rows)

	body := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response AuthResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response.Token == "" {
		t.Error("Expected token to be returned")
	}

	if response.User.Email != email {
		t.Errorf("Expected email %s, got %s", email, response.User.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	handler, mock := setupTestHandler(t)
	defer mock.ExpectationsWereMet()

	email := "test@example.com"
	correctPassword := "password123"
	hashedPassword, _ := HashPassword(correctPassword)
	userID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "role", "created_at"}).
		AddRow(userID, email, hashedPassword, "user", time.Now())

	mock.ExpectQuery(`SELECT id, email, password_hash, role, created_at FROM users WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(rows)

	body := map[string]string{
		"email":    email,
		"password": "wrongpassword",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	handler, mock := setupTestHandler(t)
	defer mock.ExpectationsWereMet()

	email := "nonexistent@example.com"

	mock.ExpectQuery(`SELECT id, email, password_hash, role, created_at FROM users WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	body := map[string]string{
		"email":    email,
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestPasswordHash(t *testing.T) {
	password := "mySecretPassword123"

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hashedPassword == "" {
		t.Error("Hashed password should not be empty")
	}

	if hashedPassword == password {
		t.Error("Hashed password should not equal plain password")
	}

	if err := CheckPassword(password, hashedPassword); err != nil {
		t.Errorf("Password check should succeed: %v", err)
	}

	if err := CheckPassword("wrongPassword", hashedPassword); err == nil {
		t.Error("Password check should fail with wrong password")
	}
}

func TestJWTService(t *testing.T) {
	jwtService := NewJWTService("test-secret-key", 24*time.Hour)

	userID := uuid.New().String()
	email := "test@example.com"
	role := "user"

	token, err := jwtService.GenerateToken(userID, email, role)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}

	if claims.Role != role {
		t.Errorf("Expected Role %s, got %s", role, claims.Role)
	}
}

func TestJWTService_InvalidToken(t *testing.T) {
	jwtService := NewJWTService("test-secret-key", 24*time.Hour)

	_, err := jwtService.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Should return error for invalid token")
	}
}

func TestUserRepository_Mock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	userID := uuid.New()
	email := "test@example.com"
	hashedPassword, _ := HashPassword("password123")

	t.Run("Create user", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(sqlmock.AnyArg(), email, sqlmock.AnyArg(), "user").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

		user := &models.User{
			ID:           userID,
			Email:        email,
			PasswordHash: hashedPassword,
			Role:         models.RoleUser,
		}

		if err := repo.Create(user); err != nil {
			t.Errorf("Failed to create user: %v", err)
		}
	})

	t.Run("Find by email", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "role", "created_at"}).
			AddRow(userID, email, hashedPassword, "user", time.Now())

		mock.ExpectQuery(`SELECT id, email, password_hash, role, created_at FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnRows(rows)

		user, err := repo.FindByEmail(email)
		if err != nil {
			t.Errorf("Failed to find user: %v", err)
		}

		if user.Email != email {
			t.Errorf("Expected email %s, got %s", email, user.Email)
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
