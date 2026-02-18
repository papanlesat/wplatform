package auth

import (
	"time"

	"github.com/google/uuid"
	"wplatform/backend/internal/models"
)

type AuthService struct {
	userRepo   UserRepository
	jwtService *JWTService
}

func NewAuthService(userRepo UserRepository, jwtSecret string, jwtExpiration time.Duration) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: NewJWTService(jwtSecret, jwtExpiration),
	}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"required,oneof=admin user"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         models.UserRole(req.Role),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	token, err := s.jwtService.GenerateToken(user.ID.String(), user.Email, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}

	if err := CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, err
	}

	token, err := s.jwtService.GenerateToken(user.ID.String(), user.Email, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) ValidateToken(token string) (*Claims, error) {
	return s.jwtService.ValidateToken(token)
}
