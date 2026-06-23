package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"pitch-on-db/config"
)

func Connect(cfg config.PostgresConfig) (*sql.DB, error) {
	slog.Info("connecting to postgres", "config", cfg)

	conn, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	slog.Info("connected to postgres")
	return conn, nil
}
