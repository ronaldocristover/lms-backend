package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Media struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Type       string     `gorm:"not null;size:20;index" json:"type"`
	URL        string     `gorm:"not null" json:"url"`
	LanguageID *uuid.UUID `gorm:"type:uuid;index" json:"language_id"`
	Language   *Language  `gorm:"foreignKey:LanguageID" json:"language,omitempty"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (m *Media) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

const (
	MediaTypeVideo = "video"
	MediaTypeAudio = "audio"
	MediaTypePDF   = "pdf"
	MediaTypeImage = "image"
)

type CreateMediaRequest struct {
	Type       string     `json:"type" binding:"required,oneof=video audio pdf image"`
	URL        string     `json:"url" binding:"required,url"`
	LanguageID *uuid.UUID `json:"language_id"`
}

type UpdateMediaRequest struct {
	Type       string     `json:"type" binding:"omitempty,oneof=video audio pdf image"`
	URL        string     `json:"url" binding:"omitempty,url"`
	LanguageID *uuid.UUID `json:"language_id"`
}

type ListMediaRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search     string `form:"search" binding:"omitempty,max=255"`
	Type       string `form:"type" binding:"omitempty,oneof=video audio pdf image"`
	LanguageID string `form:"language_id" binding:"omitempty,uuid"`
}
