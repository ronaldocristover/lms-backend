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

func setupSeriesTestDB(t *testing.T) *gorm.DB {
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
	return db
}

func createTestCategoryForSeries(t *testing.T, db *gorm.DB) *model.Category {
	t.Helper()
	category := &model.Category{
		Name: "Programming",
	}
	require.NoError(t, db.Create(category).Error)
	return category
}

func TestSeriesRepository_Create(t *testing.T) {
	db := setupSeriesTestDB(t)
	repo := NewSeriesRepository(db)

	category := createTestCategoryForSeries(t, db)

	series := &model.Series{
		Title:      "Go Programming Bundle",
		CategoryID: category.ID,
		IsPaid:     true,
	}

	err := repo.Create(context.Background(), series)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, series.ID)
}

func TestSeriesRepository_GetByID(t *testing.T) {
	db := setupSeriesTestDB(t)
	repo := NewSeriesRepository(db)

	category := createTestCategoryForSeries(t, db)

	series := &model.Series{
		Title:      "Python Basics",
		CategoryID: category.ID,
		IsPaid:     false,
	}
	require.NoError(t, repo.Create(context.Background(), series))

	found, err := repo.GetByID(context.Background(), series.ID)
	assert.NoError(t, err)
	assert.Equal(t, series.ID, found.ID)
	assert.Equal(t, "Python Basics", found.Title)
	assert.Equal(t, category.ID, found.CategoryID)
}

func TestSeriesRepository_GetByID_NotFound(t *testing.T) {
	db := setupSeriesTestDB(t)
	repo := NewSeriesRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestSeriesRepository_Update(t *testing.T) {
	db := setupSeriesTestDB(t)
	repo := NewSeriesRepository(db)

	category := createTestCategoryForSeries(t, db)

	series := &model.Series{
		Title:      "Go Programming Bundle",
		CategoryID: category.ID,
		IsPaid:     false,
	}
	require.NoError(t, repo.Create(context.Background(), series))

	series.Title = "Advanced Go Programming"
	series.IsPaid = true
	err := repo.Update(context.Background(), series)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), series.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Advanced Go Programming", found.Title)
	assert.True(t, found.IsPaid)
}

func TestSeriesRepository_Delete(t *testing.T) {
	db := setupSeriesTestDB(t)
	repo := NewSeriesRepository(db)

	category := createTestCategoryForSeries(t, db)

	series := &model.Series{
		Title:      "Python Basics",
		CategoryID: category.ID,
		IsPaid:     false,
	}
	require.NoError(t, repo.Create(context.Background(), series))

	err := repo.Delete(context.Background(), series.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), series.ID)
	assert.Error(t, err)
}

func TestSeriesRepository_List(t *testing.T) {
	db := setupSeriesTestDB(t)
	repo := NewSeriesRepository(db)

	category := createTestCategoryForSeries(t, db)

	titles := []string{"Go Programming Bundle", "Python Basics", "JavaScript Fundamentals"}
	for _, title := range titles {
		series := &model.Series{
			Title:      title,
			CategoryID: category.ID,
			IsPaid:     false,
		}
		require.NoError(t, repo.Create(context.Background(), series))
	}

	result, total, err := repo.List(context.Background(), &model.ListSeriesRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 3, len(result))
}

func TestSeriesRepository_List_Pagination(t *testing.T) {
	db := setupSeriesTestDB(t)
	repo := NewSeriesRepository(db)

	category := createTestCategoryForSeries(t, db)

	titles := []string{"Go Programming Bundle", "Python Basics", "JavaScript Fundamentals"}
	for _, title := range titles {
		series := &model.Series{
			Title:      title,
			CategoryID: category.ID,
			IsPaid:     false,
		}
		require.NoError(t, repo.Create(context.Background(), series))
	}

	result, total, err := repo.List(context.Background(), &model.ListSeriesRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 2, len(result))
}

func TestSeriesRepository_List_DefaultPagination(t *testing.T) {
	db := setupSeriesTestDB(t)
	repo := NewSeriesRepository(db)

	result, _, err := repo.List(context.Background(), &model.ListSeriesRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSeriesRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
	db := setupSeriesTestDB(t)
	repo := NewSeriesRepository(db)

	category := createTestCategoryForSeries(t, db)

	for _, title := range []string{"Go Programming Bundle", "Python Basics", "JavaScript Fundamentals"} {
		series := &model.Series{
			Title:      title,
			CategoryID: category.ID,
			IsPaid:     false,
		}
		require.NoError(t, repo.Create(context.Background(), series))
	}

	result, total, err := repo.List(context.Background(), &model.ListSeriesRequest{
		Page:     1,
		PageSize: 20,
		Search:   "go",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	assert.GreaterOrEqual(t, len(result), 1)
}

func TestSeriesRepository_List_FilterByCategoryID(t *testing.T) {
	db := setupSeriesTestDB(t)
	repo := NewSeriesRepository(db)

	category1 := &model.Category{Name: "Programming"}
	require.NoError(t, db.Create(category1).Error)

	category2 := &model.Category{Name: "Science"}
	require.NoError(t, db.Create(category2).Error)

	for _, title := range []string{"Go Programming", "Python Basics"} {
		series := &model.Series{
			Title:      title,
			CategoryID: category1.ID,
			IsPaid:     false,
		}
		require.NoError(t, repo.Create(context.Background(), series))
	}

	series := &model.Series{
		Title:      "Physics 101",
		CategoryID: category2.ID,
		IsPaid:     false,
	}
	require.NoError(t, repo.Create(context.Background(), series))

	result, total, err := repo.List(context.Background(), &model.ListSeriesRequest{
		Page:       1,
		PageSize:   20,
		CategoryID: category1.ID,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, 2, len(result))
}

func TestSeriesRepository_List_FilterByIsPaid(t *testing.T) {
	db := setupSeriesTestDB(t)
	repo := NewSeriesRepository(db)

	category := createTestCategoryForSeries(t, db)

	freeSeries := &model.Series{
		Title:      "Free Go Course",
		CategoryID: category.ID,
		IsPaid:     false,
	}
	require.NoError(t, repo.Create(context.Background(), freeSeries))

	paidSeries := &model.Series{
		Title:      "Premium Go Course",
		CategoryID: category.ID,
		IsPaid:     true,
	}
	require.NoError(t, repo.Create(context.Background(), paidSeries))

	isPaid := true
	result, total, err := repo.List(context.Background(), &model.ListSeriesRequest{
		Page:     1,
		PageSize: 20,
		IsPaid:   &isPaid,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "Premium Go Course", result[0].Title)
}

func TestNewSeriesRepository(t *testing.T) {
	db := setupSeriesTestDB(t)
	repo := NewSeriesRepository(db)
	assert.NotNil(t, repo)
}
