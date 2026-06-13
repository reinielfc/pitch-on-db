package errors

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Status  int
	Code    string
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(status int, message string) *AppError {
	return &AppError{
		Status:  status,
		Code:    http.StatusText(status),
		Message: message,
	}
}

func Wrap(err error, status int, message string) *AppError {
	return &AppError{
		Status:  status,
		Code:    http.StatusText(status),
		Message: message,
		Cause:   err,
	}
}

func NotFound(res string) *AppError {
	return NewAppError(http.StatusNotFound, fmt.Sprintf("%s not found", res))
}

func Conflict(msg string) *AppError {
	return NewAppError(http.StatusConflict, msg)
}

func Unauthorized() *AppError {
	return NewAppError(http.StatusUnauthorized, "authentication required")
}

func BadRequest(msg string) *AppError {
	return NewAppError(http.StatusBadRequest, msg)
}

func Internal(cause error, msg string) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Code:    http.StatusText(http.StatusInternalServerError),
		Message: msg,
		Cause:   cause,
	}
}

func DBResource(resource string, err error) *AppError {
	if errors.Is(err, sql.ErrNoRows) {
		return NotFound(resource)
	}
	e := Internal(err, "database error")
	return e
}
