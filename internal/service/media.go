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
	ErrMediaNotFound = errors.New("media not found")
)

type MediaService interface {
	Create(ctx context.Context, req *model.CreateMediaRequest) (*model.Media, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Media, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateMediaRequest) (*model.Media, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *model.ListMediaRequest) ([]*model.Media, int64, error)
}

type mediaService struct {
	repo     repository.MediaRepository
	langRepo repository.LanguageRepository
	logger *zap.SugaredLogger
}

func NewMediaService(repo repository.MediaRepository, langRepo repository.LanguageRepository, logger *zap.SugaredLogger) MediaService {
	return &mediaService{repo: repo, langRepo: langRepo, logger: logger}
}

func (s *mediaService) Create(ctx context.Context, req *model.CreateMediaRequest) (*model.Media, error) {
	if req.LanguageID != nil {
		if _, err := s.langRepo.GetByID(ctx, *req.LanguageID); err != nil {
			return nil, apierror.BadRequest("Invalid language ID")
		}
	}

	media := &model.Media{
		Type:       req.Type,
		URL:        req.URL,
		LanguageID: req.LanguageID,
	}

	if err := s.repo.Create(ctx, media); err != nil {
		s.logger.Errorw("operation failed", "error", err)
		return nil, apierror.Internal("Failed to create media")
	}

	s.logger.Infow("media created", "id", media.ID)
	return media, nil
}

func (s *mediaService) GetByID(ctx context.Context, id uuid.UUID) (*model.Media, error) {
	media, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrMediaNotFound
	}
	return media, nil
}

func (s *mediaService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateMediaRequest) (*model.Media, error) {
	media, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrMediaNotFound
	}

	if req.Type != "" {
		media.Type = req.Type
	}

	if req.URL != "" {
		media.URL = req.URL
	}

	if req.LanguageID != nil {
		if _, err := s.langRepo.GetByID(ctx, *req.LanguageID); err != nil {
			return nil, apierror.BadRequest("Invalid language ID")
		}
		media.LanguageID = req.LanguageID
	}

	if err := s.repo.Update(ctx, media); err != nil {
		s.logger.Errorw("operation failed", "error", err)
		return nil, apierror.Internal("Failed to update media")
	}

	return media, nil
}

func (s *mediaService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrMediaNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Errorw("operation failed", "error", err)
		return apierror.Internal("Failed to delete media")
	}
	return nil
}

func (s *mediaService) List(ctx context.Context, req *model.ListMediaRequest) ([]*model.Media, int64, error) {
	return s.repo.List(ctx, req)
}
