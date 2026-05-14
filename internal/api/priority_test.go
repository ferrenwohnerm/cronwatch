package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPriorityManager_SetAndGet(t *testing.T) {
	m := NewPriorityManager()
	m.Set("backup", PriorityHigh)
	level, ok := m.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if level != PriorityHigh {
		t.Fatalf("expected high, got %s", level)
	}
}

func TestPriorityManager_GetUnknown(t *testing.T) {
	m := NewPriorityManager()
	_, ok := m.Get("unknown")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestPriorityManager_Delete(t *testing.T) {
	m := NewPriorityManager()
	m.Set("cleanup", PriorityLow)
	m.Delete("cleanup")
	_, ok := m.Get("cleanup")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestPriorityManager_Snapshot(t *testing.T) {
	m := NewPriorityManager()
	m.Set("jobA", PriorityCritical)
	m.Set("jobB", PriorityNormal)
	snap := m.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
	snap["jobA"] = PriorityLow // mutation must not affect manager
	if v, _ := m.Get("jobA"); v != PriorityCritical {
		t.Fatal("snapshot mutation affected manager")
	}
}

func TestHandlePriority_PostAndGet(t *testing.T) {
	m := NewPriorityManager()

	body, _ := json.Marshal(map[string]string{"priority": "critical"})
	req := httptest.NewRequest(http.MethodPost, "/priority?job=nightly", bytes.NewReader(body))
	w := httptest.NewRecorder()
	m.HandlePriority(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/priority?job=nightly", nil)
	w = httptest.NewRecorder()
	m.HandlePriority(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	if result["priority"] != "critical" {
		t.Fatalf("expected critical, got %s", result["priority"])
	}
}

func TestHandlePriority_NotFound(t *testing.T) {
	m := NewPriorityManager()
	req := httptest.NewRequest(http.MethodGet, "/priority?job=missing", nil)
	w := httptest.NewRecorder()
	m.HandlePriority(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandlePriority_MethodNotAllowed(t *testing.T) {
	m := NewPriorityManager()
	req := httptest.NewRequest(http.MethodPatch, "/priority?job=x", nil)
	w := httptest.NewRecorder()
	m.HandlePriority(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandlePriority_MissingJobParam(t *testing.T) {
	m := NewPriorityManager()
	req := httptest.NewRequest(http.MethodGet, "/priority", nil)
	w := httptest.NewRecorder()
	m.HandlePriority(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
