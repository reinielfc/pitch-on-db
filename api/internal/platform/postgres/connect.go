package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(dataSourceName string) (*sql.DB, error) {
	conn, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return conn, nil
}
