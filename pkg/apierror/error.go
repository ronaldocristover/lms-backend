package apierror

import (
	"fmt"
	"net/http"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

func BadRequest(message string) *Error {
	return &Error{Code: http.StatusBadRequest, Message: message}
}

func Unauthorized(message string) *Error {
	return &Error{Code: http.StatusUnauthorized, Message: message}
}

func Forbidden(message string) *Error {
	return &Error{Code: http.StatusForbidden, Message: message}
}

func NotFound(message string) *Error {
	return &Error{Code: http.StatusNotFound, Message: message}
}

func Conflict(message string) *Error {
	return &Error{Code: http.StatusConflict, Message: message}
}

func Internal(message string) *Error {
	return &Error{Code: http.StatusInternalServerError, Message: message}
}

func UnprocessableEntity(message string) *Error {
	return &Error{Code: http.StatusUnprocessableEntity, Message: message}
}

func TooManyRequests(message string) *Error {
	return &Error{Code: http.StatusTooManyRequests, Message: message}
}

func New(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

func FromError(err error) *Error {
	if e, ok := err.(*Error); ok {
		return e
	}
	return Internal(err.Error())
}

func Wrap(err error, message string) *Error {
	return &Error{
		Code:    http.StatusInternalServerError,
		Message: fmt.Sprintf("%s: %s", message, err.Error()),
	}
}
