package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/reinielfc/pitch-on-db/apps/api/domain"
)

func parseID(c *gin.Context, id string) (int64, bool) {
	parsedID, err := domain.ParseID(id)
	if err != nil {
		c.Error(err)
		return 0, false
	}
	return parsedID, true
}

func parseParamAsID(c *gin.Context, name string) (int64, bool) { return parseID(c, c.Param(name)) }
func parseQueryAsID(c *gin.Context, name string) (int64, bool) { return parseID(c, c.Query(name)) }
func parseParamID(c *gin.Context) (int64, bool)                { return parseParamAsID(c, "id") }
