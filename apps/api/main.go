package main

import (
	"log"

	"github.com/reinielfc/pitch-on-db/apps/api/config"
	"github.com/reinielfc/pitch-on-db/apps/api/db"
	"github.com/reinielfc/pitch-on-db/apps/api/logging"
	"github.com/reinielfc/pitch-on-db/apps/api/middleware"
	"github.com/reinielfc/pitch-on-db/apps/api/repos"
	"github.com/reinielfc/pitch-on-db/apps/api/routes"
	"github.com/reinielfc/pitch-on-db/apps/api/services"

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
