package pigeon

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type QueryService interface {
	// List retrieves a list of all pigeons in the database.
	//
	// Returns a slice of [Summary] and the total count of pigeons matching the filter criteria.
	List(ctx context.Context, f ListFilter) ([]Summary, int, error)
}

// ListFilter represents the filtering options for listing pigeons, including pagination parameters.
type ListFilter struct {
	Page  int
	Limit int
}

// Summary represents a summary of a pigeon entity, including its ID, name, ring number, etc.
// This struct is used for listing pigeons with essential information.
type Summary struct {
	ID            uuid.UUID
	Name          *string
	RingNumber    *string
	Sex           Sex
	SexConfidence *SexConfidence
	Status        Status
	AcquiredDate  *time.Time
	AcquiredVia   AcquisitionMethod
	CreatedAt     time.Time
}
