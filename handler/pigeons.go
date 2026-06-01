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

	"github.com/jackc/pgx/v5/pgconn"
	"pitch-on-db/repository"
)
// optional represents a JSON field with three states: absent, explicit null, or a value.
// This allows PATCH handlers to distinguish "not provided" from "set to null".
type optional[T any] struct {
	set   bool
	null  bool
	value T
}

func (o *optional[T]) UnmarshalJSON(b []byte) error {
	o.set = true
	o.null = false
	var zero T
	o.value = zero
	if bytes.Equal(b, []byte("null")) {
		o.null = true
		return nil
	}
	return json.Unmarshal(b, &o.value)
}

type PigeonHandler struct {
	q repository.Querier
}

func NewPigeonHandler(q repository.Querier) *PigeonHandler {
	return &PigeonHandler{q: q}
}

type pigeonResponse struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	BandNumber *string `json:"band_number,omitempty"`
	BirthDate  *string `json:"birth_date,omitempty"`
	Sex        *string `json:"sex,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func nullDatePtr(nt sql.NullTime) *string {
	if !nt.Valid {
		return nil
	}
	// Format in the time's own location to avoid calendar-day shifts on DATE fields.
	s := nt.Time.Format(time.DateOnly)
	return &s
}

func toPigeonResponse(id int64, name string, band sql.NullString, birth sql.NullTime, sex sql.NullString, createdAt time.Time) pigeonResponse {
	return pigeonResponse{
		ID:         id,
		Name:       name,
		BandNumber: nullStringPtr(band),
		BirthDate:  nullDatePtr(birth),
		Sex:        nullStringPtr(sex),
		CreatedAt:  createdAt.UTC().Format(time.RFC3339),
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

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
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
		resp[i] = toPigeonResponse(p.ID, p.Name, p.BandNumber, p.BirthDate, p.Sex, p.CreatedAt)
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /pigeons/{id}
func (h *PigeonHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	p, err := h.q.GetPigeon(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, toPigeonResponse(p.ID, p.Name, p.BandNumber, p.BirthDate, p.Sex, p.CreatedAt))
}

// POST /pigeons
func (h *PigeonHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body struct {
		Name       string  `json:"name"`
		BandNumber *string `json:"band_number"`
		BirthDate  *string `json:"birth_date"`
		Sex        *string `json:"sex"`
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
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	params, err := buildCreateParams(name, body.BandNumber, body.BirthDate, body.Sex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p, err := h.q.CreatePigeon(r.Context(), params)
	if err != nil {
		if isBandNumberConflict(err) {
			http.Error(w, "band_number already exists", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, toPigeonResponse(p.ID, p.Name, p.BandNumber, p.BirthDate, p.Sex, p.CreatedAt))
}

// PATCH /pigeons/{id}
func (h *PigeonHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body struct {
		Name       optional[string] `json:"name"`
		BandNumber optional[string] `json:"band_number"`
		BirthDate  optional[string] `json:"birth_date"`
		Sex        optional[string] `json:"sex"`
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
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	params, err := buildUpdateParams(id, body.Name, body.BandNumber, body.BirthDate, body.Sex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p, err := h.q.UpdatePigeon(r.Context(), params)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		if isBandNumberConflict(err) {
			http.Error(w, "band_number already exists", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, toPigeonResponse(p.ID, p.Name, p.BandNumber, p.BirthDate, p.Sex, p.CreatedAt))
}

// DELETE /pigeons/{id}
func (h *PigeonHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.q.DeletePigeon(r.Context(), id); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseSex(s string) (sql.NullString, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s != "M" && s != "F" {
		return sql.NullString{}, errors.New("sex must be 'M' or 'F'")
	}
	return sql.NullString{String: s, Valid: true}, nil
}

func parseBirthDate(s string) (sql.NullTime, error) {
	t, err := time.Parse(time.DateOnly, strings.TrimSpace(s))
	if err != nil {
		return sql.NullTime{}, errors.New("birth_date must be in YYYY-MM-DD format")
	}
	return sql.NullTime{Time: t, Valid: true}, nil
}

// parseBand trims the value and returns NULL for empty/whitespace strings so
// clients can both omit and clear band_number without triggering UNIQUE conflicts.
func parseBand(s string) sql.NullString {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func buildCreateParams(name string, band, birthDate, sex *string) (repository.CreatePigeonParams, error) {
	p := repository.CreatePigeonParams{Name: name}
	if band != nil {
		p.BandNumber = parseBand(*band)
	}
	if birthDate != nil {
		var err error
		p.BirthDate, err = parseBirthDate(*birthDate)
		if err != nil {
			return p, err
		}
	}
	if sex != nil {
		var err error
		p.Sex, err = parseSex(*sex)
		if err != nil {
			return p, err
		}
	}
	return p, nil
}

func buildUpdateParams(id int64, name, band, birthDate, sex optional[string]) (repository.UpdatePigeonParams, error) {
	p := repository.UpdatePigeonParams{ID: id}

	if name.set {
		n := strings.TrimSpace(name.value)
		if name.null || n == "" {
			return p, errors.New("name cannot be null or empty")
		}
		p.Name = sql.NullString{String: n, Valid: true}
	}

	if band.set {
		p.SetBandNumber = true
		if !band.null {
			p.BandNumber = parseBand(band.value)
		}
	}

	if birthDate.set {
		p.SetBirthDate = true
		if !birthDate.null {
			var err error
			p.BirthDate, err = parseBirthDate(birthDate.value)
			if err != nil {
				return p, err
			}
		}
	}

	if sex.set {
		p.SetSex = true
		if !sex.null {
			var err error
			p.Sex, err = parseSex(sex.value)
			if err != nil {
				return p, err
			}
		}
	}

	return p, nil
}

func isBandNumberConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "pigeons_band_number_key"
}
