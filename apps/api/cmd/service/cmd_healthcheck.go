package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/reinielfc/pitchondb/apps/api/internal/config"
	"github.com/reinielfc/pitchondb/apps/api/internal/platform/httpapi/handlers"
	"github.com/spf13/cobra"
)

func healthcheckCmd() *cobra.Command {
	var interval string
	var retries int

	healthCheckCmd := &cobra.Command{
		Use:   "healthcheck",
		Short: "Check the health of the API service",
		Run: humacli.WithOptions(func(cmd *cobra.Command, args []string, opts *config.Options) {
			dur, err := time.ParseDuration(interval)
			must(err, "invalid interval duration", "interval", interval)

			url := fmt.Sprintf("http://localhost:%s/health", opts.Port)
			res, err := pollHealth(url, retries, dur)
			must(err, "API service is unhealthy after health check", "url", url, "retries", retries, "interval", interval)

			slog.Info("API service is healthy",
				"status", res.Status,
				"version", res.Body.Version,
				"gitCommit", res.Body.GitCommit,
				"buildTime", res.Body.BuildTime,
				"uptime", res.Body.Uptime,
			)
		}),
	}
	healthCheckCmd.Flags().StringVarP(&interval, "interval", "i", "5s", "Interval between health checks (default: 5s)")
	healthCheckCmd.Flags().IntVarP(&retries, "retries", "r", 5, "Number of retries for health check (default: 5)")

	return healthCheckCmd
}

func pollHealth(url string, retries int, interval time.Duration) (*handlers.HealthCheckOutput, error) {
	var lastOutput *handlers.HealthCheckOutput
	for attempt := 1; attempt <= retries; attempt++ {
		log.Printf("Checking API service health (attempt %d/%d)...", attempt, retries)

		if output, err := fetchHealth(url); err != nil {
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

func fetchHealth(url string) (*handlers.HealthCheckOutput, error) {
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

	var healthStatus handlers.HealthStatus
	if err := json.Unmarshal(body, &healthStatus); err != nil {
		return nil, fmt.Errorf("failed to unmarshal health response: %w", err)
	}

	return &handlers.HealthCheckOutput{
		Status: resp.StatusCode,
		Body:   healthStatus,
	}, nil
}
