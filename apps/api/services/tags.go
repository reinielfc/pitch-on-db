package services

import (
	"context"

	"github.com/reinielfc/pitch-on-db/apps/api/domain"
	"github.com/reinielfc/pitch-on-db/apps/api/repos"
)

type TagService interface {
	List(ctx context.Context) ([]string, error)
}

type tagsService struct {
	repo repos.TagRepository
}

func NewTagService(r repos.TagRepository) TagService {
	return &tagsService{repo: r}
}

func (s *tagsService) List(ctx context.Context) ([]string, error) {
	tags, err := s.repo.List(ctx)
	if err != nil {
		return nil, domain.NewInternalError(
			domain.WithErr("list tags: %w", err),
		)
	}
	return tags, nil
}
