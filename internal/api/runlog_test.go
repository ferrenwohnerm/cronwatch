package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRunLogManager_RecordAndSnapshot(t *testing.T) {
	mgr := NewRunLogManager(10)
	entry := RunLogEntry{
		JobName:    "backup",
		StartedAt:  time.Now().Add(-5 * time.Second),
		FinishedAt: time.Now(),
		Duration:   5 * time.Second,
		Drifted:    false,
	}
	mgr.Record(entry)
	snap := mgr.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(snap))
	}
	if snap[0].JobName != "backup" {
		t.Errorf("expected job name 'backup', got %s", snap[0].JobName)
	}
}

func TestRunLogManager_EvictsWhenFull(t *testing.T) {
	mgr := NewRunLogManager(3)
	for i := 0; i < 4; i++ {
		mgr.Record(RunLogEntry{JobName: "job", Duration: time.Duration(i) * time.Second})
	}
	snap := mgr.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(snap))
	}
	if snap[0].Duration != time.Second {
		t.Errorf("expected oldest evicted, first duration should be 1s, got %v", snap[0].Duration)
	}
}

func TestHandleRunLog_EmptyReturnsArray(t *testing.T) {
	mgr := NewRunLogManager(10)
	req := httptest.NewRequest(http.MethodGet, "/runlog", nil)
	rec := httptest.NewRecorder()
	mgr.handleRunLog(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result []RunLogEntry
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d entries", len(result))
	}
}

func TestHandleRunLog_ReturnsEntries(t *testing.T) {
	mgr := NewRunLogManager(10)
	mgr.Record(RunLogEntry{JobName: "sync", Drifted: true, Duration: 2 * time.Second})
	req := httptest.NewRequest(http.MethodGet, "/runlog", nil)
	rec := httptest.NewRecorder()
	mgr.handleRunLog(rec, req)
	var result []RunLogEntry
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if !result[0].Drifted {
		t.Errorf("expected drifted=true")
	}
}

func TestHandleRunLog_MethodNotAllowed(t *testing.T) {
	mgr := NewRunLogManager(10)
	req := httptest.NewRequest(http.MethodPost, "/runlog", nil)
	rec := httptest.NewRecorder()
	mgr.handleRunLog(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
