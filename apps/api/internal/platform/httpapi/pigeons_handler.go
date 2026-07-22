package httpapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/reinielfc/pitchondb/apps/api/internal/pigeon"
)

type PigeonsHandler struct {
	svc   pigeon.Service
	query pigeon.QueryService
}

func NewPigeonsHandler(svc pigeon.Service, query pigeon.QueryService) *PigeonsHandler {
	return &PigeonsHandler{svc: svc, query: query}
}

type ListPigeonsInput struct {
	Page  int `query:"page" doc:"Page number" default:"1" minimum:"1"`
	Limit int `query:"limit" doc:"Page size" default:"32" minimum:"1" maximum:"128"`
}

type ListPigeonsOutput struct {
	Status int `json:"-"`
	Body   PigeonList
}

func (h *PigeonsHandler) List(ctx context.Context, input *ListPigeonsInput) (*ListPigeonsOutput, error) {
	f := pigeon.ListFilter{
		Page:  input.Page,
		Limit: input.Limit,
	}
	summaries, total, err := h.query.List(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list pigeons: %w", err)
	}

	var response ListPigeonsOutput
	response.Status = http.StatusOK
	response.Body.Pigeons = mapFromDomainPigeonSummaries(summaries)
	response.Body.Total = total

	return &response, nil
}

type AddPigeonOutput struct {
	Status int `json:"-"`
	Body   Pigeon
}

func (h *PigeonsHandler) Add(ctx context.Context, input *I) (*AddPigeonOutput, error) {
	pigeon, err := h.svc.Add(ctx)
	if err != nil {
		return nil, fmt.Errorf("add pigeon: %w", err)
	}
	return &AddPigeonOutput{
		Status: http.StatusCreated,
		Body:   fromDomainPigeon(pigeon),
	}, nil
}
