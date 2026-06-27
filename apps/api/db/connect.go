package db

import (
	"database/sql"
	"fmt"
	"log/slog"
)

func Connect(dsn string) (*sql.DB, error) {
	slog.Info("connecting to postgres", "config", dsn)

	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	slog.Info("connected to postgres")
	return conn, nil
}
