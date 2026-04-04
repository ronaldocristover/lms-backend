package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMeta(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		pageSize       int
		total          int64
		expectedPages  int
	}{
		{
			name:          "exact pages",
			page:          1,
			pageSize:      10,
			total:         100,
			expectedPages: 10,
		},
		{
			name:          "partial last page",
			page:          1,
			pageSize:      10,
			total:         95,
			expectedPages: 10,
		},
		{
			name:          "single item",
			page:          1,
			pageSize:      10,
			total:         1,
			expectedPages: 1,
		},
		{
			name:          "zero items",
			page:          1,
			pageSize:      10,
			total:         0,
			expectedPages: 0,
		},
		{
			name:          "large dataset",
			page:          1,
			pageSize:      25,
			total:         1000,
			expectedPages: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := NewMeta(tt.page, tt.pageSize, tt.total)

			assert.Equal(t, tt.page, meta.Page)
			assert.Equal(t, tt.pageSize, meta.PageSize)
			assert.Equal(t, tt.total, meta.TotalItems)
			assert.Equal(t, tt.expectedPages, meta.TotalPages)
		})
	}
}

func TestNewMeta_TotalPagesCalculation(t *testing.T) {
	// Test that TotalPages rounds up correctly
	meta := NewMeta(1, 10, 11)
	assert.Equal(t, 2, meta.TotalPages)

	meta = NewMeta(1, 10, 20)
	assert.Equal(t, 2, meta.TotalPages)

	meta = NewMeta(1, 10, 21)
	assert.Equal(t, 3, meta.TotalPages)
}
