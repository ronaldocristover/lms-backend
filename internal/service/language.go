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
	ErrLanguageNotFound = errors.New("language not found")
	ErrLanguageExists   = errors.New("language already exists")
)

type LanguageService interface {
	Create(ctx context.Context, req *model.CreateLanguageRequest) (*model.Language, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Language, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateLanguageRequest) (*model.Language, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *model.ListLanguagesRequest) ([]*model.Language, int64, error)
}

type languageService struct {
	repo repository.LanguageRepository
}

func NewLanguageService(repo repository.LanguageRepository) LanguageService {
	return &languageService{repo: repo}
}

func (s *languageService) Create(ctx context.Context, req *model.CreateLanguageRequest) (*model.Language, error) {
	existing, _ := s.repo.GetByCode(ctx, req.Code)
	if existing != nil {
		return nil, ErrLanguageExists
	}

	language := &model.Language{
		Code: req.Code,
		Name: req.Name,
	}

	if err := s.repo.Create(ctx, language); err != nil {
		return nil, apierror.Internal("Failed to create language")
	}

	return language, nil
}

func (s *languageService) GetByID(ctx context.Context, id uuid.UUID) (*model.Language, error) {
	language, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrLanguageNotFound
	}
	return language, nil
}

func (s *languageService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateLanguageRequest) (*model.Language, error) {
	language, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrLanguageNotFound
	}

	if req.Code != "" {
		existing, _ := s.repo.GetByCode(ctx, req.Code)
		if existing != nil && existing.ID != id {
			return nil, ErrLanguageExists
		}
		language.Code = req.Code
	}

	if req.Name != "" {
		language.Name = req.Name
	}

	if err := s.repo.Update(ctx, language); err != nil {
		return nil, apierror.Internal("Failed to update language")
	}

	return language, nil
}

func (s *languageService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrLanguageNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return apierror.Internal("Failed to delete language")
	}
	return nil
}

func (s *languageService) List(ctx context.Context, req *model.ListLanguagesRequest) ([]*model.Language, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	return s.repo.List(ctx, req)
}
