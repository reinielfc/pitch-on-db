package config

import (
	"fmt"
	"log/slog"
)

type Options struct {
	Name     string          `doc:"Application name" default:"pitchondb"`
	Env      string          `doc:"Application environment" default:"dev"`
	Port     string          `short:"p" doc:"Port to listen on" default:"8080"`
	Postgres PostgresOptions `doc:"PostgreSQL database options"`
}

type PostgresOptions struct {
	Host     string `doc:"Database host" default:"localhost"`
	Port     string `doc:"Database port" default:"5432"`
	User     string `doc:"Database user" default:"pitchondb"`
	Password string `doc:"Database password" default:"pitchondb"`
	DB       string `doc:"Database name" default:"pitchondb"`
	SSLMode  string `doc:"Database SSL mode" default:"disable"`
}

func (o PostgresOptions) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		o.Host, o.Port, o.User, o.Password, o.DB, o.SSLMode,
	)
}

func (o PostgresOptions) DSN() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		o.User, o.Password, o.Host, o.Port, o.DB, o.SSLMode,
	)
}

func (o PostgresOptions) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", o.Host),
		slog.String("port", o.Port),
		slog.String("user", o.User),
		slog.String("db", o.DB),
		slog.String("sslmode", o.SSLMode),
	)
}
