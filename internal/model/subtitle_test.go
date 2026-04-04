package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSubtitle_BeforeCreate_WithNilID(t *testing.T) {
	subtitle := Subtitle{
		MediaID:    uuid.New(),
		LanguageID: uuid.New(),
		Content:    "Test content",
	}

	assert.Equal(t, uuid.Nil, subtitle.ID)

	err := subtitle.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, subtitle.ID)
}

func TestSubtitle_BeforeCreate_WithExistingID(t *testing.T) {
	existingID := uuid.New()
	subtitle := Subtitle{
		ID:         existingID,
		MediaID:    uuid.New(),
		LanguageID: uuid.New(),
		Content:    "Test content",
	}

	err := subtitle.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, existingID, subtitle.ID)
}
