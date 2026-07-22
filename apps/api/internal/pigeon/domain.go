package pigeon

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/reinielfc/pitchondb/apps/api/internal/utils/enums"
)

var (
	ErrNotFound                 = errors.New("pigeon not found")
	ErrInvalidSex               = errors.New("invalid sex value")
	ErrInvalidSexConfidence     = errors.New("invalid sex confidence value")
	ErrInvalidStatus            = errors.New("invalid status value")
	ErrInvalidAcquisitionMethod = errors.New("invalid acquisition method value")
)

type Pigeon struct {
	id            uuid.UUID
	name          *string
	ringNumber    *string
	sex           Sex
	sexConfidence *SexConfidence
	status        Status
	acquiredDate  *time.Time
	acquiredVia   AcquisitionMethod
	createdAt     time.Time
	updatedAt     time.Time
}

func New() *Pigeon {
	return &Pigeon{
		id:          uuid.New(),
		sex:         SexUnknown,
		status:      StatusActive,
		acquiredVia: AcquisitionMethodUnknown,
	}
}

func Reconstitute(
	id uuid.UUID,
	name *string,
	ringNumber *string,
	sex Sex,
	sexConfidence *SexConfidence,
	status Status,
	acquiredDate *time.Time,
	acquiredVia AcquisitionMethod,
	createdAt time.Time,
	updatedAt time.Time,
) *Pigeon {
	return &Pigeon{
		id:            id,
		name:          name,
		ringNumber:    ringNumber,
		sex:           sex,
		sexConfidence: sexConfidence,
		status:        status,
		acquiredDate:  acquiredDate,
		acquiredVia:   acquiredVia,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

func (p *Pigeon) Rename(name string)                 { p.name = &name }
func (p *Pigeon) AssignRingNumber(ringNumber string) { p.ringNumber = &ringNumber }

func (p *Pigeon) MarkDeceased() { p.status = StatusDeceased }
func (p *Pigeon) MarkSold()     { p.status = StatusSold }
func (p *Pigeon) MarkLost()     { p.status = StatusLost }

func (p *Pigeon) DetermineSex(sex Sex, confidence SexConfidence) {
	p.sex = sex
	p.sexConfidence = &confidence
}

func (p *Pigeon) Acquire(acquiredDate time.Time, acquiredVia AcquisitionMethod) {
	p.acquiredDate = &acquiredDate
	p.acquiredVia = acquiredVia
}

type Snapshot struct {
	ID            uuid.UUID
	Name          *string
	RingNumber    *string
	Sex           Sex
	SexConfidence *SexConfidence
	Status        Status
	AcquiredDate  *time.Time
	AcquiredVia   AcquisitionMethod
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (p *Pigeon) Snapshot() Snapshot {
	return Snapshot{
		ID:            p.id,
		Name:          p.name,
		RingNumber:    p.ringNumber,
		Sex:           p.sex,
		SexConfidence: p.sexConfidence,
		Status:        p.status,
		AcquiredDate:  p.acquiredDate,
		AcquiredVia:   p.acquiredVia,
		CreatedAt:     p.createdAt,
		UpdatedAt:     p.updatedAt,
	}
}

//go:generate go tool enumer -type=Sex,SexConfidence,Status,AcquisitionMethod -trimprefix=SexConfidence,Sex,Status,AcquisitionMethod -transform=snake -values -output=domain_enumer.go

// Sex encodes the sex of a pigeon.
type Sex int

const (
	// SexMale indicates a male pigeon.
	SexMale Sex = iota
	// SexFemale indicates a female pigeon.
	SexFemale
	// SexUnknown indicates an unknown sex for a pigeon.
	SexUnknown
)

// SexConfidence encodes the confidence type of the sex determination for a pigeon.
type SexConfidence int

const (
	// SexConfidenceConfirmed indicates that the sex of the pigeon is confirmed.
	SexConfidenceConfirmed SexConfidence = iota
	// SexConfidencePresumed indicates that the sex of the pigeon is presumed but not confirmed.
	SexConfidencePresumed
)

func (s *SexConfidence) StringPtr() *string { return enums.StringPtr(s) }

func SexConfidenceStringPtr(s *string) (*SexConfidence, error) {
	return enums.ParsePtr(s, SexConfidenceString)
}

// Status encodes the status of a pigeon.
type Status int

const (
	// StatusActive indicates that the pigeon is active.
	StatusActive Status = iota
	// StatusDeceased indicates that the pigeon is deceased.
	StatusDeceased
	// StatusSold indicates that the pigeon has been sold.
	StatusSold
	// StatusLost indicates that the pigeon is lost.
	StatusLost
	// StatusUnknown indicates an unknown status for the pigeon.
	StatusUnknown
)

// AcquisitionMethod encodes the method by which a pigeon was acquired.
type AcquisitionMethod int

const (
	// AcquisitionMethodBred indicates that the pigeon was bred by the owner.
	AcquisitionMethodBred AcquisitionMethod = iota
	// AcquisitionMethodPurchased indicates that the pigeon was purchased.
	AcquisitionMethodPurchased
	// AcquisitionMethodGifted indicates that the pigeon was gifted to the owner.
	AcquisitionMethodGifted
	// AcquisitionMethodCaptured indicates that the pigeon was captured.
	AcquisitionMethodCaptured
	// AcquisitionMethodUnknown indicates that the acquisition method of the pigeon is unknown.
	AcquisitionMethodUnknown
)
