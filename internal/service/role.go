package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ronaldocristover/lms-backend/internal/model"
	"github.com/ronaldocristover/lms-backend/internal/repository"
	"github.com/ronaldocristover/lms-backend/pkg/apierror"
	"go.uber.org/zap"
)

var (
	ErrRoleNotFound = errors.New("role not found")
	ErrRoleExists   = errors.New("role already exists")
)

type RoleService interface {
	Create(ctx context.Context, req *model.CreateRoleRequest) (*model.Role, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Role, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateRoleRequest) (*model.Role, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *model.ListRolesRequest) ([]*model.Role, int64, error)
}

type roleService struct {
	repo repository.RoleRepository
	logger *zap.SugaredLogger
}

func NewRoleService(repo repository.RoleRepository, logger *zap.SugaredLogger) RoleService {
	return &roleService{repo: repo, logger: logger}
}

func (s *roleService) Create(ctx context.Context, req *model.CreateRoleRequest) (*model.Role, error) {
	existing, _ := s.repo.GetByName(ctx, req.Name)
	if existing != nil {
		return nil, ErrRoleExists
	}

	role := &model.Role{
		Name: req.Name,
	}

	if err := s.repo.Create(ctx, role); err != nil {
		s.logger.Errorw("operation failed", "error", err)
		return nil, apierror.Internal("Failed to create role")
	}

	s.logger.Infow("role created", "id", role.ID)
	return role, nil
}

func (s *roleService) GetByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	return role, nil
}

func (s *roleService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateRoleRequest) (*model.Role, error) {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrRoleNotFound
	}

	if req.Name != "" {
		existing, _ := s.repo.GetByName(ctx, req.Name)
		if existing != nil && existing.ID != id {
			return nil, ErrRoleExists
		}
		role.Name = req.Name
	}

	if err := s.repo.Update(ctx, role); err != nil {
		s.logger.Errorw("operation failed", "error", err)
		return nil, apierror.Internal("Failed to update role")
	}

	return role, nil
}

func (s *roleService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrRoleNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Errorw("operation failed", "error", err)
		return apierror.Internal("Failed to delete role")
	}
	return nil
}

func (s *roleService) List(ctx context.Context, req *model.ListRolesRequest) ([]*model.Role, int64, error) {
	return s.repo.List(ctx, req)
}
