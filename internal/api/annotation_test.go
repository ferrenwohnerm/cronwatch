package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnnotationManager_AddAndGet(t *testing.T) {
	m := NewAnnotationManager(5)
	m.Add("backup", "first note")
	m.Add("backup", "second note")
	list := m.Get("backup")
	if len(list) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(list))
	}
	if list[0].Note != "first note" || list[1].Note != "second note" {
		t.Errorf("unexpected notes: %+v", list)
	}
}

func TestAnnotationManager_GetUnknown(t *testing.T) {
	m := NewAnnotationManager(5)
	list := m.Get("nonexistent")
	if len(list) != 0 {
		t.Errorf("expected empty slice, got %d items", len(list))
	}
}

func TestAnnotationManager_EvictsWhenFull(t *testing.T) {
	m := NewAnnotationManager(3)
	for i := 0; i < 4; i++ {
		m.Add("job", "note")
	}
	if len(m.Get("job")) != 3 {
		t.Errorf("expected cap of 3")
	}
}

func TestAnnotationManager_Snapshot(t *testing.T) {
	m := NewAnnotationManager(5)
	m.Add("jobA", "note1")
	m.Add("jobB", "note2")
	snap := m.Snapshot()
	if len(snap) != 2 {
		t.Errorf("expected 2 jobs in snapshot, got %d", len(snap))
	}
}

func TestHandleAnnotations_PostAndGet(t *testing.T) {
	m := NewAnnotationManager(10)

	// POST
	body := bytes.NewBufferString(`{"note":"deployed v2"}`)
	req := httptest.NewRequest(http.MethodPost, "/annotations?job=deploy", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	m.handleAnnotations(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	// GET
	req2 := httptest.NewRequest(http.MethodGet, "/annotations?job=deploy", nil)
	rec2 := httptest.NewRecorder()
	m.handleAnnotations(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var list []Annotation
	if err := json.NewDecoder(rec2.Body).Decode(&list); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(list) != 1 || list[0].Note != "deployed v2" {
		t.Errorf("unexpected list: %+v", list)
	}
}

func TestHandleAnnotations_MissingJob(t *testing.T) {
	m := NewAnnotationManager(5)
	req := httptest.NewRequest(http.MethodGet, "/annotations", nil)
	rec := httptest.NewRecorder()
	m.handleAnnotations(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleAnnotations_MethodNotAllowed(t *testing.T) {
	m := NewAnnotationManager(5)
	req := httptest.NewRequest(http.MethodDelete, "/annotations?job=x", nil)
	rec := httptest.NewRecorder()
	m.handleAnnotations(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
