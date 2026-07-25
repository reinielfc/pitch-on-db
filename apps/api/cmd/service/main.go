package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

	parsedBuildTime := startupTime
	if BuildTime != "" && BuildTime != "unknown" {
		if t, err := time.Parse(time.RFC3339, BuildTime); err == nil {
			parsedBuildTime = t
		} else {
			slog.Warn("invalid build time format, expected RFC3339", "buildTime", BuildTime, "error", err)
		}
	}

	var api huma.API

	cli := humacli.New(func(hooks humacli.Hooks, opts *config.Options) {
		// Set up logging and telemetry
		telemetry.SetupLogging(opts.Debug)

		// Connect to the database
		postgresDSN := opts.Postgres.DSN()
		slog.Debug("connecting to database", "dsn", postgresDSN)

		postgresDB, err := postgres.Connect(postgresDSN)
		if err != nil {
			slog.Error("failed to connect to database",
				"config", opts.Postgres,
				"error", err)
			os.Exit(1)
		}
		slog.Info("connected to database", "config", opts.Postgres)

		// Set up repositories
		pigeonRepo := postgres.NewPigeonRepository(postgresDB)
		pigeonQuery := postgres.NewPigeonQueryService(postgresDB)

		// Wire up services
		pigeonSvc := pigeon.NewService(pigeonRepo)

		// Set up HTTP API
		router := httpapi.SetupRouter(&api,
			httpapi.DefaultConfig(opts.Name, Version),
			httpapi.WithHealthRoute(startupTime, Version, GitCommit, parsedBuildTime),
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
				os.Exit(1)
			}
		})

		// Handle graceful shutdown
		hooks.OnStop(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				slog.Error("server shutdown", "error", err)
				os.Exit(1)
			}

			if err := postgresDB.Close(); err != nil {
				slog.Error("database close", "error", err)
				os.Exit(1)
			}
		})
	})

	var output string

	openapiCmd := &cobra.Command{
		Use:   "openapi",
		Short: "Generate OpenAPI specification",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Generating OpenAPI specification...")
			if api == nil {
				log.Fatal("API is not initialized")
			}

			b, err := api.OpenAPI().YAML()
			if err != nil {
				log.Fatalf("failed to generate OpenAPI spec: %v", err)
			}

			if output != "" {
				if err := os.WriteFile(output, b, 0644); err != nil {
					log.Fatalf("failed to write OpenAPI spec to file: %v", err)
				}
				log.Printf("OpenAPI specification written to file: %s", output)
			} else {
				fmt.Println(string(b))
			}
		},
	}
	openapiCmd.Flags().StringVarP(&output, "output", "o", "", "Output file for OpenAPI specification (optional)")
	cli.Root().AddCommand(openapiCmd)

	var interval string
	var retries int

	healthCheckCmd := &cobra.Command{
		Use:   "healthcheck",
		Short: "Check the health of the API service",
		Run: humacli.WithOptions(func(cmd *cobra.Command, args []string, opts *config.Options) {
			dur, err := time.ParseDuration(interval)
			if err != nil {
				log.Fatalf("invalid interval %q: %v", interval, err)
			}

			url := fmt.Sprintf("http://localhost:%s/health", opts.Port)
			if res, err := pollHealth(url, retries, dur); err != nil {
				log.Fatal(err)
			} else {
				slog.Info("API service is healthy",
					"status", res.Status,
					"version", res.Body.Version,
					"gitCommit", res.Body.GitCommit,
					"buildTime", res.Body.BuildTime,
					"uptime", res.Body.Uptime,
				)
			}
		}),
	}
	healthCheckCmd.Flags().StringVarP(&interval, "interval", "i", "5s", "Interval between health checks (default: 5s)")
	healthCheckCmd.Flags().IntVarP(&retries, "retries", "r", 5, "Number of retries for health check (default: 5)")

	cli.Root().AddCommand(healthCheckCmd)

	cli.Run()
}

func pollHealth(url string, retries int, interval time.Duration) (*httpapi.HealthCheckOutput, error) {
	var lastOutput *httpapi.HealthCheckOutput
	for attempt := 1; attempt <= retries; attempt++ {
		log.Printf("Checking API service health (attempt %d/%d)...", attempt, retries)

		output, err := fetchHealth(url)
		if err != nil {
			log.Printf("health check failed: %v", err)
		} else {
			lastOutput = output
		}

		if lastOutput != nil && lastOutput.Status == http.StatusOK {
			return lastOutput, nil
		}

		if attempt < retries {
			log.Printf("retrying in %s...", interval)
			time.Sleep(interval)
		}
	}
	return lastOutput, fmt.Errorf("API service is unhealthy after %d attempt(s)", retries)
}

func fetchHealth(url string) (*httpapi.HealthCheckOutput, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("failed to close health response body: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read health response body: %w", err)
	}

	var healthStatus httpapi.HealthStatus
	if err := json.Unmarshal(body, &healthStatus); err != nil {
		return nil, fmt.Errorf("failed to unmarshal health response: %w", err)
	}

	return &httpapi.HealthCheckOutput{
		Status: resp.StatusCode,
		Body:   healthStatus,
	}, nil
}
