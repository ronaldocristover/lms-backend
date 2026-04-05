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
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type UserService interface {
	Register(ctx context.Context, req *model.RegisterRequest) (*model.LoginResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*model.LoginResponse, error)
	Create(ctx context.Context, req *model.CreateUserRequest) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateUserRequest) (*model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *model.ListUsersRequest) ([]*model.User, int64, error)
}

type userService struct {
	repo           repository.UserRepository
	roleRepo       repository.RoleRepository
	jwtSecret      string
	jwtExpiry      time.Duration
	refreshExpiry  time.Duration
}

func NewUserService(repo repository.UserRepository, roleRepo repository.RoleRepository, jwtSecret string, jwtExpiry, refreshExpiry time.Duration) UserService {
	if jwtExpiry == 0 {
		jwtExpiry = 24 * time.Hour
	}
	if refreshExpiry == 0 {
		refreshExpiry = 7 * 24 * time.Hour // 7 days default
	}
	return &userService{
		repo:          repo,
		roleRepo:      roleRepo,
		jwtSecret:     jwtSecret,
		jwtExpiry:     jwtExpiry,
		refreshExpiry: refreshExpiry,
	}
}

type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Type   string    `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

func (s *userService) Register(ctx context.Context, req *model.RegisterRequest) (*model.LoginResponse, error) {
	existing, _ := s.repo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrUserExists
	}

	role, err := s.roleRepo.GetByID(ctx, req.RoleID)
	if err != nil {
		return nil, apierror.BadRequest("Invalid role")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apierror.Internal("Failed to hash password")
	}

	user := &model.User{
		Email:          req.Email,
		PasswordHash:   string(hashedPassword),
		Name:           req.Name,
		RoleID:         req.RoleID,
		Role:           role,
		OrganizationID: req.OrganizationID,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.generateTokenPair(user)
}

func (s *userService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokenPair(user)
}

func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*model.LoginResponse, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims.Type != "refresh" {
		return nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.UserID.String())
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return s.generateTokenPair(user)
}

func (s *userService) Create(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	existing, _ := s.repo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrUserExists
	}

	role, err := s.roleRepo.GetByID(ctx, req.RoleID)
	if err != nil {
		return nil, apierror.BadRequest("Invalid role")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apierror.Internal("Failed to hash password")
	}

	user := &model.User{
		Email:          req.Email,
		PasswordHash:   string(hashedPassword),
		Name:           req.Name,
		RoleID:         req.RoleID,
		Role:           role,
		OrganizationID: req.OrganizationID,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, apierror.Internal("Failed to create user")
	}

	return user, nil
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

	if req.Email != "" {
		existing, _ := s.repo.GetByEmail(ctx, req.Email)
		if existing != nil && existing.ID != id {
			return nil, ErrUserExists
		}
		user.Email = req.Email
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.RoleID != uuid.Nil {
		role, err := s.roleRepo.GetByID(ctx, req.RoleID)
		if err != nil {
			return nil, apierror.BadRequest("Invalid role")
		}
		user.RoleID = req.RoleID
		user.Role = role
	}
	if req.OrganizationID != nil {
		user.OrganizationID = req.OrganizationID
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, apierror.Internal("Failed to update user")
	}

	return user, nil
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return apierror.Internal("Failed to delete user")
	}
	return nil
}

func (s *userService) List(ctx context.Context, req *model.ListUsersRequest) ([]*model.User, int64, error) {
	return s.repo.List(ctx, req)
}

func (s *userService) generateTokenPair(user *model.User) (*model.LoginResponse, error) {
	accessToken, err := s.generateToken(user, "access", s.jwtExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(user, "refresh", s.refreshExpiry)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *userService) generateToken(user *model.User, tokenType string, expiry time.Duration) (string, error) {
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	now := time.Now()
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   roleName,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "lms-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *userService) parseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
