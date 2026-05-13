package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSLAManager_SetAndGet(t *testing.T) {
	m := NewSLAManager()
	m.Set("backup", 5*time.Minute)

	e, ok := m.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.MaxDuration != 5*time.Minute {
		t.Errorf("got %v, want %v", e.MaxDuration, 5*time.Minute)
	}
}

func TestSLAManager_GetUnknown(t *testing.T) {
	m := NewSLAManager()
	_, ok := m.Get("nonexistent")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestSLAManager_Delete(t *testing.T) {
	m := NewSLAManager()
	m.Set("cleanup", time.Minute)
	m.Delete("cleanup")
	_, ok := m.Get("cleanup")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestSLAManager_Snapshot(t *testing.T) {
	m := NewSLAManager()
	m.Set("job-a", time.Minute)
	m.Set("job-b", 2*time.Minute)

	snap := m.Snapshot()
	if len(snap) != 2 {
		t.Errorf("expected 2 entries, got %d", len(snap))
	}
}

func TestHandleSLA_PostAndGet(t *testing.T) {
	m := NewSLAManager()

	body, _ := json.Marshal(map[string]int64{"max_duration_ns": int64(10 * time.Minute)})
	req := httptest.NewRequest(http.MethodPost, "/sla?job=nightly", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	m.HandleSLA(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST: got %d, want 204", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/sla?job=nightly", nil)
	rec2 := httptest.NewRecorder()
	m.HandleSLA(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("GET: got %d, want 200", rec2.Code)
	}
	var entry SLAEntry
	if err := json.NewDecoder(rec2.Body).Decode(&entry); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if entry.MaxDuration != 10*time.Minute {
		t.Errorf("got %v, want %v", entry.MaxDuration, 10*time.Minute)
	}
}

func TestHandleSLA_MethodNotAllowed(t *testing.T) {
	m := NewSLAManager()
	req := httptest.NewRequest(http.MethodPatch, "/sla?job=x", nil)
	rec := httptest.NewRecorder()
	m.HandleSLA(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("got %d, want 405", rec.Code)
	}
}

func TestHandleSLA_PostInvalidBody(t *testing.T) {
	m := NewSLAManager()
	req := httptest.NewRequest(http.MethodPost, "/sla?job=x", bytes.NewBufferString(`{"max_duration_ns": -1}`))
	rec := httptest.NewRecorder()
	m.HandleSLA(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestHandleSLA_GetNotFound(t *testing.T) {
	m := NewSLAManager()
	req := httptest.NewRequest(http.MethodGet, "/sla?job=ghost", nil)
	rec := httptest.NewRecorder()
	m.HandleSLA(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rec.Code)
	}
}
