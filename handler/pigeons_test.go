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

	"github.com/jackc/pgx/v5/pgconn"
	"pitch-on-db/handler"
	"pitch-on-db/repository"
)

var errDB = errors.New("db error")
var errBandConflict = &pgconn.PgError{Code: "23505", ConstraintName: "pigeons_band_number_key"}

// fixedTime is a stable time for test assertions.
var fixedTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func strPtr(s string) *string { return &s }

type mockQuerier struct {
	pigeons []repository.ListPigeonsRow
	err     error
}

func (m *mockQuerier) ListPigeons(_ context.Context) ([]repository.ListPigeonsRow, error) {
	return m.pigeons, m.err
}

func (m *mockQuerier) GetPigeon(_ context.Context, id int64) (repository.GetPigeonRow, error) {
	if m.err != nil {
		return repository.GetPigeonRow{}, m.err
	}
	for _, p := range m.pigeons {
		if p.ID == id {
			return repository.GetPigeonRow{ID: p.ID, Name: p.Name, BandNumber: p.BandNumber, BirthDate: p.BirthDate, Sex: p.Sex, CreatedAt: p.CreatedAt}, nil
		}
	}
	return repository.GetPigeonRow{}, sql.ErrNoRows
}

func (m *mockQuerier) CreatePigeon(_ context.Context, arg repository.CreatePigeonParams) (repository.CreatePigeonRow, error) {
	if m.err != nil {
		return repository.CreatePigeonRow{}, m.err
	}
	return repository.CreatePigeonRow{ID: 1, Name: arg.Name, BandNumber: arg.BandNumber, BirthDate: arg.BirthDate, Sex: arg.Sex, CreatedAt: fixedTime}, nil
}

func (m *mockQuerier) UpdatePigeon(_ context.Context, arg repository.UpdatePigeonParams) (repository.UpdatePigeonRow, error) {
	if m.err != nil {
		return repository.UpdatePigeonRow{}, m.err
	}
	for _, p := range m.pigeons {
		if p.ID == arg.ID {
			name := p.Name
			if arg.Name.Valid {
				name = arg.Name.String
			}
			band := p.BandNumber
			if arg.SetBandNumber {
				band = arg.BandNumber
			}
			birth := p.BirthDate
			if arg.SetBirthDate {
				birth = arg.BirthDate
			}
			sex := p.Sex
			if arg.SetSex {
				sex = arg.Sex
			}
			return repository.UpdatePigeonRow{ID: p.ID, Name: name, BandNumber: band, BirthDate: birth, Sex: sex, CreatedAt: p.CreatedAt}, nil
		}
	}
	return repository.UpdatePigeonRow{}, sql.ErrNoRows
}

func (m *mockQuerier) DeletePigeon(_ context.Context, id int64) error {
	return m.err
}

// --- List ---

func TestList(t *testing.T) {
	q := &mockQuerier{pigeons: []repository.ListPigeonsRow{{ID: 1, Name: "Percy", CreatedAt: fixedTime}}}
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
	q := &mockQuerier{pigeons: []repository.ListPigeonsRow{{ID: 42, Name: "Duke", CreatedAt: fixedTime}}}
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

func TestCreateWithAllFields(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"Percy","band_number":"AU2024-001","birth_date":"2024-03-15","sex":"M"}`)
	req := httptest.NewRequest(http.MethodPost, "/pigeons", body)
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp["band_number"] != "AU2024-001" {
		t.Fatalf("expected band_number AU2024-001, got %v", resp["band_number"])
	}
	if resp["sex"] != "M" {
		t.Fatalf("expected sex M, got %v", resp["sex"])
	}
}

func TestCreateInvalidSex(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"Percy","sex":"unknown"}`)
	req := httptest.NewRequest(http.MethodPost, "/pigeons", body)
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateInvalidBirthDate(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"Percy","birth_date":"not-a-date"}`)
	req := httptest.NewRequest(http.MethodPost, "/pigeons", body)
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
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

func TestCreateBandNumberConflict(t *testing.T) {
	q := &mockQuerier{err: errBandConflict}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"Tweety","band_number":"AU2024-001"}`)
	req := httptest.NewRequest(http.MethodPost, "/pigeons", body)
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

// --- Update ---

