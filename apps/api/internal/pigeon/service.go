package pigeon

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	// Add creates a new pigeon record in the database and returns the created record with its assigned ID.
	Add(ctx context.Context) (*Pigeon, error)
	// UpdatePigeonDetails updates the details of a pigeon, including name, ring number, sex, sex confidence, acquired date, and acquisition method.
	UpdatePigeonDetails(ctx context.Context, id uuid.UUID, req UpdatePigeonDetailsRequest) error
}

type pigeonService struct{ repo Repository }

func NewService(repo Repository) Service {
	return &pigeonService{repo: repo}
}

func (s *pigeonService) Add(ctx context.Context) (*Pigeon, error) {
	p := New()
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, fmt.Errorf("add pigeon: %w", err)
	}
	return p, nil
}

type UpdatePigeonDetailsRequest struct {
	Name          *string
	RingNumber    *string
	Sex           *Sex
	SexConfidence *SexConfidence
	AcquiredDate  *time.Time
	AcquiredVia   *AcquisitionMethod
}

func (s *pigeonService) UpdatePigeonDetails(ctx context.Context, id uuid.UUID, req UpdatePigeonDetailsRequest) error {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find pigeon by ID: %w", err)
	}

	if req.Name != nil {
		p.Rename(*req.Name)
	}
	if req.RingNumber != nil {
		p.AssignRingNumber(*req.RingNumber)
	}
	if req.Sex != nil && req.SexConfidence != nil {
		p.DetermineSex(*req.Sex, *req.SexConfidence)
	}
	if req.AcquiredDate != nil && req.AcquiredVia != nil {
		p.Acquire(*req.AcquiredDate, *req.AcquiredVia)
	}

	if err := s.repo.Save(ctx, p); err != nil {
		return fmt.Errorf("save pigeon: %w", err)
	}

	return nil
}

func (s *pigeonService) Remove(ctx context.Context, id uuid.UUID) error {
	// TODO: Instead of deleting the pigeon, we should mark it as inactive or archived to maintain historical data.
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete pigeon: %w", err)
	}
	return nil
}
