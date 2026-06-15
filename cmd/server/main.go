package main

import (
	"log"

	"pitch-on-db/internal/config"
	"pitch-on-db/internal/db"
	"pitch-on-db/internal/logging"
	"pitch-on-db/internal/middleware"
	"pitch-on-db/internal/repos"
	"pitch-on-db/internal/routes"
	"pitch-on-db/internal/services"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	_, err = logging.Init(cfg.App.LogLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	db, err := db.Connect(cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	pigeonRepo := repos.NewPigeonRepository(db)
	tagsRepo := repos.NewTagRepository(db)
	pigeonSvc := services.NewPigeonService(pigeonRepo, tagsRepo)
	tagSvc := services.NewTagService(tagsRepo)

	r := gin.New()
	r.Use(gin.Recovery())

	if cfg.App.Env == "development" {
		r.Use(middleware.VerboseRequestLogger())
	} else {
		r.Use(middleware.RequestLogger())
	}

	r.Use(middleware.ErrorHandler())

	routes.Register(r, &routes.Dependencies{
		PigeonService: pigeonSvc,
		TagService:    tagSvc,
	})

	r.Run(":" + cfg.App.Port)
}
