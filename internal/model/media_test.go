package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMedia_BeforeCreate_NilID(t *testing.T) {
	media := &Media{
		Type: MediaTypeVideo,
		URL:  "https://example.com/video.mp4",
	}

	assert.Equal(t, uuid.Nil, media.ID)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = media.BeforeCreate(db)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, media.ID)
}

func TestMedia_BeforeCreate_ExistingID(t *testing.T) {
	existingID := uuid.New()
	media := &Media{
		ID:   existingID,
		Type: MediaTypeAudio,
		URL:  "https://example.com/audio.mp3",
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = media.BeforeCreate(db)
	assert.NoError(t, err)
	assert.Equal(t, existingID, media.ID)
}

func TestMediaTypeConstants(t *testing.T) {
	assert.Equal(t, "video", MediaTypeVideo)
	assert.Equal(t, "audio", MediaTypeAudio)
	assert.Equal(t, "pdf", MediaTypePDF)
	assert.Equal(t, "image", MediaTypeImage)
}
