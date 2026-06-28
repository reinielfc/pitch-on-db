package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/reinielfc/pitch-on-db/apps/api/domain"
	"github.com/reinielfc/pitch-on-db/apps/api/repos"
)

// PigeonService defines the business logic contract for pigeon operations.
type PigeonService interface {
	// Create persists a new pigeon record and returns the created pigeon.
	Create(ctx context.Context, pigeon domain.Pigeon) (domain.Pigeon, error)

	// Get returns the pigeon with the given id.
	// Returns [domain.ResourceNotFoundError] if the pigeon does not exist.
	Get(ctx context.Context, id int64) (*domain.Pigeon, error)

	// GetTags returns the tag names assigned to the given pigeon.
	// Returns [domain.ResourceNotFoundError] if the pigeon does not exist.
	GetTags(ctx context.Context, pigeonID int64) ([]string, error)

	// GetParents returns the assigned parents of the given child pigeon.
	// Returns [domain.ResourceNotFoundError] if the pigeon does not exist.
	GetParents(ctx context.Context, childID int64) (*domain.PigeonParents, error)

	// GetChildren returns all children of the given parent pigeon.
	// Errors:
	//   - [domain.ResourceNotFoundError] if the pigeon does not exist.
	//   - [domain.ValidationError] if the pigeon's sex is unset.
	GetChildren(ctx context.Context, parentID int64) ([]domain.Pigeon, error)

	// List returns all pigeon records.
	List(ctx context.Context) ([]domain.Pigeon, error)

	// Update applies the given patch to the pigeon with the given id and returns the updated pigeon.
	// Errors:
	//   - [domain.ResourceNotFoundError] if the pigeon does not exist.
	//   - [domain.ValidationError] if patching sex and the pigeon already has children.
	Update(ctx context.Context, id int64, pigeonPatch domain.PigeonPatch) (domain.Pigeon, error)

	// SetTags replaces the full set of tags for the given pigeon with the provided tag names.
	// Returns [domain.ResourceNotFoundError] if the pigeon does not exist.
	SetTags(ctx context.Context, pigeonID int64, tagNames []string) error

	// AssignChild assigns the child pigeon to the given parent pigeon.
	// Errors:
	//   - [domain.ResourceNotFoundError] if either pigeon does not exist.
	//   - [domain.ValidationError] if parentID == childID or the parent's sex is unset.
	AssignChild(ctx context.Context, parentID int64, childID int64) error

	// Delete removes the pigeon with the given id.
	Delete(ctx context.Context, id int64) error
}

type pigeonService struct {
	repo    repos.PigeonRepository
	tagRepo repos.TagRepository
}

func NewPigeonService(r repos.PigeonRepository, t repos.TagRepository) PigeonService {
	return &pigeonService{repo: r, tagRepo: t}
}

func (s *pigeonService) Create(ctx context.Context, pigeon domain.Pigeon) (domain.Pigeon, error) {
	createdPigeon, err := s.repo.Create(ctx, pigeon)
	if err != nil {
		return domain.Pigeon{}, domain.NewInternalError(
			domain.WithErr("create pigeon: %w", err),
			domain.WithCtx("pigeon", pigeon),
		)
	}
	return createdPigeon, nil
}

func (s *pigeonService) List(ctx context.Context) ([]domain.Pigeon, error) {
	pigeons, err := s.repo.List(ctx)
	if err != nil {
		return nil, domain.NewInternalError(
			domain.WithErr("list pigeons: %w", err),
		)
	}
	return pigeons, nil
}

func (s *pigeonService) Get(ctx context.Context, id int64) (*domain.Pigeon, error) {
	p, err := s.repo.Get(ctx, id)
	if errors.Is(err, repos.ErrNotFound) {
		return nil, domain.NewResourceNotFoundError("pigeon", id,
			domain.WithErr("get pigeon: %w", err))
	}
	if err != nil {
		return nil, domain.NewInternalError(
			domain.WithErr("get pigeon: %w", err),
			domain.WithCtx("pigeonID", id),
		)
	}
	if p == nil {
		return nil, domain.NewResourceNotFoundError("pigeon", id,
			domain.WithErr("get pigeon"))
	}
	return p, nil
}

func (s *pigeonService) Update(ctx context.Context, id int64, patch domain.PigeonPatch) (domain.Pigeon, error) {
	if patch.Sex != nil {
		if err := s.ensureHasNoChildren(ctx, id); err != nil {
			return domain.Pigeon{}, fmt.Errorf("update pigeon: %w", err)
		}
	}
	updatedPigeon, err := s.repo.Update(ctx, id, patch)
	if errors.Is(err, repos.ErrNotFound) {
		return domain.Pigeon{}, domain.NewResourceNotFoundError("pigeon", id,
			domain.WithErr("update pigeon: %w", err))
	}
	if err != nil {
		return domain.Pigeon{}, domain.NewInternalError(
			domain.WithErr("update pigeon: %w", err),
			domain.WithCtx("pigeonID", id),
		)
	}
	return updatedPigeon, nil
}

