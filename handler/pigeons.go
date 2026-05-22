package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pitch-on-db/repository"
)

type PigeonHandler struct {
	q repository.Querier
}

func NewPigeonHandler(q repository.Querier) *PigeonHandler {
	return &PigeonHandler{q: q}
}

type pigeonResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func toPigeonResponse(p repository.Pigeon) pigeonResponse {
	return pigeonResponse{
		ID:        p.ID,
		Name:      p.Name,
		CreatedAt: p.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		log.Printf("writeJSON encode: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(buf.Bytes())
}

// GET /pigeons
func (h *PigeonHandler) List(w http.ResponseWriter, r *http.Request) {
	pigeons, err := h.q.ListPigeons(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	resp := make([]pigeonResponse, len(pigeons))
	for i, p := range pigeons {
		resp[i] = toPigeonResponse(p)
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /pigeons/{id}
func (h *PigeonHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	pigeon, err := h.q.GetPigeon(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, toPigeonResponse(pigeon))
}

// POST /pigeons
func (h *PigeonHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body struct {
		Name string `json:"name"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	// reject trailing data after the first JSON object
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	pigeon, err := h.q.CreatePigeon(r.Context(), name)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, toPigeonResponse(pigeon))
}
