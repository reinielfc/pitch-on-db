package config

import (
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	App      AppConfig
	Postgres PostgresConfig
}

type AppConfig struct {
	Env      string `env:"APP_ENV"       envDefault:"development"`
	Port     string `env:"APP_PORT"      envDefault:"8080"`
	LogLevel string `env:"APP_LOG_LEVEL" envDefault:"info"`
}

func (c AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("env", c.Env),
		slog.String("port", c.Port),
		slog.String("log_level", c.LogLevel),
	)
}

type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST"     envDefault:"localhost"`
	Port     string `env:"POSTGRES_PORT"     envDefault:"5432"`
	User     string `env:"POSTGRES_USER"     envDefault:"pitchondb"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"pitchondb"`
	DB       string `env:"POSTGRES_DB"       envDefault:"pitchondb"`
	SSLMode  string `env:"POSTGRES_SSLMODE"  envDefault:"disable"`
}

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DB, c.SSLMode,
	)
}

func (c PostgresConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", c.Host),
		slog.String("port", c.Port),
		slog.String("user", c.User),
		slog.String("password", "[REDACTED]"),
		slog.String("db", c.DB),
		slog.String("sslmode", c.SSLMode),
	)
}

func Load() (Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}
