// Package domain defines the core types, domain errors, and error codes shared
// across the application. It contains the [Pigeon] model, supporting value types
// such as [Sex], patch types, and the [DomainError] interface with its concrete
// error constructors (e.g. [NewResourceNotFoundError], [NewValidationError]).
package domain
