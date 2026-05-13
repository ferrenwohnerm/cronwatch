package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckpointManager_SetAndGet(t *testing.T) {
	m := NewCheckpointManager()
	m.Set("backup", "row-1000")
	e, ok := m.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Checkpoint != "row-1000" {
		t.Errorf("got checkpoint %q, want %q", e.Checkpoint, "row-1000")
	}
	if e.JobName != "backup" {
		t.Errorf("got job_name %q, want %q", e.JobName, "backup")
	}
}

func TestCheckpointManager_GetUnknown(t *testing.T) {
	m := NewCheckpointManager()
	_, ok := m.Get("nonexistent")
	if ok {
		t.Error("expected no entry for unknown job")
	}
}

func TestCheckpointManager_Delete(t *testing.T) {
	m := NewCheckpointManager()
	m.Set("job1", "page-42")
	m.Delete("job1")
	_, ok := m.Get("job1")
	if ok {
		t.Error("expected entry to be deleted")
	}
}

func TestCheckpointManager_Snapshot(t *testing.T) {
	m := NewCheckpointManager()
	m.Set("job-a", "cp-1")
	m.Set("job-b", "cp-2")
	snap := m.Snapshot()
	if len(snap) != 2 {
		t.Errorf("expected 2 entries, got %d", len(snap))
	}
}

func TestHandleCheckpoints_PostAndGet(t *testing.T) {
	m := NewCheckpointManager()

	body := `{"checkpoint":"offset-500"}`
	req := httptest.NewRequest(http.MethodPost, "/checkpoints?job=etl", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	m.HandleCheckpoints(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST: expected 204, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/checkpoints?job=etl", nil)
	rec2 := httptest.NewRecorder()
	m.HandleCheckpoints(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("GET: expected 200, got %d", rec2.Code)
	}
	var e CheckpointEntry
	if err := json.NewDecoder(rec2.Body).Decode(&e); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if e.Checkpoint != "offset-500" {
		t.Errorf("got %q, want %q", e.Checkpoint, "offset-500")
	}
}

func TestHandleCheckpoints_GetAll(t *testing.T) {
	m := NewCheckpointManager()
	m.Set("j1", "v1")
	m.Set("j2", "v2")

	req := httptest.NewRequest(http.MethodGet, "/checkpoints", nil)
	rec := httptest.NewRecorder()
	m.HandleCheckpoints(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []CheckpointEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestHandleCheckpoints_NotFound(t *testing.T) {
	m := NewCheckpointManager()
	req := httptest.NewRequest(http.MethodGet, "/checkpoints?job=ghost", nil)
	rec := httptest.NewRecorder()
	m.HandleCheckpoints(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandleCheckpoints_MethodNotAllowed(t *testing.T) {
	m := NewCheckpointManager()
	req := httptest.NewRequest(http.MethodPatch, "/checkpoints", nil)
	rec := httptest.NewRecorder()
	m.HandleCheckpoints(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
