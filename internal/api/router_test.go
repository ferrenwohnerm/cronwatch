package api

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/cronwatch/internal/tracker"
)

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	tr := tracker.New()
	logger := log.New(io_discard(), "", 0)
	return NewRouter(tr, logger)
}

// io_discard returns a writer that discards all output, avoiding an import
// of io in the test file.
func io_discard() *nopWriter { return &nopWriter{} }

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestRouter_HealthzRoute(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRouter_StatusRoute(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestRouter_UnknownRouteReturns404(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestRouter_PanicRecovery(t *testing.T) {
	tr := tracker.New()
	logger := log.New(io_discard(), "", 0)

	mux := http.NewServeMux()
	mux.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	var handler http.Handler = mux
	handler = LoggingMiddleware(logger, handler)
	handler = RecoveryMiddleware(logger, handler)
	_ = tr

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 after panic, got %d", rec.Code)
	}
}
