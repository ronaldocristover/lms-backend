package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourusername/lms/internal/model"
	"github.com/yourusername/lms/internal/repository"
	"github.com/yourusername/lms/pkg/apierror"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

type UserService interface {
	Register(ctx context.Context, req *model.RegisterRequest) (*model.LoginResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateUserRequest) (*model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error)
}

type userService struct {
	repo      repository.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewUserService(repo repository.UserRepository, jwtSecret string, jwtExpiry time.Duration) UserService {
	return &userService{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

func (s *userService) Register(ctx context.Context, req *model.RegisterRequest) (*model.LoginResponse, error) {
	// Check if user exists
	existing, _ := s.repo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apierror.Internal("Failed to hash password")
	}

	user := &model.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Role:         "user",
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate JWT
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *userService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateUserRequest) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Role != "" {
		user.Role = req.Role
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

func (s *userService) List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, pageSize, offset)
}

func (s *userService) generateToken(user *model.User) (string, error) {
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
