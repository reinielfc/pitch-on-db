package services

import (
	"context"
	"pitch-on-db/internal/domain"
	"pitch-on-db/internal/repos"
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
	return s.repo.Create(ctx, pigeon)
}

func (s *pigeonService) ListAll(ctx context.Context) ([]domain.Pigeon, error) {
	return s.repo.ListAll(ctx)
}

func (s *pigeonService) Get(ctx context.Context, id int64) (*domain.Pigeon, error) {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, domain.ErrNotFound("pigeon by id %d", id)
	}
	return p, nil
}

func (s *pigeonService) Update(ctx context.Context, id int64, patch domain.PigeonPatch) (domain.Pigeon, error) {
	if patch.Sex != nil {
		if err := s.ensureHasNoChildren(ctx, s.repo, id); err != nil {
			return domain.Pigeon{}, err
		}
	}
	return s.repo.Update(ctx, id, patch)
}

func (s *pigeonService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *pigeonService) AssignChild(ctx context.Context, parentID int64, childID int64) error {
	if parentID == childID {
		return domain.ErrInvalid("cannot assign pigeon %d as its own parent", parentID)
	}

	if err := s.ensureExists(ctx, childID); err != nil {
		return err
	}

	if err := s.ensureExists(ctx, parentID); err != nil {
		return err
	}

	return s.repo.AssignChild(ctx, parentID, childID)
}

func (s *pigeonService) GetParents(ctx context.Context, childID int64) (*domain.PigeonParents, error) {
	if err := s.ensureExists(ctx, childID); err != nil {
		return nil, err
	}
	return s.repo.GetParents(ctx, childID)
}

func (s *pigeonService) GetChildren(ctx context.Context, parentID int64) ([]domain.Pigeon, error) {
	if err := s.ensureExists(ctx, parentID); err != nil {
		return nil, err
	}
	return s.repo.GetChildren(ctx, parentID)
}

func (s *pigeonService) GetTags(ctx context.Context, id int64) ([]string, error) {
	if err := s.ensureExists(ctx, id); err != nil {
		return nil, err
	}
	return s.tagRepo.GetPigeonTags(ctx, id)
}

func (s *pigeonService) SetTags(ctx context.Context, id int64, tagNames []string) error {
	if err := s.ensureExists(ctx, id); err != nil {
		return err
	}
	return s.tagRepo.SetPigeonTags(ctx, id, tagNames)
}

func (s *pigeonService) ensureExists(ctx context.Context, id int64) error {
	if exists, err := s.repo.Exists(ctx, id); err != nil {
		return err
	} else if !exists {
		return domain.ErrNotFound("pigeon by id %d", id)
	}
	return nil
}

func (*pigeonService) ensureHasNoChildren(ctx context.Context, repo repos.PigeonRepository, id int64) error {
	if hasChildren, err := repo.HasChildren(ctx, id); err != nil {
		return err
	} else if hasChildren {
		return domain.ErrInvalid("pigeon %d has children", id)
	}
	return nil
}
