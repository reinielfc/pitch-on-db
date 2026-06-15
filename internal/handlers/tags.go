package handlers

import (
	"net/http"
	"pitch-on-db/internal/services"

	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	svc services.TagService
}

func NewTagHandler(svc services.TagService) *TagHandler {
	return &TagHandler{svc: svc}
}

func (h *TagHandler) List(c *gin.Context) {
	tags, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}
