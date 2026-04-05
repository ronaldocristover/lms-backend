package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ronaldocristover/lms-backend/internal/model"
	"github.com/ronaldocristover/lms-backend/internal/repository"
	"github.com/ronaldocristover/lms-backend/pkg/apierror"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryExists   = errors.New("category already exists")
)

type CategoryService interface {
	Create(ctx context.Context, req *model.CreateCategoryRequest) (*model.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateCategoryRequest) (*model.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *model.ListCategoriesRequest) ([]*model.Category, int64, error)
}

type categoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) Create(ctx context.Context, req *model.CreateCategoryRequest) (*model.Category, error) {
	existing, _ := s.repo.GetByName(ctx, req.Name)
	if existing != nil {
		return nil, ErrCategoryExists
	}

	category := &model.Category{
		Name: req.Name,
	}

	if err := s.repo.Create(ctx, category); err != nil {
		return nil, apierror.Internal("Failed to create category")
	}

	return category, nil
}

func (s *categoryService) GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCategoryNotFound
	}
	return category, nil
}

func (s *categoryService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateCategoryRequest) (*model.Category, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCategoryNotFound
	}

	if req.Name != "" {
		existing, _ := s.repo.GetByName(ctx, req.Name)
		if existing != nil && existing.ID != id {
			return nil, ErrCategoryExists
		}
		category.Name = req.Name
	}

	if err := s.repo.Update(ctx, category); err != nil {
		return nil, apierror.Internal("Failed to update category")
	}

	return category, nil
}

func (s *categoryService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrCategoryNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return apierror.Internal("Failed to delete category")
	}
	return nil
}

func (s *categoryService) List(ctx context.Context, req *model.ListCategoriesRequest) ([]*model.Category, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	return s.repo.List(ctx, req)
}
