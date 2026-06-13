package handlers

import (
	appErr "pitch-on-db/internal/errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(appErr.BadRequest("invalid ID"))
		return 0, false
	}
	return id, true
}
