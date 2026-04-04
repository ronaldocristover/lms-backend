package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Language struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code      string    `gorm:"uniqueIndex;not null;size:10" json:"code"`
	Name      string    `gorm:"not null;size:100" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (l *Language) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

type CreateLanguageRequest struct {
	Code string `json:"code" binding:"required,max=10"`
	Name string `json:"name" binding:"required,max=100"`
}

type UpdateLanguageRequest struct {
	Code string `json:"code" binding:"omitempty,max=10"`
	Name string `json:"name" binding:"omitempty,max=100"`
}

type ListLanguagesRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search   string `form:"search" binding:"omitempty,max=255"`
}
