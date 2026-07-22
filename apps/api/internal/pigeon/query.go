package pigeon

import (
	"context"
	"time"

	"github.com/google/uuid"
)

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

type ListFilter struct {
	Page  int
	Limit int
}

type QueryService interface {
	// List retrieves a list of all pigeons in the database.
	//
	// Returns a slice of [Summary] and the total count of pigeons matching the filter criteria.
	List(ctx context.Context, f ListFilter) ([]Summary, int, error)
}
