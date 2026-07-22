package pigeon

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// FindByID retrieves a pigeon by its ID.
	//
	// Returns [ErrNotFound] if no pigeon with the given ID exists.
	FindByID(ctx context.Context, id uuid.UUID) (*Pigeon, error)

	// Save creates a new pigeon record or updates an existing one in the database.
	Save(ctx context.Context, p *Pigeon) error
}
