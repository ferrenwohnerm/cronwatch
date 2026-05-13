package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newStreakRouter() (*http.ServeMux, *StreakManager) {
	mux := http.NewServeMux()
	sm := NewStreakManager()
	RegisterStreakRoutes(mux, sm)
	return mux, sm
}

func TestStreakRoute_GetEmpty(t *testing.T) {
	mux, _ := newStreakRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/streaks", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []StreakEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(entries))
	}
}

func TestStreakRoute_RecordAndGet(t *testing.T) {
	mux, sm := newStreakRouter()
	sm.Record("nightly", "success")
	sm.Record("nightly", "success")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/streaks?job=nightly", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var e StreakEntry
	if err := json.NewDecoder(rec.Body).Decode(&e); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if e.SuccessStreak != 2 {
		t.Errorf("expected success streak 2, got %d", e.SuccessStreak)
	}
}

func TestStreakRoute_NotFound(t *testing.T) {
	mux, _ := newStreakRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/streaks?job=missing", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestStreakRoute_MethodNotAllowed(t *testing.T) {
	mux, _ := newStreakRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/streaks", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
