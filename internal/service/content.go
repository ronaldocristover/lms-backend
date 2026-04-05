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
	ErrContentNotFound = errors.New("content not found")
)

type ContentService interface {
	Create(ctx context.Context, req *model.CreateContentRequest) (*model.Content, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Content, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateContentRequest) (*model.Content, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *model.ListContentsRequest) ([]*model.Content, int64, error)
}

type contentService struct {
	repo        repository.ContentRepository
	sessionRepo repository.SessionRepository
}

func NewContentService(repo repository.ContentRepository, sessionRepo repository.SessionRepository) ContentService {
	return &contentService{
		repo:        repo,
		sessionRepo: sessionRepo,
	}
}

func (s *contentService) Create(ctx context.Context, req *model.CreateContentRequest) (*model.Content, error) {
	if _, err := s.sessionRepo.GetByID(ctx, req.SessionID); err != nil {
		return nil, ErrSessionNotFound
	}

	content := &model.Content{
		SessionID:   req.SessionID,
		Type:        req.Type,
		MediaID:     req.MediaID,
		ContentText: req.ContentText,
	}

	if err := s.repo.Create(ctx, content); err != nil {
		return nil, apierror.Internal("Failed to create content")
	}

	return content, nil
}

func (s *contentService) GetByID(ctx context.Context, id uuid.UUID) (*model.Content, error) {
	content, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrContentNotFound
	}
	return content, nil
}

func (s *contentService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateContentRequest) (*model.Content, error) {
	content, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrContentNotFound
	}

	if _, err := s.sessionRepo.GetByID(ctx, req.SessionID); err != nil {
		return nil, ErrSessionNotFound
	}

	content.SessionID = req.SessionID
	content.Type = req.Type
	content.MediaID = req.MediaID
	content.ContentText = req.ContentText

	if err := s.repo.Update(ctx, content); err != nil {
		return nil, apierror.Internal("Failed to update content")
	}

	return content, nil
}

func (s *contentService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrContentNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return apierror.Internal("Failed to delete content")
	}
	return nil
}

func (s *contentService) List(ctx context.Context, req *model.ListContentsRequest) ([]*model.Content, int64, error) {
	return s.repo.List(ctx, req)
}
