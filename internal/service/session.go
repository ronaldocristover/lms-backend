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
	ErrSessionNotFound = errors.New("session not found")
)

type SessionService interface {
	Create(ctx context.Context, req *model.CreateSessionRequest) (*model.Session, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Session, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateSessionRequest) (*model.Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *model.ListSessionsRequest) ([]*model.Session, int64, error)
}

type sessionService struct {
	repo       repository.SessionRepository
	seriesRepo repository.SeriesRepository
	logger *zap.SugaredLogger
}

func NewSessionService(repo repository.SessionRepository, seriesRepo repository.SeriesRepository, logger *zap.SugaredLogger) SessionService {
	return &sessionService{
		repo:       repo,
		seriesRepo: seriesRepo,
		logger:     logger,
	}
}

func (s *sessionService) Create(ctx context.Context, req *model.CreateSessionRequest) (*model.Session, error) {
	if _, err := s.seriesRepo.GetByID(ctx, req.SeriesID); err != nil {
		return nil, ErrSeriesNotFound
	}

	session := &model.Session{
		SeriesID:    req.SeriesID,
		Title:       req.Title,
		Description: req.Description,
		Order:       req.Order,
	}

	if err := s.repo.Create(ctx, session); err != nil {
		s.logger.Errorw("operation failed", "error", err)
		return nil, apierror.Internal("Failed to create session")
	}

	s.logger.Infow("session created", "id", session.ID)
	return session, nil
}

func (s *sessionService) GetByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (s *sessionService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateSessionRequest) (*model.Session, error) {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if _, err := s.seriesRepo.GetByID(ctx, req.SeriesID); err != nil {
		return nil, ErrSeriesNotFound
	}

	session.SeriesID = req.SeriesID
	session.Title = req.Title
	session.Description = req.Description
	session.Order = req.Order

	if err := s.repo.Update(ctx, session); err != nil {
		s.logger.Errorw("operation failed", "error", err)
		return nil, apierror.Internal("Failed to update session")
	}

	return session, nil
}

func (s *sessionService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrSessionNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Errorw("operation failed", "error", err)
		return apierror.Internal("Failed to delete session")
	}
	return nil
}

func (s *sessionService) List(ctx context.Context, req *model.ListSessionsRequest) ([]*model.Session, int64, error) {
	return s.repo.List(ctx, req)
}
