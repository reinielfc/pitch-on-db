package handler_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pitch-on-db/handler"
	"pitch-on-db/repository"
)

var errDB = errors.New("db error")

type mockQuerier struct {
	pigeons []repository.Pigeon
	err     error
}

func (m *mockQuerier) ListPigeons(_ context.Context) ([]repository.Pigeon, error) {
	return m.pigeons, m.err
}

func (m *mockQuerier) GetPigeon(_ context.Context, id int64) (repository.Pigeon, error) {
	if m.err != nil {
		return repository.Pigeon{}, m.err
	}
	for _, p := range m.pigeons {
		if p.ID == id {
			return p, nil
		}
	}
	return repository.Pigeon{}, sql.ErrNoRows
}

func (m *mockQuerier) CreatePigeon(_ context.Context, name string) (repository.Pigeon, error) {
	if m.err != nil {
		return repository.Pigeon{}, m.err
	}
	return repository.Pigeon{ID: 1, Name: name, CreatedAt: time.Now()}, nil
}

// --- List ---

func TestList(t *testing.T) {
	q := &mockQuerier{pigeons: []repository.Pigeon{{ID: 1, Name: "Percy", CreatedAt: time.Now()}}}
	h := handler.NewPigeonHandler(q)

	req := httptest.NewRequest(http.MethodGet, "/pigeons", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
	var resp []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 1 || resp[0]["name"] != "Percy" {
		t.Fatalf("unexpected response: %v", resp)
	}
}

func TestListEmpty(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	req := httptest.NewRequest(http.MethodGet, "/pigeons", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 0 {
		t.Fatalf("expected empty array, got %v", resp)
	}
}

func TestListDBError(t *testing.T) {
	q := &mockQuerier{err: errDB}
	h := handler.NewPigeonHandler(q)

	req := httptest.NewRequest(http.MethodGet, "/pigeons", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

// --- Get ---

func TestGetFound(t *testing.T) {
	q := &mockQuerier{pigeons: []repository.Pigeon{{ID: 42, Name: "Duke", CreatedAt: time.Now()}}}
	h := handler.NewPigeonHandler(q)

	req := httptest.NewRequest(http.MethodGet, "/pigeons/42", nil)
	req.SetPathValue("id", "42")
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestGetNotFound(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	req := httptest.NewRequest(http.MethodGet, "/pigeons/99", nil)
	req.SetPathValue("id", "99")
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetInvalidID(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	req := httptest.NewRequest(http.MethodGet, "/pigeons/abc", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetDBError(t *testing.T) {
	q := &mockQuerier{err: errDB}
	h := handler.NewPigeonHandler(q)

	req := httptest.NewRequest(http.MethodGet, "/pigeons/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

// --- Create ---

func TestCreate(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"Tweety"}`)
	req := httptest.NewRequest(http.MethodPost, "/pigeons", body)
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestCreateEmptyName(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"   "}`)
	req := httptest.NewRequest(http.MethodPost, "/pigeons", body)
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateMissingName(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/pigeons", body)
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateUnknownField(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"Bird","extra":"field"}`)
	req := httptest.NewRequest(http.MethodPost, "/pigeons", body)
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateMalformedJSON(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{bad json}`)
	req := httptest.NewRequest(http.MethodPost, "/pigeons", body)
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateBodyTooLarge(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"` + strings.Repeat("a", 1<<20+1) + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/pigeons", body)
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

func TestCreateTrailingData(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"Tweety"} {"name":"Other"}`)
	req := httptest.NewRequest(http.MethodPost, "/pigeons", body)
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateDBError(t *testing.T) {
	q := &mockQuerier{err: errDB}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"Tweety"}`)
	req := httptest.NewRequest(http.MethodPost, "/pigeons", body)
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
