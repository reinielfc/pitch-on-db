package handlers

import (
	"strconv"

	"pitch-on-db/internal/domain"

	"github.com/gin-gonic/gin"
)

func parseID(c *gin.Context, id string) (int64, bool) {
	parsedID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.Error(domain.ErrInvalid("invalid ID: %s", id))
		return 0, false
	}
	return parsedID, true
}

func parseParamAsID(c *gin.Context, name string) (int64, bool) { return parseID(c, c.Param(name)) }
func parseQueryAsID(c *gin.Context, name string) (int64, bool) { return parseID(c, c.Query(name)) }
func parseParamID(c *gin.Context) (int64, bool)                { return parseParamAsID(c, "id") }
