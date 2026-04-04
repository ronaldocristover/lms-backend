package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Series struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title      string    `gorm:"not null;size:255" json:"title"`
	CategoryID uuid.UUID `gorm:"type:uuid;not null;index" json:"category_id"`
	IsPaid     bool      `gorm:"not null;default:false" json:"is_paid"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (s *Series) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type CreateSeriesRequest struct {
	Title      string    `json:"title" binding:"required,max=255"`
	CategoryID uuid.UUID `json:"category_id" binding:"required"`
	IsPaid     bool      `json:"is_paid"`
}

type UpdateSeriesRequest struct {
	Title      string    `json:"title" binding:"required,max=255"`
	CategoryID uuid.UUID `json:"category_id" binding:"required"`
	IsPaid     bool      `json:"is_paid"`
}

type ListSeriesRequest struct {
	Page       int       `form:"page" binding:"omitempty,min=1"`
	PageSize   int       `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search     string    `form:"search" binding:"omitempty,max=255"`
	CategoryID uuid.UUID `form:"category_id"`
	IsPaid     *bool     `form:"is_paid"`
}
