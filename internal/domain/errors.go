package domain

import "fmt"

type ErrorKind string

const (
	KindNotFound     ErrorKind = "Not Found"
	KindConflict     ErrorKind = "Conflict"
	KindInvalid      ErrorKind = "Invalid"
	KindUnauthorized ErrorKind = "Unauthorized"
	KindForbidden    ErrorKind = "Forbidden"
)

type DomainError struct {
	Kind    ErrorKind
	Message string
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func ErrNotFound(format string, a ...any) *DomainError {
	return &DomainError{
		Kind:    KindNotFound,
		Message: fmt.Sprintf(format, a...),
	}
}

func ErrConflict(format string, a ...any) *DomainError {
	return &DomainError{
		Kind:    KindConflict,
		Message: fmt.Sprintf(format, a...),
	}
}

func ErrInvalid(format string, a ...any) *DomainError {
	return &DomainError{
		Kind:    KindInvalid,
		Message: fmt.Sprintf(format, a...),
	}
}

func ErrUnauthorized() *DomainError {
	return &DomainError{
		Kind:    KindUnauthorized,
		Message: "authentication required",
	}
}

func ErrForbidden() *DomainError {
	return &DomainError{
		Kind:    KindForbidden,
		Message: "access denied",
	}
}
