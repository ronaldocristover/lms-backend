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

func setupContentTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE categories (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE series (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			category_id TEXT NOT NULL,
			is_paid INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			series_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			"order" INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE contents (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			type TEXT NOT NULL,
			media_id TEXT,
			content_text TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	return db
}

func createTestSessionForContent(t *testing.T, db *gorm.DB) (*model.Category, *model.Series, *model.Session) {
	t.Helper()
	category := &model.Category{
		Name: "Programming",
	}
	require.NoError(t, db.Create(category).Error)

	series := &model.Series{
		Title:      "Go Programming Bundle",
		CategoryID: category.ID,
		IsPaid:     false,
	}
	require.NoError(t, db.Create(series).Error)

	session := &model.Session{
		SeriesID:    series.ID,
		Title:       "Introduction to Go",
		Description: "Learn the basics",
		Order:       1,
	}
	require.NoError(t, db.Create(session).Error)

	return category, series, session
}

func TestContentRepository_Create(t *testing.T) {
	db := setupContentTestDB(t)
	repo := NewContentRepository(db)

	_, _, session := createTestSessionForContent(t, db)

	content := &model.Content{
		SessionID:   session.ID,
		Type:        model.ContentTypeVideo,
		ContentText: "Welcome to Go programming",
	}

	err := repo.Create(context.Background(), content)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, content.ID)
}

func TestContentRepository_GetByID(t *testing.T) {
	db := setupContentTestDB(t)
	repo := NewContentRepository(db)

	_, _, session := createTestSessionForContent(t, db)

	content := &model.Content{
		SessionID:   session.ID,
		Type:        model.ContentTypeText,
		ContentText: "This is a text lesson",
	}
	require.NoError(t, repo.Create(context.Background(), content))

	found, err := repo.GetByID(context.Background(), content.ID)
	assert.NoError(t, err)
	assert.Equal(t, content.ID, found.ID)
	assert.Equal(t, model.ContentTypeText, found.Type)
	assert.Equal(t, "This is a text lesson", found.ContentText)
}

func TestContentRepository_GetByID_NotFound(t *testing.T) {
	db := setupContentTestDB(t)
	repo := NewContentRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestContentRepository_Update(t *testing.T) {
	db := setupContentTestDB(t)
	repo := NewContentRepository(db)

	_, _, session := createTestSessionForContent(t, db)

	content := &model.Content{
		SessionID:   session.ID,
		Type:        model.ContentTypeVideo,
		ContentText: "Original text",
	}
	require.NoError(t, repo.Create(context.Background(), content))

	content.Type = model.ContentTypeText
	content.ContentText = "Updated text"
	err := repo.Update(context.Background(), content)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), content.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.ContentTypeText, found.Type)
	assert.Equal(t, "Updated text", found.ContentText)
}

func TestContentRepository_Delete(t *testing.T) {
	db := setupContentTestDB(t)
	repo := NewContentRepository(db)

	_, _, session := createTestSessionForContent(t, db)

	content := &model.Content{
		SessionID:   session.ID,
		Type:        model.ContentTypePDF,
		ContentText: "PDF content",
	}
	require.NoError(t, repo.Create(context.Background(), content))

	err := repo.Delete(context.Background(), content.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), content.ID)
	assert.Error(t, err)
}

func TestContentRepository_List(t *testing.T) {
	db := setupContentTestDB(t)
	repo := NewContentRepository(db)

	_, _, session := createTestSessionForContent(t, db)

	contents := []struct {
		contentType string
		text        string
	}{
		{model.ContentTypeVideo, "Video lesson 1"},
		{model.ContentTypeText, "Text lesson"},
		{model.ContentTypePDF, "PDF resource"},
	}
	for _, c := range contents {
		content := &model.Content{
			SessionID:   session.ID,
			Type:        c.contentType,
			ContentText: c.text,
		}
		require.NoError(t, repo.Create(context.Background(), content))
	}

	result, total, err := repo.List(context.Background(), &model.ListContentsRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 3, len(result))
}

