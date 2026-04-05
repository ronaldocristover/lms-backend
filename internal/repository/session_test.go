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

func setupSessionTestDB(t *testing.T) *gorm.DB {
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
	return db
}

func createTestSeriesForSession(t *testing.T, db *gorm.DB) (*model.Category, *model.Series) {
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

	return category, series
}

func TestSessionRepository_Create(t *testing.T) {
	db := setupSessionTestDB(t)
	repo := NewSessionRepository(db)

	_, series := createTestSeriesForSession(t, db)

	session := &model.Session{
		SeriesID:    series.ID,
		Title:       "Introduction to Go",
		Description: "Learn the basics of Go programming",
		Order:       1,
	}

	err := repo.Create(context.Background(), session)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, session.ID)
}

func TestSessionRepository_GetByID(t *testing.T) {
	db := setupSessionTestDB(t)
	repo := NewSessionRepository(db)

	_, series := createTestSeriesForSession(t, db)

	session := &model.Session{
		SeriesID:    series.ID,
		Title:       "Variables and Types",
		Description: "Understanding Go variables and types",
		Order:       2,
	}
	require.NoError(t, repo.Create(context.Background(), session))

	found, err := repo.GetByID(context.Background(), session.ID)
	assert.NoError(t, err)
	assert.Equal(t, session.ID, found.ID)
	assert.Equal(t, "Variables and Types", found.Title)
	assert.Equal(t, series.ID, found.SeriesID)
}

func TestSessionRepository_GetByID_NotFound(t *testing.T) {
	db := setupSessionTestDB(t)
	repo := NewSessionRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestSessionRepository_Update(t *testing.T) {
	db := setupSessionTestDB(t)
	repo := NewSessionRepository(db)

	_, series := createTestSeriesForSession(t, db)

	session := &model.Session{
		SeriesID:    series.ID,
		Title:       "Introduction to Go",
		Description: "Learn the basics",
		Order:       1,
	}
	require.NoError(t, repo.Create(context.Background(), session))

	session.Title = "Advanced Introduction to Go"
	session.Description = "Learn the basics and more"
	session.Order = 2
	err := repo.Update(context.Background(), session)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), session.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Advanced Introduction to Go", found.Title)
	assert.Equal(t, "Learn the basics and more", found.Description)
	assert.Equal(t, 2, found.Order)
}

func TestSessionRepository_Delete(t *testing.T) {
	db := setupSessionTestDB(t)
	repo := NewSessionRepository(db)

	_, series := createTestSeriesForSession(t, db)

	session := &model.Session{
		SeriesID:    series.ID,
		Title:       "Introduction to Go",
		Description: "Learn the basics",
		Order:       1,
	}
	require.NoError(t, repo.Create(context.Background(), session))

	err := repo.Delete(context.Background(), session.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), session.ID)
	assert.Error(t, err)
}

func TestSessionRepository_List(t *testing.T) {
	db := setupSessionTestDB(t)
	repo := NewSessionRepository(db)

	_, series := createTestSeriesForSession(t, db)

	titles := []string{"Introduction to Go", "Variables and Types", "Functions and Methods"}
	for i, title := range titles {
		session := &model.Session{
			SeriesID:    series.ID,
			Title:       title,
			Description: "Session description",
			Order:       i + 1,
		}
		require.NoError(t, repo.Create(context.Background(), session))
	}

	result, total, err := repo.List(context.Background(), &model.ListSessionsRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 3, len(result))
}

func TestSessionRepository_List_Pagination(t *testing.T) {
	db := setupSessionTestDB(t)
	repo := NewSessionRepository(db)

	_, series := createTestSeriesForSession(t, db)

	titles := []string{"Introduction to Go", "Variables and Types", "Functions and Methods"}
	for i, title := range titles {
		session := &model.Session{
			SeriesID:    series.ID,
			Title:       title,
			Description: "Session description",
			Order:       i + 1,
		}
		require.NoError(t, repo.Create(context.Background(), session))
	}

	result, total, err := repo.List(context.Background(), &model.ListSessionsRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 2, len(result))
}

func TestSessionRepository_List_DefaultPagination(t *testing.T) {
	db := setupSessionTestDB(t)
	repo := NewSessionRepository(db)

	result, _, err := repo.List(context.Background(), &model.ListSessionsRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSessionRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
	db := setupSessionTestDB(t)
	repo := NewSessionRepository(db)

	_, series := createTestSeriesForSession(t, db)

	for i, title := range []string{"Introduction to Go", "Variables and Types", "Functions and Methods"} {
		session := &model.Session{
			SeriesID:    series.ID,
			Title:       title,
			Description: "Session description",
			Order:       i + 1,
		}
		require.NoError(t, repo.Create(context.Background(), session))
	}

	result, total, err := repo.List(context.Background(), &model.ListSessionsRequest{
		Page:     1,
		PageSize: 20,
		Search:   "go",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	assert.GreaterOrEqual(t, len(result), 1)
}

func TestSessionRepository_List_FilterBySeriesID(t *testing.T) {
	db := setupSessionTestDB(t)
	repo := NewSessionRepository(db)

	_, series1 := createTestSeriesForSession(t, db)

	category := &model.Category{Name: "Science"}
	require.NoError(t, db.Create(category).Error)

	series2 := &model.Series{
		Title:      "Physics 101",
		CategoryID: category.ID,
		IsPaid:     false,
	}
	require.NoError(t, db.Create(series2).Error)

	for i, title := range []string{"Introduction to Go", "Variables and Types"} {
		session := &model.Session{
			SeriesID:    series1.ID,
			Title:       title,
			Description: "Go session",
			Order:       i + 1,
		}
		require.NoError(t, repo.Create(context.Background(), session))
	}

	session := &model.Session{
		SeriesID:    series2.ID,
		Title:       "Newton's Laws",
		Description: "Physics session",
		Order:       1,
	}
	require.NoError(t, repo.Create(context.Background(), session))

	result, total, err := repo.List(context.Background(), &model.ListSessionsRequest{
		Page:     1,
		PageSize: 20,
		SeriesID: series1.ID,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, 2, len(result))
}

func TestNewSessionRepository(t *testing.T) {
	db := setupSessionTestDB(t)
	repo := NewSessionRepository(db)
	assert.NotNil(t, repo)
}
