package handlers

import (
	"database/sql"
	"net/http"

	appErr "pitch-on-db/internal/errors"
	"pitch-on-db/internal/repository"

	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	db *sql.DB
	q  *repository.Queries
}

func NewTagHandler(db *sql.DB) *TagHandler {
	return &TagHandler{db: db, q: repository.New(db)}
}

// List returns all tags for a pigeon by ID.
func (h *TagHandler) List(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	tags, err := h.q.GetPigeonTags(c.Request.Context(), id)
	if err != nil {
		c.Error(appErr.DBResource("pigeon tags", err))
		return
	}

	resp := make([]string, len(tags))
	for i, t := range tags {
		resp[i] = t.Name
	}

	c.JSON(http.StatusOK, resp)
}

// Set replaces all tags for a pigeon atomically.
func (h *TagHandler) Set(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req struct {
		Tags []string `json:"tags" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	tx, err := h.db.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.Error(appErr.DBResource("pigeon tags", err))
		return
	}
	defer tx.Rollback()

	q := repository.New(tx)

	if err := q.ClearPigeonTags(c.Request.Context(), id); err != nil {
		c.Error(appErr.DBResource("pigeon tags", err))
		return
	}

	for _, name := range req.Tags {
		tag, err := q.UpsertTag(c.Request.Context(), name)
		if err != nil {
			c.Error(appErr.DBResource("pigeon tags", err))
			return
		}
		if err := q.AddPigeonTag(c.Request.Context(), repository.AddPigeonTagParams{
			PigeonID: id,
			TagID:    tag.ID,
		}); err != nil {
			c.Error(appErr.DBResource("pigeon tags", err))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.Error(appErr.DBResource("pigeon tags", err))
		return
	}

	c.JSON(http.StatusOK, req.Tags)
}

// Add upserts a tag and attaches it to a pigeon.
func (h *TagHandler) Add(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	tx, err := h.db.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.Error(appErr.DBResource("pigeon tags", err))
		return
	}
	defer tx.Rollback()

	q := repository.New(tx)

	tag, err := q.UpsertTag(c.Request.Context(), c.Param("name"))
	if err != nil {
		c.Error(appErr.DBResource("pigeon tags", err))
		return
	}
	if err := q.AddPigeonTag(c.Request.Context(), repository.AddPigeonTagParams{
		PigeonID: id,
		TagID:    tag.ID,
	}); err != nil {
		c.Error(appErr.DBResource("pigeon tags", err))
		return
	}

	if err := tx.Commit(); err != nil {
		c.Error(appErr.DBResource("pigeon tags", err))
		return
	}

	c.Status(http.StatusNoContent)
}

// Remove detaches a tag from a pigeon by name.
func (h *TagHandler) Remove(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.q.RemovePigeonTag(c.Request.Context(), repository.RemovePigeonTagParams{
		PigeonID: id,
		Name:     c.Param("name"),
	}); err != nil {
		c.Error(appErr.DBResource("pigeon tags", err))
		return
	}

	c.Status(http.StatusNoContent)
}