func TestContentRepository_List_Pagination(t *testing.T) {
	db := setupContentTestDB(t)
	repo := NewContentRepository(db)

	_, _, session := createTestSessionForContent(t, db)

	contents := []struct {
		contentType string
		text        string
	}{
		{model.ContentTypeVideo, "Video lesson 1"},
		{model.ContentTypeText, "Text lesson"},
		{model.ContentTypePDF, "PDF resource"},
	}
	for _, c := range contents {
		content := &model.Content{
			SessionID:   session.ID,
			Type:        c.contentType,
			ContentText: c.text,
		}
		require.NoError(t, repo.Create(context.Background(), content))
	}

	result, total, err := repo.List(context.Background(), &model.ListContentsRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 2, len(result))
}

func TestContentRepository_List_DefaultPagination(t *testing.T) {
	db := setupContentTestDB(t)
	repo := NewContentRepository(db)

	result, _, err := repo.List(context.Background(), &model.ListContentsRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestContentRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
	db := setupContentTestDB(t)
	repo := NewContentRepository(db)

	_, _, session := createTestSessionForContent(t, db)

	for _, c := range []struct {
		contentType string
		text        string
	}{
		{model.ContentTypeVideo, "Introduction to Go"},
		{model.ContentTypeText, "Variables explained"},
		{model.ContentTypePDF, "Go cheatsheet"},
	} {
		content := &model.Content{
			SessionID:   session.ID,
			Type:        c.contentType,
			ContentText: c.text,
		}
		require.NoError(t, repo.Create(context.Background(), content))
	}

	result, total, err := repo.List(context.Background(), &model.ListContentsRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	assert.GreaterOrEqual(t, len(result), 1)
}

func TestContentRepository_List_FilterBySessionID(t *testing.T) {
	db := setupContentTestDB(t)
	repo := NewContentRepository(db)

	_, _, session1 := createTestSessionForContent(t, db)

	category2 := &model.Category{Name: "Data Science"}
	require.NoError(t, db.Create(category2).Error)

	series2 := &model.Series{
		Title:      "Python Basics",
		CategoryID: category2.ID,
		IsPaid:     false,
	}
	require.NoError(t, db.Create(series2).Error)

	session2 := &model.Session{
		SeriesID:    series2.ID,
		Title:       "Python Intro",
		Description: "Learn Python",
		Order:       1,
	}
	require.NoError(t, db.Create(session2).Error)

	for _, text := range []string{"Session 1 video", "Session 1 text"} {
		content := &model.Content{
			SessionID:   session1.ID,
			Type:        model.ContentTypeVideo,
			ContentText: text,
		}
		require.NoError(t, repo.Create(context.Background(), content))
	}

	content := &model.Content{
		SessionID:   session2.ID,
		Type:        model.ContentTypePDF,
		ContentText: "Session 2 PDF",
	}
	require.NoError(t, repo.Create(context.Background(), content))

	result, total, err := repo.List(context.Background(), &model.ListContentsRequest{
		Page:      1,
		PageSize:  20,
		SessionID: session1.ID,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, 2, len(result))
}

func TestContentRepository_List_FilterByType(t *testing.T) {
	db := setupContentTestDB(t)
	repo := NewContentRepository(db)

	_, _, session := createTestSessionForContent(t, db)

	videoContent := &model.Content{
		SessionID:   session.ID,
		Type:        model.ContentTypeVideo,
		ContentText: "Video lesson",
	}
	require.NoError(t, repo.Create(context.Background(), videoContent))

	textContent := &model.Content{
		SessionID:   session.ID,
		Type:        model.ContentTypeText,
		ContentText: "Text lesson",
	}
	require.NoError(t, repo.Create(context.Background(), textContent))

	pdfContent := &model.Content{
		SessionID:   session.ID,
		Type:        model.ContentTypePDF,
		ContentText: "PDF resource",
	}
	require.NoError(t, repo.Create(context.Background(), pdfContent))

	result, total, err := repo.List(context.Background(), &model.ListContentsRequest{
		Page:     1,
		PageSize: 20,
		Type:     model.ContentTypeVideo,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "Video lesson", result[0].ContentText)
}

func TestNewContentRepository(t *testing.T) {
	db := setupContentTestDB(t)
	repo := NewContentRepository(db)
	assert.NotNil(t, repo)
}
