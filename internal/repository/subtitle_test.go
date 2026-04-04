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

func setupSubtitleTestDB(t *testing.T) *gorm.DB {
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
			created_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE subtitles (
			id TEXT PRIMARY KEY,
			media_id TEXT NOT NULL,
			language_id TEXT NOT NULL,
			content TEXT,
			created_at DATETIME
		)
	`).Error)
	return db
}

func TestSubtitleRepository_Create(t *testing.T) {
	db := setupSubtitleTestDB(t)
	repo := NewSubtitleRepository(db)

	langID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO languages (id, code, name) VALUES (?, ?, ?)", langID, "en", "English").Error)

	mediaID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO media (id, type, url) VALUES (?, ?, ?)", mediaID, "video", "http://example.com/video.mp4").Error)

	subtitle := &model.Subtitle{
		MediaID:    mediaID,
		LanguageID: langID,
		Content:    "Hello world",
	}

	err := repo.Create(context.Background(), subtitle)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, subtitle.ID)
}

func TestSubtitleRepository_GetByID(t *testing.T) {
	db := setupSubtitleTestDB(t)
	repo := NewSubtitleRepository(db)

	langID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO languages (id, code, name) VALUES (?, ?, ?)", langID, "en", "English").Error)

	mediaID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO media (id, type, url) VALUES (?, ?, ?)", mediaID, "video", "http://example.com/video.mp4").Error)

	subtitleID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO subtitles (id, media_id, language_id, content) VALUES (?, ?, ?, ?)", subtitleID, mediaID, langID, "Test content").Error)

	found, err := repo.GetByID(context.Background(), subtitleID)
	assert.NoError(t, err)
	assert.Equal(t, subtitleID, found.ID)
	assert.Equal(t, "Test content", found.Content)
}

func TestSubtitleRepository_GetByID_NotFound(t *testing.T) {
	db := setupSubtitleTestDB(t)
	repo := NewSubtitleRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestSubtitleRepository_GetByMediaAndLanguage(t *testing.T) {
	db := setupSubtitleTestDB(t)
	repo := NewSubtitleRepository(db)

	langID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO languages (id, code, name) VALUES (?, ?, ?)", langID, "en", "English").Error)

	mediaID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO media (id, type, url) VALUES (?, ?, ?)", mediaID, "video", "http://example.com/video.mp4").Error)

	subtitleID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO subtitles (id, media_id, language_id, content) VALUES (?, ?, ?, ?)", subtitleID, mediaID, langID, "Test content").Error)

	found, err := repo.GetByMediaAndLanguage(context.Background(), mediaID, langID)
	assert.NoError(t, err)
	assert.Equal(t, subtitleID, found.ID)
	assert.Equal(t, mediaID, found.MediaID)
	assert.Equal(t, langID, found.LanguageID)
}

func TestSubtitleRepository_GetByMediaAndLanguage_NotFound(t *testing.T) {
	db := setupSubtitleTestDB(t)
	repo := NewSubtitleRepository(db)

	_, err := repo.GetByMediaAndLanguage(context.Background(), uuid.New(), uuid.New())
	assert.Error(t, err)
}

func TestSubtitleRepository_Update(t *testing.T) {
	db := setupSubtitleTestDB(t)
	repo := NewSubtitleRepository(db)

	langID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO languages (id, code, name) VALUES (?, ?, ?)", langID, "en", "English").Error)

	mediaID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO media (id, type, url) VALUES (?, ?, ?)", mediaID, "video", "http://example.com/video.mp4").Error)

	subtitleID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO subtitles (id, media_id, language_id, content) VALUES (?, ?, ?, ?)", subtitleID, mediaID, langID, "Old content").Error)

	subtitle := &model.Subtitle{
		ID:         subtitleID,
		MediaID:    mediaID,
		LanguageID: langID,
		Content:    "New content",
	}

	err := repo.Update(context.Background(), subtitle)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), subtitleID)
	assert.NoError(t, err)
	assert.Equal(t, "New content", found.Content)
}

func TestSubtitleRepository_Delete(t *testing.T) {
	db := setupSubtitleTestDB(t)
	repo := NewSubtitleRepository(db)

	langID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO languages (id, code, name) VALUES (?, ?, ?)", langID, "en", "English").Error)

	mediaID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO media (id, type, url) VALUES (?, ?, ?)", mediaID, "video", "http://example.com/video.mp4").Error)

	subtitleID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO subtitles (id, media_id, language_id, content) VALUES (?, ?, ?, ?)", subtitleID, mediaID, langID, "Test content").Error)

	err := repo.Delete(context.Background(), subtitleID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), subtitleID)
	assert.Error(t, err)
}

func TestSubtitleRepository_List(t *testing.T) {
	db := setupSubtitleTestDB(t)
	repo := NewSubtitleRepository(db)

	langID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO languages (id, code, name) VALUES (?, ?, ?)", langID, "en", "English").Error)

	mediaID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO media (id, type, url) VALUES (?, ?, ?)", mediaID, "video", "http://example.com/video.mp4").Error)

	for i := 0; i < 4; i++ {
		subID := uuid.New()
		require.NoError(t, db.Exec("INSERT INTO subtitles (id, media_id, language_id, content) VALUES (?, ?, ?, ?)", subID, mediaID, langID, "Content").Error)
	}

	result, total, err := repo.List(context.Background(), &model.ListSubtitlesRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Equal(t, 4, len(result))
}

func TestSubtitleRepository_List_Pagination(t *testing.T) {
	db := setupSubtitleTestDB(t)
	repo := NewSubtitleRepository(db)

	langID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO languages (id, code, name) VALUES (?, ?, ?)", langID, "en", "English").Error)

	mediaID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO media (id, type, url) VALUES (?, ?, ?)", mediaID, "video", "http://example.com/video.mp4").Error)

	for i := 0; i < 4; i++ {
		subID := uuid.New()
		require.NoError(t, db.Exec("INSERT INTO subtitles (id, media_id, language_id, content) VALUES (?, ?, ?, ?)", subID, mediaID, langID, "Content").Error)
	}

	result, total, err := repo.List(context.Background(), &model.ListSubtitlesRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Equal(t, 2, len(result))
}

func TestSubtitleRepository_List_DefaultPagination(t *testing.T) {
	db := setupSubtitleTestDB(t)
	repo := NewSubtitleRepository(db)

	result, _, err := repo.List(context.Background(), &model.ListSubtitlesRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSubtitleRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
	db := setupSubtitleTestDB(t)
	repo := NewSubtitleRepository(db)

	langID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO languages (id, code, name) VALUES (?, ?, ?)", langID, "en", "English").Error)

	mediaID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO media (id, type, url) VALUES (?, ?, ?)", mediaID, "video", "http://example.com/video.mp4").Error)

	require.NoError(t, db.Exec("INSERT INTO subtitles (id, media_id, language_id, content) VALUES (?, ?, ?, ?)", uuid.New(), mediaID, langID, "Hello world").Error)
	require.NoError(t, db.Exec("INSERT INTO subtitles (id, media_id, language_id, content) VALUES (?, ?, ?, ?)", uuid.New(), mediaID, langID, "Goodbye world").Error)

	result, total, err := repo.List(context.Background(), &model.ListSubtitlesRequest{
		Page:     1,
		PageSize: 20,
		Search:   "hello",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	assert.GreaterOrEqual(t, len(result), 1)
}

func TestSubtitleRepository_List_WithMediaFilter(t *testing.T) {
	db := setupSubtitleTestDB(t)
	repo := NewSubtitleRepository(db)

	langID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO languages (id, code, name) VALUES (?, ?, ?)", langID, "en", "English").Error)

	mediaID1 := uuid.New()
	mediaID2 := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO media (id, type, url) VALUES (?, ?, ?)", mediaID1, "video", "http://example.com/video1.mp4").Error)
	require.NoError(t, db.Exec("INSERT INTO media (id, type, url) VALUES (?, ?, ?)", mediaID2, "video", "http://example.com/video2.mp4").Error)

	require.NoError(t, db.Exec("INSERT INTO subtitles (id, media_id, language_id, content) VALUES (?, ?, ?, ?)", uuid.New(), mediaID1, langID, "Content 1").Error)
	require.NoError(t, db.Exec("INSERT INTO subtitles (id, media_id, language_id, content) VALUES (?, ?, ?, ?)", uuid.New(), mediaID1, langID, "Content 2").Error)
	require.NoError(t, db.Exec("INSERT INTO subtitles (id, media_id, language_id, content) VALUES (?, ?, ?, ?)", uuid.New(), mediaID2, langID, "Content 3").Error)

	result, total, err := repo.List(context.Background(), &model.ListSubtitlesRequest{
		Page:     1,
		PageSize: 20,
		MediaID:  mediaID1.String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, 2, len(result))
}

func TestSubtitleRepository_List_WithLanguageFilter(t *testing.T) {
	db := setupSubtitleTestDB(t)
	repo := NewSubtitleRepository(db)

	langID1 := uuid.New()
	langID2 := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO languages (id, code, name) VALUES (?, ?, ?)", langID1, "en", "English").Error)
	require.NoError(t, db.Exec("INSERT INTO languages (id, code, name) VALUES (?, ?, ?)", langID2, "es", "Spanish").Error)

	mediaID := uuid.New()
	require.NoError(t, db.Exec("INSERT INTO media (id, type, url) VALUES (?, ?, ?)", mediaID, "video", "http://example.com/video.mp4").Error)

	require.NoError(t, db.Exec("INSERT INTO subtitles (id, media_id, language_id, content) VALUES (?, ?, ?, ?)", uuid.New(), mediaID, langID1, "English content").Error)
	require.NoError(t, db.Exec("INSERT INTO subtitles (id, media_id, language_id, content) VALUES (?, ?, ?, ?)", uuid.New(), mediaID, langID2, "Spanish content").Error)

	result, total, err := repo.List(context.Background(), &model.ListSubtitlesRequest{
		Page:       1,
		PageSize:   20,
		LanguageID: langID1.String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, 1, len(result))
}

func TestNewSubtitleRepository(t *testing.T) {
	db := setupSubtitleTestDB(t)
	repo := NewSubtitleRepository(db)
	assert.NotNil(t, repo)
}
