package repos

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/reinielfc/pitch-on-db/apps/api/db"
)

// TagRepository defines the data access contract for tag records.
type TagRepository interface {
	// List returns all existing tag names.
	List(ctx context.Context) ([]string, error)

	// PruneOrphanedTags removes tags that are not assigned to any pigeon.
	PruneOrphanedTags(ctx context.Context) error

	// GetPigeonTags returns the tag names assigned to the given pigeon.
	GetPigeonTags(ctx context.Context, pigeonID int64) ([]string, error)

	// SetPigeonTags replaces the full set of tags for the given pigeon with the provided tag names.
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
	tags, err := r.queries.ListTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	return tags, nil
}

func (r *tagRepository) PruneOrphanedTags(ctx context.Context) error {
	if err := r.queries.PruneOrphanedTags(ctx); err != nil {
		return fmt.Errorf("prune orphaned tags: %w", err)
	}
	return nil
}

func (r *tagRepository) GetPigeonTags(ctx context.Context, pigeonID int64) ([]string, error) {
	row, err := r.queries.GetPigeonTags(ctx, pigeonID)
	if err != nil {
		return nil, fmt.Errorf("get tags for pigeon %d: %w", pigeonID, err)
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

	if err = q.RemoveAllPigeonTags(ctx, pigeonID); err != nil {
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
