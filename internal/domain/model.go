package domain

import "time"

type Pigeon struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	BandNumber *string    `json:"band_number,omitempty"`
	BirthDate  *time.Time `json:"birth_date,omitempty"`
	Sex        *string    `json:"sex,omitempty"`
}

type PigeonPatch struct {
	Name       *string
	BandNumber *string
	BirthDate  *time.Time
	Sex        *string
}
