package handlers

import (
	"fmt"
	"net/http"
)

type HttpError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func (e *HttpError) Error() string {
	return e.Message
}

func NewApiError(status int, format string, args ...any) *HttpError {
	return &HttpError{
		Status:  status,
		Code:    http.StatusText(status),
		Message: fmt.Sprintf(format, args...),
	}
}

func ErrWrap(status int, err error, format string, args ...any) *HttpError {
	return &HttpError{
		Status:  status,
		Code:    http.StatusText(status),
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
	}
}

func ErrBadRequest(format string, args ...any) *HttpError {
	return NewApiError(http.StatusBadRequest, format, args...)
}

func ErrNotFound(format string, args ...any) *HttpError {
	return NewApiError(http.StatusNotFound, format, args...)
}

func ErrInternal(format string, args ...any) *HttpError {
	return NewApiError(http.StatusInternalServerError, format, args...)
}
