package domain

import "time"

type Pigeon struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	BandNumber *string    `json:"band_number,omitempty"`
	BirthDate  *time.Time `json:"birth_date,omitempty"`
	Sex        *Sex       `json:"sex,omitempty"`
}

type Sex string

const (
	SexMale   Sex = "M"
	SexFemale Sex = "F"
)

func (s Sex) IsValid() bool {
	return s == SexMale || s == SexFemale
}

func ParseSex(s string) (Sex, error) {
	sex := Sex(s)
	if !sex.IsValid() {
		return "", ErrInvalid("invalid sex value '%s'", s)
	}
	return sex, nil
}

type PigeonPatch struct {
	Name       *string
	BandNumber *string
	BirthDate  *time.Time
	Sex        *Sex
}

type PigeonParents struct {
	Father *Pigeon `json:"father,omitempty"`
	Mother *Pigeon `json:"mother,omitempty"`
}
