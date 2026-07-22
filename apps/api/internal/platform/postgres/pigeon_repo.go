package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/reinielfc/pitchondb/apps/api/internal/pigeon"
	"github.com/reinielfc/pitchondb/apps/api/internal/platform/postgres/sqlc"
)

type pigeonRepository struct {
	queries *sqlc.Queries
}

func NewPigeonRepository(db sqlc.DBTX) pigeon.Repository {
	return &pigeonRepository{queries: sqlc.New(db)}
}

func (r *pigeonRepository) FindByID(ctx context.Context, id uuid.UUID) (*pigeon.Pigeon, error) {
	row, err := r.queries.FindPigeonByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pigeon.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return toDomainPigeon(row), nil
}

func (r *pigeonRepository) Save(ctx context.Context, p *pigeon.Pigeon) error {
	row := fromDomainPigeon(p)
	_, err := r.queries.UpsertPigeon(ctx, sqlc.UpsertPigeonParams{
		ID:            row.ID,
		Name:          row.Name,
		RingNumber:    row.RingNumber,
		Sex:           row.Sex,
		SexConfidence: row.SexConfidence,
		Status:        row.Status,
		AcquiredDate:  row.AcquiredDate,
		AcquiredVia:   row.AcquiredVia,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	})
	if err != nil {
		return err
	}
	return nil
}

func fromDomainPigeon(p *pigeon.Pigeon) sqlc.Pigeon {
	s := p.Snapshot()
	return sqlc.Pigeon{
		ID:            s.ID,
		Name:          toNullString(s.Name),
		RingNumber:    toNullString(s.RingNumber),
		Sex:           s.Sex.String(),
		SexConfidence: toNullString(s.SexConfidence.StringPtr()),
		Status:        s.Status.String(),
		AcquiredDate:  toNullTime(s.AcquiredDate),
		AcquiredVia:   s.AcquiredVia.String(),
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

func toDomainPigeon(row sqlc.Pigeon) *pigeon.Pigeon {
	sex, _ := pigeon.SexString(row.Sex)
	sexConfidence, _ := pigeon.SexConfidenceStringPtr(fromNullString(row.SexConfidence))
	status, _ := pigeon.StatusString(row.Status)
	acquiredVia, _ := pigeon.AcquisitionMethodString(row.AcquiredVia)

	return pigeon.Reconstitute(
		row.ID,
		fromNullString(row.Name),
		fromNullString(row.RingNumber),
		sex,
		sexConfidence,
		status,
		fromNullTime(row.AcquiredDate),
		acquiredVia,
		row.CreatedAt,
		row.UpdatedAt,
	)
}
