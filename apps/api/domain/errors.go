package domain

import (
	"fmt"
	"log/slog"
)

type DomainError interface {
	error
	slog.LogValuer
	Code() ErrorCode
}

type ErrorCode string

func (c ErrorCode) Error() string { return string(c) }

const (
	ErrInternal     ErrorCode = "INTERNAL"
	ErrNotFound     ErrorCode = "NOT_FOUND"
	ErrConflict     ErrorCode = "CONFLICT"
	ErrInvalid      ErrorCode = "INVALID"
	ErrUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrForbidden    ErrorCode = "FORBIDDEN"
)

type ErrorOptions func(*domainError)

func applyOpts(err *domainError, opts ...ErrorOptions) DomainError {
	for _, opt := range opts {
		opt(err)
	}
	return err
}

func WithPfx(format string, a ...any) ErrorOptions {
	return func(e *domainError) {
		e.Public = fmt.Sprintf(format, a...) + ": " + e.Public
	}
}

func WithMsg(format string, a ...any) ErrorOptions {
	return func(e *domainError) {
		e.Public = fmt.Sprintf(format, a...)
	}
}

func WithFmt(format string, a ...any) ErrorOptions {
	return func(e *domainError) {
		if e.Internal == nil {
			e.Internal = fmt.Errorf(format, a...)
		} else {
			e.Internal = fmt.Errorf(format+": %w", append(a, e.Internal)...)
		}
	}
}

func WithCause(cause error) ErrorOptions {
	return func(e *domainError) {
		e.Internal = cause
	}
}

func WithCtx(key string, value any) ErrorOptions {
	return func(e *domainError) {
		if e.Context == nil {
			e.Context = make(map[string]any)
		}
		e.Context[key] = value
	}
}

type domainError struct {
	code     ErrorCode
	Public   string
	Internal error
	Context  map[string]any
}

func (e *domainError) Code() ErrorCode { return e.code }
func (e *domainError) Error() string   { return e.Public }
func (e *domainError) Unwrap() error   { return e.Internal }

func (e *domainError) Is(target error) bool {
	if e == nil {
		return false
	}

	if code, ok := target.(ErrorCode); ok {
		return e.code == code
	}

	err, ok := target.(interface{ Code() ErrorCode })
	return ok && e.code == err.Code()
}

func (e *domainError) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("code", string(e.code)),
		slog.String("msg", e.Public),
		slog.Any("cause", e.Internal),
		slog.Any("context", e.Context),
	)
}

func NewDomainError(code ErrorCode, msg string, opts ...ErrorOptions) DomainError {
	err := &domainError{code: code, Public: msg}
	return applyOpts(err, opts...)
}

func NewInternalError(opts ...ErrorOptions) DomainError {
	return NewDomainError(ErrInternal, "internal error", opts...)
}

func NewValidationError(opts ...ErrorOptions) DomainError {
	return NewDomainError(ErrInvalid, "validation error", opts...)
}

func NewNotFoundError(opts ...ErrorOptions) DomainError {
	return NewDomainError(ErrNotFound, "resource not found", opts...)
}

func NewResourceNotFoundError(resource string, id any, opts ...ErrorOptions) DomainError {
	allOpts := append([]ErrorOptions{
		WithMsg("%s not found: %v", resource, id),
		WithCtx("resource", resource),
		WithCtx("id", id),
	}, opts...)
	return NewNotFoundError(allOpts...)
}

func WrapError(err error, opts ...ErrorOptions) DomainError {
	if err == nil {
		return nil
	}

	if de, ok := err.(*domainError); ok {
		return applyOpts(de, opts...)
	}

	return NewInternalError(WithCause(err))
}

func Errorf(format string, a ...any) DomainError {
	for _, arg := range a {
		if err, ok := arg.(error); ok {
			return WrapError(err, WithFmt(format, a...))
		}
	}
	return NewInternalError(WithFmt(format, a...))
}
