package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/reinielfc/pitchondb/api/internal/config"
	"github.com/reinielfc/pitchondb/api/internal/pigeon"
	"github.com/reinielfc/pitchondb/api/internal/platform/httpapi/handlers"
	"github.com/reinielfc/pitchondb/api/internal/platform/httpapi/middleware"
	"github.com/reinielfc/pitchondb/api/internal/platform/httpapi/routes"
	"github.com/reinielfc/pitchondb/api/internal/platform/postgres"
	"github.com/reinielfc/pitchondb/api/internal/telemetry"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

type Deps struct{ api huma.API }

func main() {
	startupTime := time.Now()

	var deps = &Deps{}

	cli := humacli.New(func(hooks humacli.Hooks, opts *config.Options) {
		// Determine if the environment is development
		appEnv, err := config.AppEnvString(opts.Env)
		must(err, "invalid application environment", "env", opts.Env)

		// Set up logging and telemetry
		telemetry.SetupLogging(appEnv == config.AppEnvDev)

		// Connect to the database
		db, err := postgres.Open(opts.Postgres.ConnectionString())
		must(err, "failed to connect to database", "config", opts.Postgres)
		slog.Info("connected to database", "config", opts.Postgres)

		// Set up repositories
		pigeonRepo := postgres.NewPigeonRepository(db)
		pigeonQuery := postgres.NewPigeonQueryService(db)

		// Wire up services
		pigeonSvc := pigeon.NewService(pigeonRepo)

		// Wire up handlers
		healthHandler := handlers.NewHealthHandler(startupTime, version, commit, parseBuildTime())
		pigeonsHandler := handlers.NewPigeonsHandler(pigeonSvc, pigeonQuery)

		// Set up HTTP API
		router := routes.NewRouter(&deps.api,
			routes.NewRouterConfig(opts.Name, version, appEnv),
			routes.WithHandlers(healthHandler),
			routes.WithGroup("/v1",
				routes.WithMiddleware(middleware.LogRequests()),
				routes.WithHandlers(pigeonsHandler),
			),
		)

		// Create HTTP server
		server := &http.Server{
			Addr:    fmt.Sprintf(":%s", opts.Port),
			Handler: router,
		}

		// Start the server
		hooks.OnStart(func() {
			slog.Info("starting server",
				"port", opts.Port,
				"version", version,
				"commit", commit,
				"buildTime", buildTime,
			)
			if err = db.Ping(); err != nil {
				slog.Error("ping database", "error", err)
			}

			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("server listen", "error", err)
			}
		})

		// Handle graceful shutdown
		hooks.OnStop(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				slog.Error("server shutdown", "error", err)
			}

			if err := db.Close(); err != nil {
				slog.Error("database close", "error", err)
			}
		})
	})

	cli.Root().AddCommand(deps.healthcheckCmd(), deps.openapiCmd())
	cli.Root().PersistentFlags().SortFlags = false
	cli.Root().Flags().SortFlags = false
	cobra.EnableCommandSorting = false

	cli.Run()
}

func parseBuildTime() time.Time {
	t, err := time.Parse(time.RFC3339, buildTime)
	if err != nil {
		slog.Warn("failed to parse build time", "buildTime", buildTime, "error", err)
		return time.Time{}
	}
	return t
}

var exitFunc = os.Exit

func must(err error, msg string, args ...any) {
	if err == nil {
		return
	}
	slog.Error(msg, append(args, "error", err)...)
	exitFunc(1)
}
