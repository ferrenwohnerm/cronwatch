package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleStreaks_ReturnsAllJobs(t *testing.T) {
	sm := NewStreakManager()
	sm.Record("alpha", "success")
	sm.Record("beta", "failure")
	sm.Record("beta", "failure")

	req := httptest.NewRequest(http.MethodGet, "/streaks", nil)
	rec := httptest.NewRecorder()
	sm.handleStreaks(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []StreakEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestHandleStreaks_ContentTypeJSON(t *testing.T) {
	sm := NewStreakManager()
	req := httptest.NewRequest(http.MethodGet, "/streaks", nil)
	rec := httptest.NewRecorder()
	sm.handleStreaks(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestHandleStreaks_MethodNotAllowed(t *testing.T) {
	sm := NewStreakManager()
	req := httptest.NewRequest(http.MethodDelete, "/streaks", nil)
	rec := httptest.NewRecorder()
	sm.handleStreaks(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleStreaks_SingleJobQuery(t *testing.T) {
	sm := NewStreakManager()
	sm.Record("cleanup", "failure")

	req := httptest.NewRequest(http.MethodGet, "/streaks?job=cleanup", nil)
	rec := httptest.NewRecorder()
	sm.handleStreaks(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var e StreakEntry
	if err := json.NewDecoder(rec.Body).Decode(&e); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if e.FailureStreak != 1 {
		t.Errorf("expected failure streak 1, got %d", e.FailureStreak)
	}
	if e.JobName != "cleanup" {
		t.Errorf("expected job name cleanup, got %s", e.JobName)
	}
}
