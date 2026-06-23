package services

import (
	"context"

	"github.com/reinielfc/pitch-on-db/apps/api/domain"
	"github.com/reinielfc/pitch-on-db/apps/api/repos"
)

type PigeonService interface {
	Create(ctx context.Context, pigeon domain.Pigeon) (domain.Pigeon, error)
	ListAll(ctx context.Context) ([]domain.Pigeon, error)
	Get(ctx context.Context, id int64) (*domain.Pigeon, error)
	Update(ctx context.Context, id int64, pigeonPatch domain.PigeonPatch) (domain.Pigeon, error)
	Delete(ctx context.Context, id int64) error

	GetParents(ctx context.Context, childID int64) (*domain.PigeonParents, error)
	GetChildren(ctx context.Context, parentID int64) ([]domain.Pigeon, error)
	AssignChild(ctx context.Context, parentID int64, childID int64) error

	GetTags(ctx context.Context, pigeonID int64) ([]string, error)
	SetTags(ctx context.Context, pigeonID int64, tagNames []string) error
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
	return createdPigeon, domain.Errorf("Create: %w", err)
}

func (s *pigeonService) ListAll(ctx context.Context) ([]domain.Pigeon, error) {
	pigeons, err := s.repo.ListAll(ctx)
	return pigeons, domain.Errorf("ListAll: %w", err)
}

func (s *pigeonService) Get(ctx context.Context, id int64) (*domain.Pigeon, error) {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, domain.Errorf("Get: %w", err)
	}
	if p == nil {
		return nil, domain.NewResourceNotFoundError("pigeon", id)
	}
	return p, nil
}

func (s *pigeonService) Update(ctx context.Context, id int64, patch domain.PigeonPatch) (domain.Pigeon, error) {
	if patch.Sex != nil {
		if err := s.ensureHasNoChildren(ctx, id); err != nil {
			return domain.Pigeon{}, domain.Errorf("Update: %w", err)
		}
	}
	updatedPigeon, err := s.repo.Update(ctx, id, patch)
	if err != nil {
		return domain.Pigeon{}, domain.Errorf("Update: %w", err)
	}
	return updatedPigeon, nil
}

func (s *pigeonService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return domain.Errorf("Delete: %w", err)
	}
	return nil
}

func (s *pigeonService) AssignChild(ctx context.Context, parentID int64, childID int64) error {
	if parentID == childID {
		return domain.NewValidationError(
			domain.WithMsg("cannot assign pigeon %d as its own parent", parentID),
			domain.WithCtx("parentID", parentID),
			domain.WithCtx("childID", childID),
		)
	}

	if err := s.ensureExists(ctx, childID); err != nil {
		return domain.Errorf("AssignChild/child: %w", err)
	}

	if err := s.ensureExists(ctx, parentID); err != nil {
		return domain.Errorf("AssignChild/parent: %w", err)
	}

	if err := s.repo.AssignChild(ctx, parentID, childID); err != nil {
		return domain.Errorf("AssignChild: %w", err)
	}
	return nil
}

func (s *pigeonService) GetParents(ctx context.Context, childID int64) (*domain.PigeonParents, error) {
	if err := s.ensureExists(ctx, childID); err != nil {
		return nil, domain.Errorf("GetParents: %w", err)
	}
	parents, err := s.repo.GetParents(ctx, childID)
	if err != nil {
		return nil, domain.Errorf("GetParents: %w", err)
	}
	return parents, nil
}

func (s *pigeonService) GetChildren(ctx context.Context, parentID int64) ([]domain.Pigeon, error) {
	if err := s.ensureExists(ctx, parentID); err != nil {
		return nil, domain.Errorf("GetChildren: %w", err)
	}
	children, err := s.repo.GetChildren(ctx, parentID)
	if err != nil {
		return nil, domain.Errorf("GetChildren: %w", err)
	}
	return children, nil
}

func (s *pigeonService) GetTags(ctx context.Context, id int64) ([]string, error) {
	if err := s.ensureExists(ctx, id); err != nil {
		return nil, domain.Errorf("GetTags: %w", err)
	}
	tags, err := s.tagRepo.GetPigeonTags(ctx, id)
	if err != nil {
		return nil, domain.Errorf("GetTags: %w", err)
	}
	return tags, nil
}

func (s *pigeonService) SetTags(ctx context.Context, id int64, tagNames []string) error {
	if err := s.ensureExists(ctx, id); err != nil {
		return domain.Errorf("SetTags: %w", err)
	}
	if err := s.tagRepo.SetPigeonTags(ctx, id, tagNames); err != nil {
		return domain.Errorf("SetTags: %w", err)
	}
	return nil
}

func (s *pigeonService) ensureExists(ctx context.Context, id int64) error {
	if exists, err := s.repo.Exists(ctx, id); err != nil {
		return domain.NewInternalError(
			domain.WithFmt("ensureExists: %w", err),
			domain.WithCtx("id", id),
		)
	} else if !exists {
		return domain.NewResourceNotFoundError("pigeon", id)
	}
	return nil
}

func (s *pigeonService) ensureHasNoChildren(ctx context.Context, id int64) error {
	if hasChildren, err := s.repo.HasChildren(ctx, id); err != nil {
		return domain.NewInternalError(
			domain.WithFmt("ensureHasNoChildren: %w", err),
			domain.WithCtx("id", id),
		)
	} else if hasChildren {
		return domain.NewValidationError(
			domain.WithMsg("pigeon has children: %d", id),
			domain.WithCtx("id", id),
		)
	}
	return nil
}
