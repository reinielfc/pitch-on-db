package repos

import (
	"context"
	"database/sql"
	"fmt"
	"pitch-on-db/internal/db"
)

type TagRepository interface {
	List(ctx context.Context) ([]string, error)
	ClearUnusedTags(ctx context.Context) error

	GetPigeonTags(ctx context.Context, pigeonID int64) ([]string, error)
	SetPigeonTags(ctx context.Context, pigeonID int64, tags []string) error
}

type tagRepository struct {
	db      *sql.DB
	queries *db.Queries
}

func NewTagRepository(sqlDB *sql.DB) TagRepository {
	return &tagRepository{db: sqlDB, queries: db.New(sqlDB)}
}

func (r *tagRepository) List(ctx context.Context) ([]string, error) {
	return r.queries.ListTags(ctx)
}

func (r *tagRepository) ClearUnusedTags(ctx context.Context) error {
	return r.queries.ClearUnusedTags(ctx)
}

func (r *tagRepository) GetPigeonTags(ctx context.Context, pigeonID int64) ([]string, error) {
	return r.queries.GetPigeonTags(ctx, pigeonID)
}

func (r *tagRepository) SetPigeonTags(ctx context.Context, pigeonID int64, tags []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := db.New(tx)

	if err = q.ClearPigeonTags(ctx, pigeonID); err != nil {
		return fmt.Errorf("clear tags for pigeon %d: %w", pigeonID, err)
	}

	for _, tag := range tags {
		row, err := q.UpsertTag(ctx, tag)
		if err != nil {
			return err
		}

		err = q.AddPigeonTag(ctx, db.AddPigeonTagParams{
			PigeonID: pigeonID,
			TagID:    row.ID,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
