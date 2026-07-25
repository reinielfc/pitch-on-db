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
	"github.com/reinielfc/pitchondb/apps/api/internal/config"
	"github.com/reinielfc/pitchondb/apps/api/internal/pigeon"
	"github.com/reinielfc/pitchondb/apps/api/internal/platform/httpapi"
	"github.com/reinielfc/pitchondb/apps/api/internal/platform/postgres"
	"github.com/reinielfc/pitchondb/apps/api/internal/telemetry"
	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func main() {
	startupTime := time.Now()

	var api huma.API

	cli := humacli.New(func(hooks humacli.Hooks, opts *config.Options) {
		// Set up logging and telemetry
		telemetry.SetupLogging(opts.Debug)

		// Connect to the database
		postgresDB, err := postgres.Connect(opts.Postgres.ConnectionString())
		must(err, "failed to connect to database", "config", opts.Postgres)
		slog.Info("connected to database", "config", opts.Postgres)

		// Set up repositories
		pigeonRepo := postgres.NewPigeonRepository(postgresDB)
		pigeonQuery := postgres.NewPigeonQueryService(postgresDB)

		// Wire up services
		pigeonSvc := pigeon.NewService(pigeonRepo)

		// Set up HTTP API
		router := httpapi.SetupRouter(&api,
			httpapi.DefaultConfig(opts.Name, Version),
			httpapi.WithHealthRoute(startupTime, Version, GitCommit, buildTime()),
			httpapi.WithGroupRoutes("/v1",
				httpapi.WithPigeonRoutes(pigeonSvc, pigeonQuery),
			),
		)

		// Create HTTP server
		server := &http.Server{
			Addr:    fmt.Sprintf(":%s", opts.Port),
			Handler: router,
		}

		// Start the server
		hooks.OnStart(func() {
			slog.Info("starting server", "version", Version, "port", opts.Port)
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

			if err := postgresDB.Close(); err != nil {
				slog.Error("database close", "error", err)
			}
		})
	})

	cli.Root().AddCommand(healthcheckCmd(), openapiCmd(api))
	cli.Root().PersistentFlags().SortFlags = false
	cli.Root().Flags().SortFlags = false
	cobra.EnableCommandSorting = false

	cli.Run()
}

func buildTime() time.Time {
	t, err := time.Parse(time.RFC3339, BuildTime)
	if err != nil {
		slog.Warn("failed to parse build time", "buildTime", BuildTime, "error", err)
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
