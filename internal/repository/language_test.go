package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/yourusername/lms/internal/model"
)

func setupLanguageTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE languages (
			id TEXT PRIMARY KEY,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			created_at DATETIME
		)
	`).Error)
	return db
}

func TestLanguageRepository_Create(t *testing.T) {
	db := setupLanguageTestDB(t)
	repo := NewLanguageRepository(db)

	language := &model.Language{
		Code: "en",
		Name: "English",
	}

	err := repo.Create(context.Background(), language)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, language.ID)
}

func TestLanguageRepository_GetByID(t *testing.T) {
	db := setupLanguageTestDB(t)
	repo := NewLanguageRepository(db)

	language := &model.Language{
		Code: "en",
		Name: "English",
	}
	require.NoError(t, repo.Create(context.Background(), language))

	found, err := repo.GetByID(context.Background(), language.ID)
	assert.NoError(t, err)
	assert.Equal(t, language.ID, found.ID)
	assert.Equal(t, "en", found.Code)
	assert.Equal(t, "English", found.Name)
}

func TestLanguageRepository_GetByID_NotFound(t *testing.T) {
	db := setupLanguageTestDB(t)
	repo := NewLanguageRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestLanguageRepository_GetByCode(t *testing.T) {
	db := setupLanguageTestDB(t)
	repo := NewLanguageRepository(db)

	language := &model.Language{
		Code: "en",
		Name: "English",
	}
	require.NoError(t, repo.Create(context.Background(), language))

	found, err := repo.GetByCode(context.Background(), "en")
	assert.NoError(t, err)
	assert.Equal(t, language.ID, found.ID)
	assert.Equal(t, "en", found.Code)
	assert.Equal(t, "English", found.Name)
}

func TestLanguageRepository_GetByCode_NotFound(t *testing.T) {
	db := setupLanguageTestDB(t)
	repo := NewLanguageRepository(db)

	_, err := repo.GetByCode(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestLanguageRepository_Update(t *testing.T) {
	db := setupLanguageTestDB(t)
	repo := NewLanguageRepository(db)

	language := &model.Language{
		Code: "en",
		Name: "English",
	}
	require.NoError(t, repo.Create(context.Background(), language))

	language.Name = "English (US)"
	err := repo.Update(context.Background(), language)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), language.ID)
	assert.NoError(t, err)
	assert.Equal(t, "English (US)", found.Name)
}

func TestLanguageRepository_Delete(t *testing.T) {
	db := setupLanguageTestDB(t)
	repo := NewLanguageRepository(db)

	language := &model.Language{
		Code: "en",
		Name: "English",
	}
	require.NoError(t, repo.Create(context.Background(), language))

	err := repo.Delete(context.Background(), language.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), language.ID)
	assert.Error(t, err)
}

func TestLanguageRepository_List(t *testing.T) {
	db := setupLanguageTestDB(t)
	repo := NewLanguageRepository(db)

	languages := []struct {
		Code string
		Name string
	}{
		{"en", "English"},
		{"es", "Spanish"},
		{"fr", "French"},
		{"de", "German"},
	}
	for _, l := range languages {
		language := &model.Language{Code: l.Code, Name: l.Name}
		require.NoError(t, repo.Create(context.Background(), language))
	}

	result, total, err := repo.List(context.Background(), &model.ListLanguagesRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Equal(t, 4, len(result))
}

func TestLanguageRepository_List_Pagination(t *testing.T) {
	db := setupLanguageTestDB(t)
	repo := NewLanguageRepository(db)

	languages := []struct {
		Code string
		Name string
	}{
		{"en", "English"},
		{"es", "Spanish"},
		{"fr", "French"},
		{"de", "German"},
	}
	for _, l := range languages {
		language := &model.Language{Code: l.Code, Name: l.Name}
		require.NoError(t, repo.Create(context.Background(), language))
	}

	result, total, err := repo.List(context.Background(), &model.ListLanguagesRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Equal(t, 2, len(result))
}

func TestLanguageRepository_List_DefaultPagination(t *testing.T) {
	db := setupLanguageTestDB(t)
	repo := NewLanguageRepository(db)

	result, _, err := repo.List(context.Background(), &model.ListLanguagesRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestLanguageRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
	db := setupLanguageTestDB(t)
	repo := NewLanguageRepository(db)

	for _, l := range []struct{ Code, Name string }{
		{"en", "English"},
		{"es", "Spanish"},
		{"fr", "French"},
	} {
		language := &model.Language{Code: l.Code, Name: l.Name}
		require.NoError(t, repo.Create(context.Background(), language))
	}

	result, total, err := repo.List(context.Background(), &model.ListLanguagesRequest{
		Page:     1,
		PageSize: 20,
		Search:   "eng",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	assert.GreaterOrEqual(t, len(result), 1)
}

func TestNewLanguageRepository(t *testing.T) {
	db := setupLanguageTestDB(t)
	repo := NewLanguageRepository(db)
	assert.NotNil(t, repo)
}
