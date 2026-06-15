package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(ErrBadRequest("invalid ID: %s", c.Param("id")))
		return 0, false
	}
	return id, true
}
