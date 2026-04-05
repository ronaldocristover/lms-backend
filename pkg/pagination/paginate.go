package pagination

import "gorm.io/gorm"

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// NormalizePage returns a valid page number (minimum 1).
func NormalizePage(page int) int {
	if page < 1 {
		return DefaultPage
	}
	return page
}

// NormalizePageSize returns a valid page size (1-100, default 20).
func NormalizePageSize(pageSize int) int {
	if pageSize < 1 || pageSize > MaxPageSize {
		return DefaultPageSize
	}
	return pageSize
}

// Offset calculates the SQL offset from page and page size.
func Offset(page, pageSize int) int {
	return (NormalizePage(page) - 1) * NormalizePageSize(pageSize)
}

// Paginate applies Count + Limit/Offset to a GORM query.
// Returns total count and the paginated query ready for Find().
func Paginate(query *gorm.DB, page, pageSize int) (total int64, paginated *gorm.DB, err error) {
	page = NormalizePage(page)
	pageSize = NormalizePageSize(pageSize)

	if err := query.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	paginated = query.Limit(pageSize).Offset(Offset(page, pageSize))
	return total, paginated, nil
}
