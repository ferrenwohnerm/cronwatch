package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newRunLogRouter() (*http.ServeMux, *RunLogManager) {
	mux := http.NewServeMux()
	mgr := NewRunLogManager(50)
	RegisterRunLogRoutes(mux, mgr)
	return mux, mgr
}

func TestRunLogRoute_GetEmpty(t *testing.T) {
	mux, _ := newRunLogRouter()
	req := httptest.NewRequest(http.MethodGet, "/runlog", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json content-type, got %s", ct)
	}
}

func TestRunLogRoute_RecordAndGet(t *testing.T) {
	mux, mgr := newRunLogRouter()
	mgr.Record(RunLogEntry{
		JobName:  "cleanup",
		Duration: 3 * time.Second,
		Drifted:  false,
	})
	req := httptest.NewRequest(http.MethodGet, "/runlog", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var entries []RunLogEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].JobName != "cleanup" {
		t.Errorf("expected job 'cleanup', got %s", entries[0].JobName)
	}
}
