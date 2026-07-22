package postgres

import (
	"context"

	"github.com/reinielfc/pitchondb/apps/api/internal/pigeon"
	"github.com/reinielfc/pitchondb/apps/api/internal/platform/postgres/sqlc"
)

type pigeonQueryService struct {
	queries *sqlc.Queries
}

func NewPigeonQueryService(db sqlc.DBTX) pigeon.QueryService {
	return &pigeonQueryService{queries: sqlc.New(db)}
}

func (s *pigeonQueryService) List(ctx context.Context, f pigeon.ListFilter) ([]pigeon.Summary, int, error) {
	rows, err := s.queries.ListPigeons(ctx, sqlc.ListPigeonsParams{
		Limit:  int32(f.Limit),
		Offset: int32((f.Page - 1) * f.Limit),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := s.queries.CountPigeons(ctx)
	if err != nil {
		return nil, 0, err
	}

	var summaries []pigeon.Summary
	for _, row := range rows {
		summaries = append(summaries, toDomainPigeonSummary(row))
	}

	return summaries, int(total), nil
}

func toDomainPigeonSummary(row sqlc.PigeonSummary) pigeon.Summary {
	sex, _ := pigeon.SexString(row.Sex)
	sexConfidence, _ := pigeon.SexConfidenceStringPtr(fromNullString(row.SexConfidence))
	status, _ := pigeon.StatusString(row.Status)
	acquiredVia, _ := pigeon.AcquisitionMethodString(row.AcquiredVia)

	return pigeon.Summary{
		ID:            row.ID,
		Name:          fromNullString(row.Name),
		RingNumber:    fromNullString(row.RingNumber),
		Sex:           sex,
		SexConfidence: sexConfidence,
		Status:        status,
		AcquiredDate:  fromNullTime(row.AcquiredDate),
		AcquiredVia:   acquiredVia,
		CreatedAt:     row.CreatedAt,
	}
}
