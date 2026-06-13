package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"pitch-on-db/internal/repository"

	appErr "pitch-on-db/internal/errors"

	"github.com/gin-gonic/gin"
)

type PigeonHandler struct {
	db *sql.DB
	q  *repository.Queries
}

func NewPigeonHandler(db *sql.DB) *PigeonHandler {
	return &PigeonHandler{db: db, q: repository.New(db)}
}

func (h *PigeonHandler) List(c *gin.Context) {
	pigeons, err := h.q.ListPigeons(c.Request.Context())
	if err != nil {
		c.Error(appErr.DBResource("pigeon", err))
		return
	}

	resp := make([]pigeonResponse, len(pigeons))
	for i, p := range pigeons {
		resp[i] = pigeonResponse{
			ID:         p.ID,
			Name:       p.Name,
			BandNumber: p.BandNumber,
			BirthDate:  p.BirthDate,
			Sex:        p.Sex,
			CreatedAt:  p.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *PigeonHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	pigeon, err := h.q.GetPigeon(c.Request.Context(), id)
	if err != nil {
		c.Error(appErr.DBResource("pigeon", err))
		return
	}

	c.JSON(http.StatusOK, pigeonResponse{
		ID:         pigeon.ID,
		Name:       pigeon.Name,
		BandNumber: pigeon.BandNumber,
		BirthDate:  pigeon.BirthDate,
		Sex:        pigeon.Sex,
		CreatedAt:  pigeon.CreatedAt,
	})
}

func (h *PigeonHandler) Create(c *gin.Context) {
	var req struct {
		Name       string     `json:"name" binding:"required"`
		BandNumber *string    `json:"band_number"`
		BirthDate  *time.Time `json:"birth_date"`
		Sex        *string    `json:"sex"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	pigeon, err := h.q.CreatePigeon(c.Request.Context(), repository.CreatePigeonParams{
		Name:       req.Name,
		BandNumber: req.BandNumber,
		BirthDate:  req.BirthDate,
		Sex:        req.Sex,
	})
	if err != nil {
		c.Error(appErr.DBResource("pigeon", err))
		return
	}

	c.JSON(http.StatusCreated, pigeonRow(pigeon))
}

func (h *PigeonHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req struct {
		Name       *string    `json:"name"`
		BandNumber *string    `json:"band_number"`
		BirthDate  *time.Time `json:"birth_date"`
		Sex        *string    `json:"sex"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	pigeon, err := h.q.UpdatePigeon(c.Request.Context(), repository.UpdatePigeonParams{
		ID:            id,
		Name:          req.Name,
		SetBandNumber: req.BandNumber != nil,
		BandNumber:    req.BandNumber,
		SetBirthDate:  req.BirthDate != nil,
		BirthDate:     req.BirthDate,
		SetSex:        req.Sex != nil,
		Sex:           req.Sex,
	})
	if err != nil {
		c.Error(appErr.DBResource("pigeon", err))
		return
	}

	c.JSON(http.StatusOK, pigeonResponse{
		ID:         pigeon.ID,
		Name:       pigeon.Name,
		BandNumber: pigeon.BandNumber,
		BirthDate:  pigeon.BirthDate,
		Sex:        pigeon.Sex,
		CreatedAt:  pigeon.CreatedAt,
	})
}

func (h *PigeonHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.q.DeletePigeon(c.Request.Context(), id); err != nil {
		c.Error(appErr.DBResource("pigeon", err))
		return
	}

	c.Status(http.StatusNoContent)
}

type pigeonResponse struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	BandNumber *string    `json:"band_number,omitempty"`
	BirthDate  *time.Time `json:"birth_date,omitempty"`
	Sex        *string    `json:"sex,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func pigeonRow(p repository.CreatePigeonRow) pigeonResponse {
	return pigeonResponse{
		ID:         p.ID,
		Name:       p.Name,
		BandNumber: p.BandNumber,
		BirthDate:  p.BirthDate,
		Sex:        p.Sex,
		CreatedAt:  p.CreatedAt,
	}
}
