package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJobStatusManager_SetAndGet(t *testing.T) {
	m := NewJobStatusManager()
	m.Set("backup", "ok", "ran fine")
	e, ok := m.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Status != "ok" {
		t.Errorf("expected status ok, got %s", e.Status)
	}
	if e.Message != "ran fine" {
		t.Errorf("unexpected message: %s", e.Message)
	}
}

func TestJobStatusManager_GetUnknown(t *testing.T) {
	m := NewJobStatusManager()
	_, ok := m.Get("ghost")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestJobStatusManager_Snapshot(t *testing.T) {
	m := NewJobStatusManager()
	m.Set("a", "ok", "")
	m.Set("b", "failing", "timeout")
	snap := m.Snapshot()
	if len(snap) != 2 {
		t.Errorf("expected 2 entries, got %d", len(snap))
	}
}

func TestHandleJobStatus_PostAndGet(t *testing.T) {
	m := NewJobStatusManager()

	body := `{"job":"sync","status":"failing","message":"exit 1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/jobstatus", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	m.HandleJobStatus(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/jobstatus?job=sync", nil)
	rec2 := httptest.NewRecorder()
	m.HandleJobStatus(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var e JobStatusEntry
	if err := json.NewDecoder(rec2.Body).Decode(&e); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if e.Status != "failing" {
		t.Errorf("expected failing, got %s", e.Status)
	}
}

func TestHandleJobStatus_GetAll(t *testing.T) {
	m := NewJobStatusManager()
	m.Set("j1", "ok", "")
	m.Set("j2", "unknown", "")

	req := httptest.NewRequest(http.MethodGet, "/api/jobstatus", nil)
	rec := httptest.NewRecorder()
	m.HandleJobStatus(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []JobStatusEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2, got %d", len(entries))
	}
}

func TestHandleJobStatus_NotFound(t *testing.T) {
	m := NewJobStatusManager()
	req := httptest.NewRequest(http.MethodGet, "/api/jobstatus?job=missing", nil)
	rec := httptest.NewRecorder()
	m.HandleJobStatus(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandleJobStatus_MethodNotAllowed(t *testing.T) {
	m := NewJobStatusManager()
	req := httptest.NewRequest(http.MethodDelete, "/api/jobstatus", nil)
	rec := httptest.NewRecorder()
	m.HandleJobStatus(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleJobStatus_BadBody(t *testing.T) {
	m := NewJobStatusManager()
	req := httptest.NewRequest(http.MethodPost, "/api/jobstatus", bytes.NewBufferString(`{bad json`))
	rec := httptest.NewRecorder()
	m.HandleJobStatus(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
