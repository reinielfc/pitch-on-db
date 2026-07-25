package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/reinielfc/pitchondb/apps/api/internal/platform/httpapi/routes"
)

type HealthHandler struct {
	version     string
	gitCommit   string
	buildTime   time.Time
	startupTime time.Time
}

func NewHealthHandler(startupTime time.Time, version string, gitCommit string, buildTime time.Time) *HealthHandler {
	return &HealthHandler{
		startupTime: startupTime,
		version:     version,
		gitCommit:   gitCommit,
		buildTime:   buildTime,
	}
}

func (h *HealthHandler) RegisterRoutes(api huma.API) {
	huma.Get(api, "/health", h.Check, routes.WithSummary("Health Check"))
}

type HealthCheckOutput struct {
	Status int `json:"-"`
	Body   HealthStatus
}

type HealthStatus struct {
	Status    string    `json:"status" doc:"Health status of the application" example:"ok"`
	Timestamp time.Time `json:"timestamp" doc:"Timestamp of the health check" example:"2023-01-01T00:00:00Z"`
	Version   string    `json:"version" doc:"Version of the application" example:"v1.0.0"`
	GitCommit string    `json:"gitCommit" doc:"Git commit hash of the application" example:"abc123"`
	BuildTime time.Time `json:"buildTime" doc:"Build time of the application" example:"2023-01-01T00:00:00Z"`
	Uptime    string    `json:"uptime" doc:"Uptime of the application in human-readable format" example:"1h0m0s"`
}

func (h *HealthHandler) Check(ctx context.Context, _ *I) (*HealthCheckOutput, error) {
	uptime := time.Since(h.startupTime)
	return &HealthCheckOutput{
		Status: http.StatusOK,
		Body: HealthStatus{
			Status:    "ok",
			Timestamp: time.Now(),
			Uptime:    uptime.Truncate(time.Second).String(),
			Version:   h.version,
			GitCommit: h.gitCommit,
			BuildTime: h.buildTime,
		},
	}, nil
}
