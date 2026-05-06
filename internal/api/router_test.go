package api

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/tracker"
)

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	tr := tracker.New()
	h := NewHandler(tr)
	hist := NewHistory(10)
	m := NewMetrics()
	logger := log.New(io.Discard, "", 0)
	return NewRouter(h, hist, m, logger, io.Discard)
}

func io_discard(t *testing.T) io.Writer { //nolint:revive
	t.Helper()
	return io.Discard
}

func TestRouter_HealthzRoute(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRouter_StatusRoute(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRouter_HistoryRoute(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRouter_MetricsRoute(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRouter_MetricsReflectsRecordedData(t *testing.T) {
	tr := tracker.New()
	h := NewHandler(tr)
	hist := NewHistory(10)
	m := NewMetrics()
	m.Record("nightly", 4*time.Second, true)
	logger := log.New(io.Discard, "", 0)
	router := NewRouter(h, hist, m, logger, io.Discard)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if body == "" {
		t.Error("expected non-empty metrics body")
	}
}
