package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pitch-on-db/middleware"
)

// --- Logger ---

func TestLoggerPassesThrough(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	h := middleware.Logger(inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", rec.Code)
	}
}

// --- Recoverer ---

func TestRecovererCatchesPanic(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	h := middleware.Recoverer(inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestRecovererPassesThrough(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.Recoverer(inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- RequestID ---

func TestRequestIDGeneratesHeader(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.RequestID(inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if id := rec.Header().Get(middleware.RequestIDHeader); id == "" {
		t.Fatal("expected X-Request-ID to be set")
	}
}

func TestRequestIDPropagatesIncomingHeader(t *testing.T) {
	const existing = "my-request-id"
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.RequestID(inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(middleware.RequestIDHeader, existing)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get(middleware.RequestIDHeader); got != existing {
		t.Fatalf("expected %s, got %s", existing, got)
	}
}

func TestRequestIDRejectsInvalidHeader(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.RequestID(inner)

	invalidIDs := []string{
		strings.Repeat("a", 129), // too long
		"id with spaces",
		"id\twith\ttabs",
		"id\x00null",
	}
	for _, bad := range invalidIDs {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(middleware.RequestIDHeader, bad)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		got := rec.Header().Get(middleware.RequestIDHeader)
		if got == bad {
			t.Fatalf("expected invalid ID %q to be replaced, but it was propagated", bad)
		}
		if got == "" {
			t.Fatalf("expected a generated ID when input %q was invalid, got empty", bad)
		}
	}
}

func TestRequestIDUnique(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.RequestID(inner)

	ids := make(map[string]struct{})
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		id := rec.Header().Get(middleware.RequestIDHeader)
		if id == "" {
			t.Fatalf("iteration %d: expected non-empty X-Request-ID", i)
		}
		if _, dup := ids[id]; dup {
			t.Fatalf("duplicate request ID: %s", id)
		}
		ids[id] = struct{}{}
	}
}

func TestRequestIDPropagatedToContext(t *testing.T) {
var ctxID string
inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
ctxID = middleware.RequestIDFromContext(r.Context())
w.WriteHeader(http.StatusOK)
})
h := middleware.RequestID(inner)

req := httptest.NewRequest(http.MethodGet, "/test", nil)
rec := httptest.NewRecorder()
h.ServeHTTP(rec, req)

responseID := rec.Header().Get(middleware.RequestIDHeader)
if ctxID == "" {
t.Fatal("expected request ID in context, got empty string")
}
if ctxID != responseID {
t.Fatalf("context ID %q does not match response header ID %q", ctxID, responseID)
}
}

func TestRequestIDPropagatedToRequestHeader(t *testing.T) {
var headerID string
inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
headerID = r.Header.Get(middleware.RequestIDHeader)
w.WriteHeader(http.StatusOK)
})
h := middleware.RequestID(inner)

req := httptest.NewRequest(http.MethodGet, "/test", nil)
rec := httptest.NewRecorder()
h.ServeHTTP(rec, req)

responseID := rec.Header().Get(middleware.RequestIDHeader)
if headerID == "" {
t.Fatal("expected request ID on request header, got empty string")
}
if headerID != responseID {
t.Fatalf("request header ID %q does not match response header ID %q", headerID, responseID)
}
}
