package services

import (
	"context"
	"pitch-on-db/internal/domain"
	"pitch-on-db/internal/repos"
)

type PigeonService interface {
	Create(ctx context.Context, pigeon domain.Pigeon) (domain.Pigeon, error)
	List(ctx context.Context) ([]domain.Pigeon, error)
	Get(ctx context.Context, id int64) (*domain.Pigeon, error)
	Update(ctx context.Context, id int64, pigeonPatch domain.PigeonPatch) (domain.Pigeon, error)
	Delete(ctx context.Context, id int64) error

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

func (s *pigeonService) List(ctx context.Context) ([]domain.Pigeon, error) {
	return s.repo.List(ctx)
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

func (s *pigeonService) Create(ctx context.Context, pigeon domain.Pigeon) (domain.Pigeon, error) {
	return s.repo.Create(ctx, pigeon)
}

func (s *pigeonService) Update(ctx context.Context, id int64, patch domain.PigeonPatch) (domain.Pigeon, error) {
	return s.repo.Update(ctx, id, patch)
}

func (s *pigeonService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *pigeonService) GetTags(ctx context.Context, id int64) ([]string, error) {
	return s.tagRepo.GetPigeonTags(ctx, id)
}

func (s *pigeonService) SetTags(ctx context.Context, id int64, tagNames []string) error {
	if exists, err := s.repo.Exists(ctx, id); err != nil {
		return err
	} else if !exists {
		return domain.ErrNotFound("pigeon by id %d", id)
	}

	return s.tagRepo.SetPigeonTags(ctx, id, tagNames)
}
