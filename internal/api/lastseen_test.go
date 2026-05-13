package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLastSeenManager_RecordAndGet(t *testing.T) {
	m := NewLastSeenManager()
	before := time.Now().UTC().Add(-time.Second)
	m.Record("backup")
	after := time.Now().UTC().Add(time.Second)

	ts, ok := m.Get("backup")
	if !ok {
		t.Fatal("expected job to be found")
	}
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v outside expected window [%v, %v]", ts, before, after)
	}
}

func TestLastSeenManager_GetUnknown(t *testing.T) {
	m := NewLastSeenManager()
	_, ok := m.Get("ghost")
	if ok {
		t.Fatal("expected missing job to return false")
	}
}

func TestLastSeenManager_Snapshot(t *testing.T) {
	m := NewLastSeenManager()
	m.Record("jobA")
	m.Record("jobB")

	snap := m.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
	if _, ok := snap["jobA"]; !ok {
		t.Error("expected jobA in snapshot")
	}
}

func newLastSeenRouter(m *LastSeenManager) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterLastSeenRoutes(mux, m)
	return mux
}

func TestHandleLastSeen_GetAll(t *testing.T) {
	m := NewLastSeenManager()
	m.Record("jobX")

	req := httptest.NewRequest(http.MethodGet, "/last-seen", nil)
	w := httptest.NewRecorder()
	newLastSeenRouter(m).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var results []lastSeenResponse
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(results) != 1 || results[0].Job != "jobX" {
		t.Errorf("unexpected results: %+v", results)
	}
}

func TestHandleLastSeen_GetSingleJob(t *testing.T) {
	m := NewLastSeenManager()
	m.Record("deploy")

	req := httptest.NewRequest(http.MethodGet, "/last-seen?job=deploy", nil)
	w := httptest.NewRecorder()
	newLastSeenRouter(m).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result lastSeenResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if result.Job != "deploy" {
		t.Errorf("expected job=deploy, got %s", result.Job)
	}
}

func TestHandleLastSeen_JobNotFound(t *testing.T) {
	m := NewLastSeenManager()
	req := httptest.NewRequest(http.MethodGet, "/last-seen?job=missing", nil)
	w := httptest.NewRecorder()
	newLastSeenRouter(m).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleLastSeen_MethodNotAllowed(t *testing.T) {
	m := NewLastSeenManager()
	req := httptest.NewRequest(http.MethodPost, "/last-seen", nil)
	w := httptest.NewRecorder()
	newLastSeenRouter(m).ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
