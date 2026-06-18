package domain

import (
	"errors"
	"fmt"
)

type ErrorKind int

const (
	KindNotFound ErrorKind = iota
	KindConflict
	KindInvalid
	KindUnauthorized
	KindForbidden
)

func (k ErrorKind) String() string {
	switch k {
	case KindNotFound:
		return "Not Found"
	case KindConflict:
		return "Conflict"
	case KindInvalid:
		return "Invalid"
	case KindUnauthorized:
		return "Unauthorized"
	case KindForbidden:
		return "Forbidden"
	default:
		return "Unknown"
	}
}

type DomainError struct {
	Kind    ErrorKind
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func (e *DomainError) Unwrap() error { return e.Cause }

func ErrNotFound(format string, a ...any) *DomainError { return errWrap(KindNotFound, format, a...) }
func ErrConflict(format string, a ...any) *DomainError { return errWrap(KindConflict, format, a...) }
func ErrInvalid(format string, a ...any) *DomainError  { return errWrap(KindInvalid, format, a...) }
func ErrUnauthorized() *DomainError                    { return errWrap(KindUnauthorized, "authentication required") }
func ErrForbidden() *DomainError                       { return errWrap(KindForbidden, "access denied") }

func errWrap(kind ErrorKind, format string, a ...any) *DomainError {
	wrapped := fmt.Errorf(format, a...)
	return &DomainError{
		Kind:    kind,
		Message: wrapped.Error(),
		Cause:   errors.Unwrap(wrapped),
	}
}

func IsNotFound(err error) bool     { return isKind(err, KindNotFound) }
func IsConflict(err error) bool     { return isKind(err, KindConflict) }
func IsInvalid(err error) bool      { return isKind(err, KindInvalid) }
func IsUnauthorized(err error) bool { return isKind(err, KindUnauthorized) }
func IsForbidden(err error) bool    { return isKind(err, KindForbidden) }

func isKind(err error, kind ErrorKind) bool {
	var domErr *DomainError
	return errors.As(err, &domErr) && domErr.Kind == kind
}
