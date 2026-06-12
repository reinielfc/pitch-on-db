package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"pitch-on-db/internal/handlers"
	"pitch-on-db/internal/middleware"

	"github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	App      AppConfig
	Postgres PostgresConfig
}

type AppConfig struct {
	Env  string `env:"APP_ENV"  envDefault:"development"`
	Port string `env:"APP_PORT" envDefault:"8080"`
}

func (c AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("env", c.Env),
		slog.String("port", c.Port),
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

func main() {
	cfg := loadConfig()
	initializeLogger()

	db := connectToPostgresDB(cfg.Postgres)

	r := gin.New()
	r.Use(gin.Recovery())

	if cfg.App.Env == "development" {
		r.Use(middleware.VerboseRequestLogger())
	} else {
		r.Use(middleware.RequestLogger())
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	pigeons := r.Group("/pigeons")
	{
		ph := handlers.NewPigeonHandler(db)
		pigeons.GET("", ph.List)
		pigeons.GET("/:id", ph.Get)
		pigeons.POST("", ph.Create)
		pigeons.PATCH("/:id", ph.Update)
		pigeons.DELETE("/:id", ph.Delete)
	}

	r.Run(":" + cfg.App.Port)
}

func loadConfig() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("parse config: %v", err)
	}

	return cfg
}

func initializeLogger() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func connectToPostgresDB(cfg PostgresConfig) *sql.DB {
	slog.Info("connecting to postgres", "config", cfg)

	db, err := sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password,
		cfg.Host, cfg.Port,
		cfg.DB, cfg.SSLMode))

	if err != nil {
		slog.Error("open db", "error", err)
		os.Exit(1)
	}

	if err := db.Ping(); err != nil {
		slog.Error("ping db", "error", err)
		os.Exit(1)
	}

	slog.Info("connected to postgres")
	return db
}
