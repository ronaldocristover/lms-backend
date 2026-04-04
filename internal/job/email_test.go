package job

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmailJob(t *testing.T) {
	job := NewEmailJob("test@example.com", "Test Subject", "Test Body")

	assert.Equal(t, "test@example.com", job.To)
	assert.Equal(t, "Test Subject", job.Subject)
	assert.Equal(t, "Test Body", job.Body)
}

func TestEmailJob_Run(t *testing.T) {
	job := NewEmailJob("test@example.com", "Test Subject", "Test Body")

	err := job.Run()
	assert.NoError(t, err)
}

func TestEmailJob_EmptyFields(t *testing.T) {
	job := NewEmailJob("", "", "")

	assert.Equal(t, "", job.To)
	assert.Equal(t, "", job.Subject)
	assert.Equal(t, "", job.Body)

	err := job.Run()
	assert.NoError(t, err)
}
