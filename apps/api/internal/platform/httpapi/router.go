package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/reinielfc/pitchondb/apps/api/internal/pigeon"
	"github.com/reinielfc/pitchondb/apps/api/internal/platform/httpapi/middleware"
)

type Route func(api huma.API)

func SetupRouter(api *huma.API, config huma.Config, routes ...Route) *gin.Engine {
	router := newRouter()

	humaAPI := humagin.New(router, config)
	setupCustomErrorHandler()

	for _, route := range routes {
		route(humaAPI)
	}

	*api = humaAPI
	return router
}

func DefaultConfig(appName, version string) huma.Config {
	return huma.DefaultConfig(appName, version)
}

func WithHealthRoute(startupTime time.Time, version string, gitCommit string, buildTime time.Time) Route {
	h := NewHealthHandler(startupTime, version, gitCommit, buildTime)
	return func(api huma.API) {
		huma.Get(api, "/health", h.Check, withSummary("Health Check"))
	}
}

func WithPigeonRoutes(svc pigeon.Service, querySvc pigeon.QueryService) Route {
	h := NewPigeonsHandler(svc, querySvc)
	return func(api huma.API) {
		pigeons := huma.NewGroup(api, "/pigeons")
		{
			huma.Get(pigeons, "", h.List, withSummary("List Pigeons"))
			huma.Post(pigeons, "", h.Add, withSummary("Add Pigeon"))
		}
	}
}

func WithGroupRoutes(path string, routes ...Route) Route {
	return func(api huma.API) {
		group := huma.NewGroup(api, path)
		group.UseMiddleware(middleware.LogRequests())

		for _, route := range routes {
			route(group)
		}
	}
}

func newRouter() *gin.Engine {
	router := gin.New()
	router.Use(
		// Handle panics and return 500 Internal Server Error
		gin.Recovery(),
		// Generate a unique request ID for each incoming request
		requestid.New(),
		// Bridge the request ID from Gin to Huma for consistent tracing
		middleware.BridgeRequestID(),
	)

	return router
}

func setupCustomErrorHandler() {
	humaNewError := huma.NewError

	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		if len(errs) > 0 {
			messages := make([]string, len(errs))
			for i, err := range errs {
				messages[i] = err.Error()
			}
			slog.Error("request error",
				"status", status,
				"message", msg,
				"errors", messages)

			if status == http.StatusBadRequest {
				return humaNewError(status, msg, errs...)
			}
		}

		return humaNewError(status, msg)
	}
}

func withSummary(summary string) func(o *huma.Operation) {
	return func(o *huma.Operation) {
		o.Summary = summary
	}
}