func (s *pigeonService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return domain.NewInternalError(
			domain.WithErr("delete pigeon: %w", err),
			domain.WithCtx("pigeonID", id),
		)
	}
	return nil
}

func (s *pigeonService) AssignChild(ctx context.Context, parentID int64, childID int64) error {
	if parentID == childID {
		return domain.NewValidationError(
			domain.WithMsg("cannot assign pigeon %d as its own parent", parentID),
			domain.WithErr("assign child: cannot assign pigeon as its own parent"),
			domain.WithCtx("parentID", parentID),
			domain.WithCtx("childID", childID),
		)
	}

	if err := s.ensureExists(ctx, childID); err != nil {
		return fmt.Errorf("assign child to pigeon (child): %w", err)
	}

	if err := s.ensureExists(ctx, parentID); err != nil {
		return fmt.Errorf("assign child to pigeon (parent): %w", err)
	}

	if err := s.repo.AssignChild(ctx, parentID, childID); errors.Is(err, repos.ErrPigeonSexUnset) {
		return domain.NewValidationError(
			domain.WithMsg("cannot assign child %d to parent %d: pigeon sex is unset", childID, parentID),
			domain.WithErr("assign child to pigeon: %w", err),
			domain.WithCtx("parentID", parentID),
			domain.WithCtx("childID", childID),
		)
	} else if err != nil {
		return domain.NewInternalError(
			domain.WithErr("assign child to pigeon: %w", err),
			domain.WithCtx("parentID", parentID),
			domain.WithCtx("childID", childID),
		)
	}
	return nil
}

func (s *pigeonService) GetParents(ctx context.Context, childID int64) (*domain.PigeonParents, error) {
	if err := s.ensureExists(ctx, childID); err != nil {
		return nil, fmt.Errorf("get parents for pigeon: %w", err)
	}
	parents, err := s.repo.GetParents(ctx, childID)
	if err != nil {
		return nil, domain.NewInternalError(
			domain.WithErr("get parents for pigeon: %w", err),
			domain.WithCtx("childID", childID),
		)
	}
	return parents, nil
}

func (s *pigeonService) GetChildren(ctx context.Context, parentID int64) ([]domain.Pigeon, error) {
	if err := s.ensureExists(ctx, parentID); err != nil {
		return nil, fmt.Errorf("get children for pigeon: %w", err)
	}

	children, err := s.repo.GetChildren(ctx, parentID)
	if errors.Is(err, repos.ErrPigeonSexUnset) {
		return nil, domain.NewValidationError(
			domain.WithMsg("cannot get children for pigeon %d: pigeon sex is unset", parentID),
			domain.WithErr("get children for pigeon: %w", err),
			domain.WithCtx("parentID", parentID),
		)
	}
	if err != nil {
		return nil, domain.NewInternalError(
			domain.WithErr("get children for pigeon: %w", err),
			domain.WithCtx("parentID", parentID),
		)
	}
	return children, nil
}

func (s *pigeonService) GetTags(ctx context.Context, id int64) ([]string, error) {
	if err := s.ensureExists(ctx, id); err != nil {
		return nil, fmt.Errorf("get tags for pigeon: %w", err)
	}
	tags, err := s.tagRepo.GetPigeonTags(ctx, id)
	if err != nil {
		return nil, domain.NewInternalError(
			domain.WithErr("get tags for pigeon: %w", err),
			domain.WithCtx("pigeonID", id),
		)
	}
	return tags, nil
}

func (s *pigeonService) SetTags(ctx context.Context, id int64, tagNames []string) error {
	if err := s.ensureExists(ctx, id); err != nil {
		return fmt.Errorf("set tags for pigeon: %w", err)
	}
	if err := s.tagRepo.SetPigeonTags(ctx, id, tagNames); err != nil {
		return domain.NewInternalError(
			domain.WithErr("set tags for pigeon: %w", err),
			domain.WithCtx("pigeonID", id),
		)
	}
	return nil
}

func (s *pigeonService) ensureExists(ctx context.Context, id int64) domain.DomainError {
	if exists, err := s.repo.Exists(ctx, id); err != nil {
		return domain.NewInternalError(
			domain.WithErr("ensure exists: %w", err),
			domain.WithCtx("pigeonID", id),
		)
	} else if !exists {
		return domain.NewResourceNotFoundError("pigeon", id)
	}
	return nil
}

func (s *pigeonService) ensureHasNoChildren(ctx context.Context, id int64) error {
	hasChildren, err := s.repo.HasChildren(ctx, id)
	if errors.Is(err, repos.ErrNotFound) {
		return domain.NewResourceNotFoundError("pigeon", id,
			domain.WithErr("ensure has no children: %w", err))
	}
	if errors.Is(err, repos.ErrPigeonSexUnset) {
		return nil // sex unset means AssignChild could never have succeeded; no children possible
	}
	if err != nil {
		return domain.NewInternalError(
			domain.WithErr("ensure has no children: %w", err),
			domain.WithCtx("pigeonID", id))
	}
	if hasChildren {
		return domain.NewValidationError(
			domain.WithMsg("pigeon has children: %d", id),
			domain.WithCtx("pigeonID", id))
	}
	return nil
}
