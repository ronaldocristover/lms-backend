package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Subtitle struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MediaID    uuid.UUID `gorm:"type:uuid;not null;index" json:"media_id"`
	LanguageID uuid.UUID `gorm:"type:uuid;not null;index" json:"language_id"`
	Content    string    `gorm:"type:text" json:"content"`
	Media      *Media    `gorm:"foreignKey:MediaID" json:"media,omitempty"`
	Language   *Language `gorm:"foreignKey:LanguageID" json:"language,omitempty"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (s *Subtitle) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type CreateSubtitleRequest struct {
	MediaID    uuid.UUID `json:"media_id" binding:"required,uuid"`
	LanguageID uuid.UUID `json:"language_id" binding:"required,uuid"`
	Content    string    `json:"content" binding:"required"`
}

type UpdateSubtitleRequest struct {
	Content string `json:"content" binding:"omitempty"`
}

type ListSubtitlesRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search     string `form:"search" binding:"omitempty,max=255"`
	MediaID    string `form:"media_id" binding:"omitempty,uuid"`
	LanguageID string `form:"language_id" binding:"omitempty,uuid"`
}
