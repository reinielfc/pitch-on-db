package config

import (
	"fmt"
	"log/slog"
)

type Options struct {
	App      AppOptions
	Postgres PostgresOptions
}

type AppOptions struct {
	Name    string `doc:"The application name" default:"pitchondb"`
	Env     string `doc:"The application environment" default:"dev"`
	Port    string `doc:"The port to listen on" short:"p" default:"8080"`
	Debug   bool   `doc:"Enable debug mode" short:"d"`
	Verbose bool   `doc:"Enable verbose logging" short:"v"`
}

type PostgresOptions struct {
	Host     string `doc:"The database host" default:"localhost"`
	Port     string `doc:"The database port" default:"5432"`
	User     string `doc:"The database user" default:"pitchondb"`
	Password string `doc:"The database password" default:"pitchondb"`
	DB       string `doc:"The database name" default:"pitchondb"`
	SSLMode  string `doc:"The database SSL mode" default:"disable"`
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