func TestUpdate(t *testing.T) {
	q := &mockQuerier{pigeons: []repository.ListPigeonsRow{{ID: 1, Name: "Percy", CreatedAt: fixedTime}}}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"Duke"}`)
	req := httptest.NewRequest(http.MethodPatch, "/pigeons/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["name"] != "Duke" {
		t.Fatalf("expected name Duke, got %v", resp["name"])
	}
}

func TestUpdateNotFound(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"Duke"}`)
	req := httptest.NewRequest(http.MethodPatch, "/pigeons/99", body)
	req.SetPathValue("id", "99")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUpdateInvalidID(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"Duke"}`)
	req := httptest.NewRequest(http.MethodPatch, "/pigeons/abc", body)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateEmptyName(t *testing.T) {
	q := &mockQuerier{pigeons: []repository.ListPigeonsRow{{ID: 1, Name: "Percy", CreatedAt: fixedTime}}}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"  "}`)
	req := httptest.NewRequest(http.MethodPatch, "/pigeons/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateDBError(t *testing.T) {
	q := &mockQuerier{err: errDB}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"Duke"}`)
	req := httptest.NewRequest(http.MethodPatch, "/pigeons/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestUpdateBandNumberConflict(t *testing.T) {
	q := &mockQuerier{err: errBandConflict}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"band_number":"AU2024-001"}`)
	req := httptest.NewRequest(http.MethodPatch, "/pigeons/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestUpdateClearBandNumberWithNull(t *testing.T) {
	band := sql.NullString{String: "AU2024-001", Valid: true}
	q := &mockQuerier{pigeons: []repository.ListPigeonsRow{{ID: 1, Name: "Percy", BandNumber: band, CreatedAt: fixedTime}}}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"band_number":null}`)
	req := httptest.NewRequest(http.MethodPatch, "/pigeons/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["band_number"]; ok {
		t.Fatalf("expected band_number to be absent (cleared), got %v", resp["band_number"])
	}
}

func TestCreateEmptyBandNumberTreatedAsNull(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"name":"Percy","band_number":"   "}`)
	req := httptest.NewRequest(http.MethodPost, "/pigeons", body)
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["band_number"]; ok {
		t.Fatalf("expected band_number to be absent (null), got %v", resp["band_number"])
	}
}

// --- Delete ---

func TestDelete(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	req := httptest.NewRequest(http.MethodDelete, "/pigeons/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestDeleteInvalidID(t *testing.T) {
	q := &mockQuerier{}
	h := handler.NewPigeonHandler(q)

	req := httptest.NewRequest(http.MethodDelete, "/pigeons/abc", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteDBError(t *testing.T) {
	q := &mockQuerier{err: errDB}
	h := handler.NewPigeonHandler(q)

	req := httptest.NewRequest(http.MethodDelete, "/pigeons/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

// --- Update: birth_date and sex ---

func TestUpdateSetBirthDate(t *testing.T) {
	q := &mockQuerier{pigeons: []repository.ListPigeonsRow{{ID: 1, Name: "Percy", CreatedAt: fixedTime}}}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"birth_date":"2022-05-10"}`)
	req := httptest.NewRequest(http.MethodPatch, "/pigeons/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["birth_date"] != "2022-05-10" {
		t.Fatalf("expected birth_date 2022-05-10, got %v", resp["birth_date"])
	}
}

func TestUpdateClearBirthDateWithNull(t *testing.T) {
	birth := sql.NullTime{Time: time.Date(2022, 5, 10, 0, 0, 0, 0, time.UTC), Valid: true}
	q := &mockQuerier{pigeons: []repository.ListPigeonsRow{{ID: 1, Name: "Percy", BirthDate: birth, CreatedAt: fixedTime}}}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"birth_date":null}`)
	req := httptest.NewRequest(http.MethodPatch, "/pigeons/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["birth_date"]; ok {
		t.Fatalf("expected birth_date to be absent (cleared), got %v", resp["birth_date"])
	}
}

func TestUpdateInvalidBirthDate(t *testing.T) {
	q := &mockQuerier{pigeons: []repository.ListPigeonsRow{{ID: 1, Name: "Percy", CreatedAt: fixedTime}}}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"birth_date":"not-a-date"}`)
	req := httptest.NewRequest(http.MethodPatch, "/pigeons/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateSetSex(t *testing.T) {
	q := &mockQuerier{pigeons: []repository.ListPigeonsRow{{ID: 1, Name: "Percy", CreatedAt: fixedTime}}}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"sex":"F"}`)
	req := httptest.NewRequest(http.MethodPatch, "/pigeons/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["sex"] != "F" {
		t.Fatalf("expected sex F, got %v", resp["sex"])
	}
}

func TestUpdateClearSexWithNull(t *testing.T) {
	sex := sql.NullString{String: "M", Valid: true}
	q := &mockQuerier{pigeons: []repository.ListPigeonsRow{{ID: 1, Name: "Percy", Sex: sex, CreatedAt: fixedTime}}}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"sex":null}`)
	req := httptest.NewRequest(http.MethodPatch, "/pigeons/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["sex"]; ok {
		t.Fatalf("expected sex to be absent (cleared), got %v", resp["sex"])
	}
}

func TestUpdateInvalidSex(t *testing.T) {
	q := &mockQuerier{pigeons: []repository.ListPigeonsRow{{ID: 1, Name: "Percy", CreatedAt: fixedTime}}}
	h := handler.NewPigeonHandler(q)

	body := strings.NewReader(`{"sex":"X"}`)
	req := httptest.NewRequest(http.MethodPatch, "/pigeons/1", body)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
