package handlers

import (
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/reinielfc/pitchondb/api/internal/pigeon"
	"github.com/reinielfc/pitchondb/api/internal/utils/enums"
	"github.com/reinielfc/pitchondb/api/internal/utils/slicesx"
)

type Pigeon struct {
	ID            UUID                    `json:"id" doc:"Unique identifier of the pigeon" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name          *string                 `json:"name,omitempty" doc:"Name of the pigeon" example:"Buffy"`
	RingNumber    *string                 `json:"ringNumber,omitempty" doc:"Ring number of the pigeon" example:"RN-12345"`
	Sex           PigeonSex               `json:"sex"`
	SexConfidence *PigeonSexConfidence    `json:"sexConfidence,omitempty"`
	Status        PigeonStatus            `json:"status"`
	AcquiredDate  *time.Time              `json:"acquiredDate,omitempty" doc:"Date when the pigeon was acquired" example:"2023-01-01T00:00:00Z"`
	AcquiredVia   PigeonAcquisitionMethod `json:"acquiredVia"`
}

type PigeonList struct {
	Pigeons List[PigeonSummary] `json:"pigeons" doc:"List of pigeons"`
	Total   int                 `json:"total" doc:"Total number of pigeons" example:"100"`
}

type PigeonSummary struct {
	ID            UUID                    `json:"id" doc:"Unique identifier of the pigeon" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name          *string                 `json:"name,omitempty" doc:"Name of the pigeon" example:"Buffy"`
	RingNumber    *string                 `json:"ringNumber,omitempty" doc:"Ring number of the pigeon" example:"RN-12345"`
	Sex           PigeonSex               `json:"sex"`
	SexConfidence *PigeonSexConfidence    `json:"sexConfidence,omitempty"`
	Status        PigeonStatus            `json:"status"`
	AcquiredDate  *time.Time              `json:"acquiredDate,omitempty" doc:"Date when the pigeon was acquired" example:"2023-01-01T00:00:00Z"`
	AcquiredVia   PigeonAcquisitionMethod `json:"acquiredVia"`
}

type PigeonSex string

func (s PigeonSex) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:        "string",
		Description: "Sex of the pigeon",
		Enum:        slicesx.AsAny(pigeon.SexStrings()),
	}
}

type PigeonSexConfidence string

func (s PigeonSexConfidence) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:        "string",
		Description: "Confidence level of the sex determination",
		Enum:        slicesx.AsAny(pigeon.SexConfidenceStrings()),
	}
}

type PigeonStatus string

func (s PigeonStatus) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:        "string",
		Description: "Status of the pigeon",
		Enum:        slicesx.AsAny(pigeon.StatusStrings()),
	}
}

type PigeonAcquisitionMethod string

func (s PigeonAcquisitionMethod) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:        "string",
		Description: "Method of acquisition of the pigeon",
		Enum:        slicesx.AsAny(pigeon.AcquisitionMethodStrings()),
	}
}

func fromDomainPigeon(p *pigeon.Pigeon) Pigeon {
	s := p.Snapshot()

	return Pigeon{
		ID:            fromUUID(s.ID),
		Name:          s.Name,
		RingNumber:    s.RingNumber,
		Sex:           enums.AsType[PigeonSex](s.Sex),
		SexConfidence: enums.AsPtrType[PigeonSexConfidence](s.SexConfidence),
		Status:        enums.AsType[PigeonStatus](s.Status),
		AcquiredDate:  s.AcquiredDate,
		AcquiredVia:   enums.AsType[PigeonAcquisitionMethod](s.AcquiredVia),
	}
}

func fromDomainPigeonSummary(s pigeon.Summary) PigeonSummary {
	return PigeonSummary{
		ID:            fromUUID(s.ID),
		Name:          s.Name,
		RingNumber:    s.RingNumber,
		Sex:           enums.AsType[PigeonSex](s.Sex),
		SexConfidence: enums.AsPtrType[PigeonSexConfidence](s.SexConfidence),
		Status:        enums.AsType[PigeonStatus](s.Status),
		AcquiredDate:  s.AcquiredDate,
		AcquiredVia:   enums.AsType[PigeonAcquisitionMethod](s.AcquiredVia),
	}
}

func mapFromDomainPigeonSummaries(summaries []pigeon.Summary) []PigeonSummary {
	return slicesx.Map(summaries, fromDomainPigeonSummary)
}
