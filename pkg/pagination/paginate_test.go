package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePage(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{0, DefaultPage},
		{-1, DefaultPage},
		{-100, DefaultPage},
		{1, 1},
		{5, 5},
		{100, 100},
	}
	for _, tt := range tests {
		result := NormalizePage(tt.input)
		assert.Equal(t, tt.expected, result, "NormalizePage(%d)", tt.input)
	}
}

func TestNormalizePageSize(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{0, DefaultPageSize},
		{-1, DefaultPageSize},
		{101, DefaultPageSize},
		{200, DefaultPageSize},
		{1, 1},
		{20, 20},
		{50, 50},
		{100, 100},
	}
	for _, tt := range tests {
		result := NormalizePageSize(tt.input)
		assert.Equal(t, tt.expected, result, "NormalizePageSize(%d)", tt.input)
	}
}

func TestOffset(t *testing.T) {
	tests := []struct {
		page     int
		pageSize int
		expected int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 20, 40},
		{1, 10, 0},
		{5, 10, 40},
		{0, 20, 0},    // normalizes page to 1
		{1, 0, 0},     // normalizes pageSize to 20
		{2, 0, 20},    // normalizes pageSize to 20, page 2 → offset 20
	}
	for _, tt := range tests {
		result := Offset(tt.page, tt.pageSize)
		assert.Equal(t, tt.expected, result, "Offset(%d, %d)", tt.page, tt.pageSize)
	}
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 1, DefaultPage)
	assert.Equal(t, 20, DefaultPageSize)
	assert.Equal(t, 100, MaxPageSize)
}
