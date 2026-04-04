package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Content struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionID   uuid.UUID `gorm:"type:uuid;not null;index" json:"session_id"`
	Type        string    `gorm:"not null;size:50" json:"type"`
	MediaID     uuid.UUID `gorm:"type:uuid" json:"media_id"`
	ContentText string    `gorm:"type:text" json:"content_text"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Session *Session `gorm:"foreignKey:SessionID" json:"session,omitempty"`
}

func (c *Content) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

const (
	ContentTypeVideo = "video"
	ContentTypeAudio = "audio"
	ContentTypeText  = "text"
	ContentTypePDF   = "pdf"
)

type CreateContentRequest struct {
	SessionID   uuid.UUID `json:"session_id" binding:"required"`
	Type        string    `json:"type" binding:"required,oneof=video audio text pdf"`
	MediaID     uuid.UUID `json:"media_id"`
	ContentText string    `json:"content_text"`
}

type UpdateContentRequest struct {
	SessionID   uuid.UUID `json:"session_id" binding:"required"`
	Type        string    `json:"type" binding:"required,oneof=video audio text pdf"`
	MediaID     uuid.UUID `json:"media_id"`
	ContentText string    `json:"content_text"`
}

type ListContentsRequest struct {
	Page      int       `form:"page" binding:"omitempty,min=1"`
	PageSize  int       `form:"page_size" binding:"omitempty,min=1,max=100"`
	SessionID uuid.UUID `form:"session_id"`
	Type      string    `form:"type" binding:"omitempty,oneof=video audio text pdf"`
}
