package services

import (
	"context"
	"pitch-on-db/repos"
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
	return s.repo.List(ctx)
}
