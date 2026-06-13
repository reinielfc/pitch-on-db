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
	Env      string `env:"APP_ENV"       envDefault:"development"`
	Port     string `env:"APP_PORT"      envDefault:"8080"`
	LogLevel string `env:"APP_LOG_LEVEL" envDefault:"info"`
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
	initializeLogger(cfg.App.LogLevel)
	db := connectToPostgresDB(cfg.Postgres)

	r := gin.New()
	r.Use(gin.Recovery())

	if cfg.App.Env == "development" {
		r.Use(middleware.VerboseRequestLogger())
	} else {
		r.Use(middleware.RequestLogger())
	}

	r.Use(middleware.ErrorHandler())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	ph := handlers.NewPigeonHandler(db)
	r.GET("/pigeons", ph.List)
	r.POST("/pigeons", ph.Create)

	pigeons := r.Group("/pigeons")
	{
		pigeons.GET("/:id", ph.Get)
		pigeons.PATCH("/:id", ph.Update)
		pigeons.DELETE("/:id", ph.Delete)

		th := handlers.NewTagHandler(db)
		pigeons.GET("/:id/tags", th.List)
		pigeons.PUT("/:id/tags", th.Set)

		tags := pigeons.Group("/:id/tags")
		{
			tags.POST("/:name", th.Add)
			tags.DELETE("/:name", th.Remove)
		}
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

func initializeLogger(lvl string) {
	var level slog.Level
	err := level.UnmarshalText([]byte(lvl))
	if err != nil {
		log.Fatalf("invalid log level: %v", err)
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("starting logger", "level", level)
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
