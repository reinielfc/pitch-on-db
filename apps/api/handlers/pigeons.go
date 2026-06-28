package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/reinielfc/pitch-on-db/apps/api/domain"
	"github.com/reinielfc/pitch-on-db/apps/api/services"

	"github.com/gin-gonic/gin"
)

type PigeonHandler struct {
	svc services.PigeonService
}

func NewPigeonHandler(svc services.PigeonService) *PigeonHandler {
	return &PigeonHandler{svc: svc}
}

func (h *PigeonHandler) Create(c *gin.Context) {
	var req struct {
		Name       string           `json:"name" binding:"required"`
		BirthDate  *time.Time       `json:"birth_date"`
		Sex        *string          `json:"sex"`
		Properties *json.RawMessage `json:"properties"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	sex, ok := parseSex(c, req.Sex)
	if !ok {
		return
	}

	pigeon, err := h.svc.Create(c.Request.Context(), domain.Pigeon{
		Name:       req.Name,
		BirthDate:  req.BirthDate,
		Sex:        sex,
		Properties: req.Properties,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(pigeon))
}

func (h *PigeonHandler) List(c *gin.Context) {
	pigeons, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	resp := make([]pigeonResponse, len(pigeons))
	for i, p := range pigeons {
		resp[i] = toResponse(p)
	}

	c.JSON(http.StatusOK, gin.H{"pigeons": resp})
}

func (h *PigeonHandler) Get(c *gin.Context) {
	id, ok := parseParamID(c)
	if !ok {
		return
	}

	pigeon, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, toResponse(*pigeon))
}

func (h *PigeonHandler) Update(c *gin.Context) {
	id, ok := parseParamID(c)
	if !ok {
		return
	}

	var req struct {
		Name       *string          `json:"name"`
		BirthDate  *time.Time       `json:"birth_date"`
		Sex        *string          `json:"sex"`
		Properties *json.RawMessage `json:"properties"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.Error(err)
		return
	}

	sex, ok := parseSex(c, req.Sex)
	if !ok {
		return
	}

	pigeon, err := h.svc.Update(c.Request.Context(), id, domain.PigeonPatch{
		Name:       req.Name,
		BirthDate:  req.BirthDate,
		Sex:        sex,
		Properties: req.Properties,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, toResponse(pigeon))
}

func (h *PigeonHandler) Delete(c *gin.Context) {
	id, ok := parseParamID(c)
	if !ok {
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PigeonHandler) GetTags(c *gin.Context) {
	id, ok := parseParamID(c)
	if !ok {
		return
	}

	tags, err := h.svc.GetTags(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

func (h *PigeonHandler) SetTags(c *gin.Context) {
	id, ok := parseParamID(c)
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

	if err := h.svc.SetTags(c.Request.Context(), id, req.Tags); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": req.Tags})
}

func (h *PigeonHandler) GetParents(c *gin.Context) {
	id, ok := parseParamID(c)
	if !ok {
		return
	}

	parents, err := h.svc.GetParents(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, parents)
}

func (h *PigeonHandler) GetChildren(c *gin.Context) {
	id, ok := parseParamID(c)
	if !ok {
		return
	}

	children, err := h.svc.GetChildren(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"children": mapToResponses(children)})
}

func (h *PigeonHandler) AssignChild(c *gin.Context) {
	id, ok := parseParamID(c)
	if !ok {
		return
	}

	childID, ok := parseParamAsID(c, "childID")
	if !ok {
		return
	}

	if err := h.svc.AssignChild(c.Request.Context(), id, childID); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

type pigeonResponse struct {
	ID         int64            `json:"id"`
	Name       string           `json:"name"`
	CreatedAt  time.Time        `json:"created_at"`
	BirthDate  *time.Time       `json:"birth_date,omitempty"`
	Sex        *string          `json:"sex,omitempty"`
	Properties *json.RawMessage `json:"properties,omitempty"`
}

func toResponse(p domain.Pigeon) pigeonResponse {
	return pigeonResponse{
		ID:         p.ID,
		Name:       p.Name,
		BirthDate:  p.BirthDate,
		CreatedAt:  p.CreatedAt,
		Sex:        (*string)(p.Sex),
		Properties: p.Properties,
	}
}

func mapToResponses(ps []domain.Pigeon) []pigeonResponse {
	resps := make([]pigeonResponse, len(ps))
	for i, p := range ps {
		resps[i] = toResponse(p)
	}
	return resps
}

func parseSex(c *gin.Context, s *string) (*domain.Sex, bool) {
	if s == nil {
		return nil, true
	}
	sex, err := domain.ParseSex(*s)
	if err != nil {
		c.Error(err)
		return nil, false
	}
	return &sex, true
}
