package domain

import (
	"fmt"
	"log/slog"
)

// DomainError represents a structured domain error that includes error code and
// a client-facing error message. It implements both the standard error interface
// and slog.LogValuer for structured logging.
type DomainError interface {
	error
	slog.LogValuer

	// Code returns the error code associated with the domain error.
	Code() ErrorCode

	// Message returns the client-facing error message.
	Message() string
}

// ErrorCode is a string-based error code representing a specific error category.
// It implements the error interface for use in error comparisons.
type ErrorCode string

// Error implements the error interface for ErrorCode.
func (c ErrorCode) Error() string { return string(c) }

const (
	ErrInternal     ErrorCode = "INTERNAL"
	ErrNotFound     ErrorCode = "NOT_FOUND"
	ErrConflict     ErrorCode = "CONFLICT"
	ErrInvalid      ErrorCode = "INVALID"
	ErrUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrForbidden    ErrorCode = "FORBIDDEN"
)

// ErrorOptions is a functional option type used to configure domain errors.
// It allows optional customization of error messages, causes, and context.
type ErrorOptions func(*domainError)

// applyOpts applies a sequence of ErrorOptions to a domain error.
func applyOpts(err *domainError, opts ...ErrorOptions) DomainError {
	for _, opt := range opts {
		opt(err)
	}
	return err
}

// ErrOpts groups multiple ErrorOptions into a single ErrorOptions.
func ErrOpts(opts ...ErrorOptions) ErrorOptions {
	return func(e *domainError) {
		for _, opt := range opts {
			opt(e)
		}
	}
}

// WithMsg sets the client-facing error message using a format string.
func WithMsg(format string, a ...any) ErrorOptions {
	return func(e *domainError) {
		e.message = fmt.Sprintf(format, a...)
	}
}

// WithErr sets the internal error cause for debugging purposes.
func WithErr(format string, a ...any) ErrorOptions {
	return func(e *domainError) {
		e.cause = fmt.Errorf(format, a...)
	}
}

// WithCtx adds a key-value pair to the error's context map for structured logging and error details.
func WithCtx(key string, value any) ErrorOptions {
	return func(e *domainError) {
		if e.context == nil {
			e.context = make(map[string]any)
		}
		e.context[key] = value
	}
}

type domainError struct {
	code    ErrorCode
	message string
	context map[string]any
	cause   error
}

func (e *domainError) Code() ErrorCode { return e.code }
func (e *domainError) Message() string { return e.message }
func (e *domainError) Error() string {
	if e.cause == nil {
		return e.message
	}
	return e.cause.Error()
}
func (e *domainError) Unwrap() error { return e.cause }

func (e *domainError) Is(target error) bool {
	if e == nil {
		return false
	}

	if code, ok := target.(ErrorCode); ok {
		return e.code == code
	}

	err, ok := target.(DomainError)
	return ok && e.code == err.Code()
}

func (e *domainError) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("code", string(e.code)),
		slog.String("msg", e.message),
		slog.Any("cause", e.cause),
		slog.Any("context", e.context),
	)
}

func newDomainError(code ErrorCode, msg string, opts ...ErrorOptions) DomainError {
	err := &domainError{code: code, message: msg}
	return applyOpts(err, opts...)
}

// NewInternalError creates a CodeInternal error for unexpected server-side issues.
func NewInternalError(opts ...ErrorOptions) DomainError {
	return newDomainError(ErrInternal, "internal error", opts...)
}

// NewValidationError creates a CodeInvalid error for input validation failures.
func NewValidationError(opts ...ErrorOptions) DomainError {
	return newDomainError(ErrInvalid, "validation error", opts...)
}

// NewNotFoundError creates a CodeNotFound error for missing resources.
func NewNotFoundError(opts ...ErrorOptions) DomainError {
	return newDomainError(ErrNotFound, "resource not found", opts...)
}

// NewResourceNotFoundError is a convenience constructor for resource-specific not-found errors.
// It automatically includes the resource type and ID in the context.
func NewResourceNotFoundError(resource string, id any, opts ...ErrorOptions) DomainError {
	allOpts := append([]ErrorOptions{
		WithMsg("%s with ID %v not found", resource, id),
		WithCtx("resource", resource),
		WithCtx("id", id),
	}, opts...)
	return NewNotFoundError(allOpts...)
}
