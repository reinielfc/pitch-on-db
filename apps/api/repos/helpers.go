package repos

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("not found")

func toNullableInt64(id *int64) sql.NullInt64 {
	if id == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *id, Valid: true}
}
