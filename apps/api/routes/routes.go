package routes

import (
	"net/http"
	"pitch-on-db/handlers"
	"pitch-on-db/services"

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

	pigeons := r.Group("/pigeons")
	{
		pigeons.GET("", ph.ListAll)
		pigeons.POST("", ph.Create)

		pigeon := pigeons.Group("/:id")
		{
			pigeon.GET("", ph.Get)
			pigeon.PATCH("", ph.Update)
			pigeon.DELETE("", ph.Delete)

			pigeon.GET("/tags", ph.GetTags)
			pigeon.PUT("/tags", ph.SetTags)

			pigeon.GET("/parents", ph.GetParents)
			pigeon.GET("/children", ph.GetChildren)
			pigeon.PUT("/children/:childID", ph.AssignChild)
		}
	}
}

func registerTagRoutes(r *gin.Engine, deps *Dependencies) {
	th := handlers.NewTagHandler(deps.TagService)

	r.GET("/tags", th.List)
}
