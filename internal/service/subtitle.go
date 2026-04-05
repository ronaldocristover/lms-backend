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
	ErrSubtitleNotFound = errors.New("subtitle not found")
	ErrSubtitleExists   = errors.New("subtitle already exists for this media and language")
)

type SubtitleService interface {
	Create(ctx context.Context, req *model.CreateSubtitleRequest) (*model.Subtitle, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subtitle, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateSubtitleRequest) (*model.Subtitle, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *model.ListSubtitlesRequest) ([]*model.Subtitle, int64, error)
}

type subtitleService struct {
	repo      repository.SubtitleRepository
	mediaRepo repository.MediaRepository
	langRepo  repository.LanguageRepository
	logger *zap.SugaredLogger
}

func NewSubtitleService(repo repository.SubtitleRepository, mediaRepo repository.MediaRepository, langRepo repository.LanguageRepository, logger *zap.SugaredLogger) SubtitleService {
	return &subtitleService{repo: repo, mediaRepo: mediaRepo, langRepo: langRepo, logger: logger}
}

func (s *subtitleService) Create(ctx context.Context, req *model.CreateSubtitleRequest) (*model.Subtitle, error) {
	if _, err := s.mediaRepo.GetByID(ctx, req.MediaID); err != nil {
		return nil, apierror.BadRequest("Invalid media ID")
	}

	if _, err := s.langRepo.GetByID(ctx, req.LanguageID); err != nil {
		return nil, apierror.BadRequest("Invalid language ID")
	}

	existing, _ := s.repo.GetByMediaAndLanguage(ctx, req.MediaID, req.LanguageID)
	if existing != nil {
		return nil, ErrSubtitleExists
	}

	subtitle := &model.Subtitle{
		MediaID:    req.MediaID,
		LanguageID: req.LanguageID,
		Content:    req.Content,
	}

	if err := s.repo.Create(ctx, subtitle); err != nil {
		s.logger.Errorw("operation failed", "error", err)
		return nil, apierror.Internal("Failed to create subtitle")
	}

	s.logger.Infow("subtitle created", "id", subtitle.ID)
	return subtitle, nil
}

func (s *subtitleService) GetByID(ctx context.Context, id uuid.UUID) (*model.Subtitle, error) {
	subtitle, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrSubtitleNotFound
	}
	return subtitle, nil
}

func (s *subtitleService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateSubtitleRequest) (*model.Subtitle, error) {
	subtitle, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrSubtitleNotFound
	}

	if req.Content != "" {
		subtitle.Content = req.Content
	}

	if err := s.repo.Update(ctx, subtitle); err != nil {
		s.logger.Errorw("operation failed", "error", err)
		return nil, apierror.Internal("Failed to update subtitle")
	}

	return subtitle, nil
}

func (s *subtitleService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrSubtitleNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Errorw("operation failed", "error", err)
		return apierror.Internal("Failed to delete subtitle")
	}
	return nil
}

func (s *subtitleService) List(ctx context.Context, req *model.ListSubtitlesRequest) ([]*model.Subtitle, int64, error) {
	return s.repo.List(ctx, req)
}
