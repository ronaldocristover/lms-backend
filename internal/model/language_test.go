package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLanguage_BeforeCreate_WithNilID(t *testing.T) {
	language := Language{
		Code: "en",
		Name: "English",
	}

	assert.Equal(t, uuid.Nil, language.ID)

	err := language.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, language.ID)
}

func TestLanguage_BeforeCreate_WithExistingID(t *testing.T) {
	existingID := uuid.New()
	language := Language{
		ID:   existingID,
		Code: "en",
		Name: "English",
	}

	err := language.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, existingID, language.ID)
}
