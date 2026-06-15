package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"pitch-on-db/internal/db"
	"pitch-on-db/internal/domain"
)

type PigeonRepository interface {
	List(ctx context.Context) ([]domain.Pigeon, error)

	Create(ctx context.Context, pigeon domain.Pigeon) (domain.Pigeon, error)
	Get(ctx context.Context, id int64) (*domain.Pigeon, error)
	Exists(ctx context.Context, id int64) (bool, error)
	Update(ctx context.Context, id int64, pigeonPatch domain.PigeonPatch) (domain.Pigeon, error)
	Delete(ctx context.Context, id int64) error
}

type pigeonRepository struct {
	db      *sql.DB
	queries *db.Queries
}

func NewPigeonRepository(sqlDB *sql.DB) PigeonRepository {
	return &pigeonRepository{db: sqlDB, queries: db.New(sqlDB)}
}

func (r *pigeonRepository) List(ctx context.Context) ([]domain.Pigeon, error) {
	rows, err := r.queries.ListPigeons(ctx)
	if err != nil {
		return nil, err
	}

	pigeons := make([]domain.Pigeon, len(rows))
	for i, row := range rows {
		pigeons[i] = toDomain(row)
	}
	return pigeons, nil
}

func (r *pigeonRepository) Create(ctx context.Context, newPigeon domain.Pigeon) (domain.Pigeon, error) {
	row, err := r.queries.CreatePigeon(ctx, db.CreatePigeonParams{
		Name:       newPigeon.Name,
		BandNumber: newPigeon.BandNumber,
		BirthDate:  newPigeon.BirthDate,
		Sex:        newPigeon.Sex,
	})
	if err != nil {
		return domain.Pigeon{}, err
	}
	return toDomain(row), nil
}

func (r *pigeonRepository) Get(ctx context.Context, id int64) (*domain.Pigeon, error) {
	row, err := r.queries.GetPigeon(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	pigeon := toDomain(row)
	return &pigeon, nil
}

func (r *pigeonRepository) Exists(ctx context.Context, id int64) (bool, error) {
	exists, err := r.queries.PigeonExists(ctx, id)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *pigeonRepository) Update(ctx context.Context, id int64, pigeonPatch domain.PigeonPatch) (domain.Pigeon, error) {
	row, err := r.queries.UpdatePigeon(ctx, db.UpdatePigeonParams{
		ID:   id,
		Name: pigeonPatch.Name,

		SetBandNumber: pigeonPatch.BandNumber != nil,
		BandNumber:    pigeonPatch.BandNumber,

		SetBirthDate: pigeonPatch.BirthDate != nil,
		BirthDate:    pigeonPatch.BirthDate,

		SetSex: pigeonPatch.Sex != nil,
		Sex:    pigeonPatch.Sex,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Pigeon{}, domain.ErrNotFound("pigeon by id %d", id)
	}
	if err != nil {
		return domain.Pigeon{}, fmt.Errorf("update pigeon %d: %w", id, err)
	}

	return toDomain(row), nil
}

func (r *pigeonRepository) Delete(ctx context.Context, id int64) error {
	return r.queries.DeletePigeon(ctx, id)
}

func toDomain(row db.Pigeon) domain.Pigeon {
	return domain.Pigeon{
		ID:         row.ID,
		Name:       row.Name,
		CreatedAt:  row.CreatedAt,
		BandNumber: row.BandNumber,
		BirthDate:  row.BirthDate,
		Sex:        row.Sex,
	}
}
