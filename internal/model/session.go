package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SeriesID    uuid.UUID `gorm:"type:uuid;not null;index" json:"series_id"`
	Title       string    `gorm:"not null;size:255" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Order       int       `gorm:"not null;default:0" json:"order"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Series *Series `gorm:"foreignKey:SeriesID" json:"series,omitempty"`
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type CreateSessionRequest struct {
	SeriesID    uuid.UUID `json:"series_id" binding:"required"`
	Title       string    `json:"title" binding:"required,max=255"`
	Description string    `json:"description"`
	Order       int       `json:"order"`
}

type UpdateSessionRequest struct {
	SeriesID    uuid.UUID `json:"series_id" binding:"required"`
	Title       string    `json:"title" binding:"required,max=255"`
	Description string    `json:"description"`
	Order       int       `json:"order"`
}

type ListSessionsRequest struct {
	Page     int       `form:"page" binding:"omitempty,min=1"`
	PageSize int       `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search   string    `form:"search" binding:"omitempty,max=255"`
	SeriesID uuid.UUID `form:"series_id"`
}
