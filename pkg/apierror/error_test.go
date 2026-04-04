package apierror

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	err := &Error{
		Code:    http.StatusBadRequest,
		Message: "test error",
	}

	assert.Equal(t, "test error", err.Error())
}

func TestBadRequest(t *testing.T) {
	err := BadRequest("invalid input")

	assert.Equal(t, http.StatusBadRequest, err.Code)
	assert.Equal(t, "invalid input", err.Message)
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized("not authorized")

	assert.Equal(t, http.StatusUnauthorized, err.Code)
	assert.Equal(t, "not authorized", err.Message)
}

func TestNotFound(t *testing.T) {
	err := NotFound("resource not found")

	assert.Equal(t, http.StatusNotFound, err.Code)
	assert.Equal(t, "resource not found", err.Message)
}

func TestConflict(t *testing.T) {
	err := Conflict("resource already exists")

	assert.Equal(t, http.StatusConflict, err.Code)
	assert.Equal(t, "resource already exists", err.Message)
}

func TestInternal(t *testing.T) {
	err := Internal("something went wrong")

	assert.Equal(t, http.StatusInternalServerError, err.Code)
	assert.Equal(t, "something went wrong", err.Message)
}

func TestTooManyRequests(t *testing.T) {
	err := TooManyRequests("rate limit exceeded")

	assert.Equal(t, http.StatusTooManyRequests, err.Code)
	assert.Equal(t, "rate limit exceeded", err.Message)
}

func TestFromError(t *testing.T) {
	apiErr := &Error{Code: http.StatusBadRequest, Message: "test"}
	result := FromError(apiErr)

	assert.Equal(t, apiErr, result)

	stdErr := assert.AnError
	result = FromError(stdErr)

	assert.Equal(t, http.StatusInternalServerError, result.Code)
	assert.Equal(t, stdErr.Error(), result.Message)
}

func TestWrap(t *testing.T) {
	innerErr := assert.AnError
	wrapped := Wrap(innerErr, "outer error")

	assert.Contains(t, wrapped.Message, "outer error")
	assert.Contains(t, wrapped.Message, innerErr.Error())
}
