package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/lms/internal/model"
	"github.com/yourusername/lms/internal/repository"
	"github.com/yourusername/lms/pkg/apierror"
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
}

func NewRoleService(repo repository.RoleRepository) RoleService {
	return &roleService{repo: repo}
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
		return nil, apierror.Internal("Failed to create role")
	}

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
		return apierror.Internal("Failed to delete role")
	}
	return nil
}

func (s *roleService) List(ctx context.Context, req *model.ListRolesRequest) ([]*model.Role, int64, error) {
	return s.repo.List(ctx, req)
}
