package routes

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/reinielfc/pitchondb/api/internal/config"
	"github.com/reinielfc/pitchondb/api/internal/platform/httpapi/middleware"
)

type RouterConfig struct {
	title   string
	version string
	appEnv  config.AppEnv
}

func NewRouterConfig(title, version string, appEnv config.AppEnv) RouterConfig {
	return RouterConfig{
		title:   title,
		version: version,
		appEnv:  appEnv,
	}
}

func (c RouterConfig) humaConfig() huma.Config {
	return huma.DefaultConfig(c.title, c.version)
}

type Route func(api huma.API)

func NewRouter(api *huma.API, cfg RouterConfig, routes ...Route) *gin.Engine {
	if cfg.appEnv == config.AppEnvProd {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		// Handle panics and return 500 Internal Server Error
		gin.Recovery(),
		// Generate a unique request ID for each incoming request
		requestid.New(),
		// Bridge the request ID from Gin to Huma for consistent tracing
		middleware.BridgeRequestID(),
	)

	humaAPI := humagin.New(router, cfg.humaConfig())

	for _, route := range routes {
		route(humaAPI)
	}

	*api = humaAPI
	return router
}

func WithRoutes(routes ...Route) Route {
	return func(api huma.API) {
		for _, route := range routes {
			route(api)
		}
	}
}

func WithGroup(path string, routes ...Route) Route {
	return func(api huma.API) {
		group := huma.NewGroup(api, path)

		for _, route := range routes {
			route(group)
		}
	}
}

type HumaMiddleware func(ctx huma.Context, next func(huma.Context))

func WithMiddleware(middlewares ...HumaMiddleware) Route {
	return func(api huma.API) {
		for _, m := range middlewares {
			api.UseMiddleware(m)
		}
	}
}

type Handler interface {
	RegisterRoutes(api huma.API)
}

func WithHandlers(handlers ...Handler) Route {
	return func(api huma.API) {
		for _, handler := range handlers {
			handler.RegisterRoutes(api)
		}
	}
}

func init() {
	// Setup custom Huma error handler
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
