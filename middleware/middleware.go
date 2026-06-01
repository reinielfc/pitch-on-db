package middleware

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"sync/atomic"
	"time"
)

const RequestIDHeader = "X-Request-ID"

// requestIDKey is the context key for the request ID.
type requestIDKey struct{}

// RequestIDFromContext retrieves the request ID stored by the RequestID middleware.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// requestIDPattern allows printable ASCII excluding control characters and
// whitespace, bounded to 128 characters to prevent header bloat and log injection.
var requestIDPattern = regexp.MustCompile(`^[!-~]{1,128}$`)

// responseWriter wraps http.ResponseWriter to capture the status code,
// while forwarding optional interfaces so streaming, websockets, and
// optimised io.Copy keep working.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, fmt.Errorf("hijack not supported by underlying ResponseWriter")
}

func (rw *responseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := rw.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (rw *responseWriter) ReadFrom(src io.Reader) (int64, error) {
	if rf, ok := rw.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(src)
	}
	return io.Copy(rw.ResponseWriter, src)
}

// Logger logs each request: method, path, status, duration.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start))
	})
}

// Recoverer catches panics and returns a 500.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v", rec)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequestID attaches a unique ID to each request via X-Request-ID.
// Uses the incoming header if present and valid, otherwise generates one.
// The ID is propagated on both the response header, the outgoing request header,
// and the request context (retrievable via RequestIDFromContext).
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" || !requestIDPattern.MatchString(id) {
			id = newRequestID()
		}
		w.Header().Set(RequestIDHeader, id)
		r = r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id))
		r.Header.Set(RequestIDHeader, id)
		next.ServeHTTP(w, r)
	})
}

var requestIDSeq atomic.Int64

func newRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fall back to timestamp + counter to avoid collisions on RNG failure.
		return fmt.Sprintf("fallback-%d-%d", time.Now().UnixNano(), requestIDSeq.Add(1))
	}
	return hex.EncodeToString(b)
}
