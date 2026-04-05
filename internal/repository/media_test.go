package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/ronaldocristover/lms-backend/internal/model"
)

func setupMediaTestDB(t *testing.T) *gorm.DB {
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
	require.NoError(t, db.Exec(`
		CREATE TABLE media (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			url TEXT NOT NULL,
			language_id TEXT,
			created_at DATETIME,
			FOREIGN KEY (language_id) REFERENCES languages(id)
		)
	`).Error)
	return db
}

func TestMediaRepository_Create(t *testing.T) {
	db := setupMediaTestDB(t)
	repo := NewMediaRepository(db)

	media := &model.Media{
		Type: model.MediaTypeVideo,
		URL:  "https://example.com/video.mp4",
	}

	err := repo.Create(context.Background(), media)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, media.ID)
}

func TestMediaRepository_GetByID(t *testing.T) {
	db := setupMediaTestDB(t)
	repo := NewMediaRepository(db)

	media := &model.Media{
		Type: model.MediaTypeAudio,
		URL:  "https://example.com/audio.mp3",
	}
	require.NoError(t, repo.Create(context.Background(), media))

	found, err := repo.GetByID(context.Background(), media.ID)
	assert.NoError(t, err)
	assert.Equal(t, media.ID, found.ID)
	assert.Equal(t, model.MediaTypeAudio, found.Type)
	assert.Equal(t, "https://example.com/audio.mp3", found.URL)
}

func TestMediaRepository_GetByID_NotFound(t *testing.T) {
	db := setupMediaTestDB(t)
	repo := NewMediaRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestMediaRepository_Update(t *testing.T) {
	db := setupMediaTestDB(t)
	repo := NewMediaRepository(db)

	media := &model.Media{
		Type: model.MediaTypeVideo,
		URL:  "https://example.com/old.mp4",
	}
	require.NoError(t, repo.Create(context.Background(), media))

	media.URL = "https://example.com/new.mp4"
	err := repo.Update(context.Background(), media)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), media.ID)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/new.mp4", found.URL)
}

func TestMediaRepository_Delete(t *testing.T) {
	db := setupMediaTestDB(t)
	repo := NewMediaRepository(db)

	media := &model.Media{
		Type: model.MediaTypePDF,
		URL:  "https://example.com/doc.pdf",
	}
	require.NoError(t, repo.Create(context.Background(), media))

	err := repo.Delete(context.Background(), media.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), media.ID)
	assert.Error(t, err)
}

func TestMediaRepository_List(t *testing.T) {
	db := setupMediaTestDB(t)
	repo := NewMediaRepository(db)

	items := []struct {
		Type string
		URL  string
	}{
		{model.MediaTypeVideo, "https://example.com/video1.mp4"},
		{model.MediaTypeAudio, "https://example.com/audio1.mp3"},
		{model.MediaTypePDF, "https://example.com/doc1.pdf"},
		{model.MediaTypeImage, "https://example.com/img1.png"},
	}
	for _, item := range items {
		media := &model.Media{Type: item.Type, URL: item.URL}
		require.NoError(t, repo.Create(context.Background(), media))
	}

	result, total, err := repo.List(context.Background(), &model.ListMediaRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Equal(t, 4, len(result))
}

func TestMediaRepository_List_Pagination(t *testing.T) {
	db := setupMediaTestDB(t)
	repo := NewMediaRepository(db)

	items := []struct {
		Type string
		URL  string
	}{
		{model.MediaTypeVideo, "https://example.com/video1.mp4"},
		{model.MediaTypeAudio, "https://example.com/audio1.mp3"},
		{model.MediaTypePDF, "https://example.com/doc1.pdf"},
		{model.MediaTypeImage, "https://example.com/img1.png"},
	}
	for _, item := range items {
		media := &model.Media{Type: item.Type, URL: item.URL}
		require.NoError(t, repo.Create(context.Background(), media))
	}

	result, total, err := repo.List(context.Background(), &model.ListMediaRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Equal(t, 2, len(result))
}

func TestMediaRepository_List_DefaultPagination(t *testing.T) {
	db := setupMediaTestDB(t)
	repo := NewMediaRepository(db)

	result, _, err := repo.List(context.Background(), &model.ListMediaRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestMediaRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
	db := setupMediaTestDB(t)
	repo := NewMediaRepository(db)

	for _, item := range []struct {
		Type string
		URL  string
	}{
		{model.MediaTypeVideo, "https://example.com/video.mp4"},
		{model.MediaTypeAudio, "https://example.com/audio.mp3"},
	} {
		media := &model.Media{Type: item.Type, URL: item.URL}
		require.NoError(t, repo.Create(context.Background(), media))
	}

	result, total, err := repo.List(context.Background(), &model.ListMediaRequest{
		Page:     1,
		PageSize: 20,
		Search:   "video",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	assert.GreaterOrEqual(t, len(result), 1)
}

func TestMediaRepository_List_WithType(t *testing.T) {
	db := setupMediaTestDB(t)
	repo := NewMediaRepository(db)

	items := []struct {
		Type string
		URL  string
	}{
		{model.MediaTypeVideo, "https://example.com/video1.mp4"},
		{model.MediaTypeVideo, "https://example.com/video2.mp4"},
		{model.MediaTypeAudio, "https://example.com/audio1.mp3"},
	}
	for _, item := range items {
		media := &model.Media{Type: item.Type, URL: item.URL}
		require.NoError(t, repo.Create(context.Background(), media))
	}

	result, total, err := repo.List(context.Background(), &model.ListMediaRequest{
		Page:     1,
		PageSize: 20,
		Type:     model.MediaTypeVideo,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, 2, len(result))
	for _, m := range result {
		assert.Equal(t, model.MediaTypeVideo, m.Type)
	}
}

func TestNewMediaRepository(t *testing.T) {
	db := setupMediaTestDB(t)
	repo := NewMediaRepository(db)
	assert.NotNil(t, repo)
}
