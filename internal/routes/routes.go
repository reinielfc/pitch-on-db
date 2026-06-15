package routes

import (
	"net/http"
	"pitch-on-db/internal/handlers"
	"pitch-on-db/internal/services"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	PigeonService services.PigeonService
	TagService    services.TagService
}

func Register(r *gin.Engine, deps *Dependencies) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	registerPigeonRoutes(r, deps)
	registerTagRoutes(r, deps)
}

func registerPigeonRoutes(r *gin.Engine, deps *Dependencies) {
	ph := handlers.NewPigeonHandler(deps.PigeonService)

	pigeonsGroup := r.Group("/pigeons")
	{
		pigeonsGroup.GET("", ph.List)
		pigeonsGroup.POST("", ph.Create)

		pigeonsGroup.GET("/:id", ph.Get)
		pigeonsGroup.PATCH("/:id", ph.Update)
		pigeonsGroup.DELETE("/:id", ph.Delete)

		tagsGroup := pigeonsGroup.Group("/:id/tags")
		{
			tagsGroup.GET("", ph.GetTags)
			tagsGroup.PUT("", ph.SetTags)
		}
	}
}

func registerTagRoutes(r *gin.Engine, deps *Dependencies) {
	th := handlers.NewTagHandler(deps.TagService)

	r.GET("/tags", th.List)
}
