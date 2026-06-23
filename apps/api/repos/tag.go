package repos

import (
	"context"
	"database/sql"
	"fmt"
	"pitch-on-db/db"
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
	row, err := r.queries.GetPigeonTags(ctx, pigeonID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return []string{}, nil
	}
	return row, nil
}

func (r *tagRepository) SetPigeonTags(ctx context.Context, pigeonID int64, tags []string) error {
	desc := fmt.Sprintf("set tags for pigeon %d", pigeonID)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: begin transaction: %w", desc, err)
	}
	defer tx.Rollback()

	q := r.queries.WithTx(tx)

	if err = q.ClearPigeonTags(ctx, pigeonID); err != nil {
		return fmt.Errorf("%s: clear existing tags: %w", desc, err)
	}

	for _, tag := range tags {
		row, err := q.UpsertTag(ctx, tag)
		if err != nil {
			return fmt.Errorf("%s: upsert tag '%s': %w", desc, tag, err)
		}

		err = q.AddPigeonTag(ctx, db.AddPigeonTagParams{
			PigeonID: pigeonID,
			TagID:    row.ID,
		})
		if err != nil {
			return fmt.Errorf("%s: add tag '%s' to pigeon: %w", desc, tag, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit transaction: %w", desc, err)
	}
	return nil
}
