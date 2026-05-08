package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOwnerManager_SetAndGet(t *testing.T) {
	m := NewOwnerManager()
	m.Set("backup", OwnerEntry{Owner: "alice", Email: "alice@example.com", Team: "ops"})

	e, ok := m.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Owner != "alice" || e.Email != "alice@example.com" || e.Team != "ops" {
		t.Errorf("unexpected entry: %+v", e)
	}
	if e.Job != "backup" {
		t.Errorf("expected job field to be set, got %q", e.Job)
	}
}

func TestOwnerManager_GetUnknown(t *testing.T) {
	m := NewOwnerManager()
	_, ok := m.Get("nonexistent")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestOwnerManager_Delete(t *testing.T) {
	m := NewOwnerManager()
	m.Set("cleanup", OwnerEntry{Owner: "bob"})
	m.Delete("cleanup")
	_, ok := m.Get("cleanup")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestOwnerManager_Snapshot(t *testing.T) {
	m := NewOwnerManager()
	m.Set("job-a", OwnerEntry{Owner: "alice"})
	m.Set("job-b", OwnerEntry{Owner: "bob"})

	snap := m.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
}

func TestHandleOwners_PostAndGet(t *testing.T) {
	m := NewOwnerManager()

	body, _ := json.Marshal(OwnerEntry{Owner: "carol", Email: "carol@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/owners?job=report", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	m.HandleOwners(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/owners?job=report", nil)
	rec2 := httptest.NewRecorder()
	m.HandleOwners(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var result OwnerEntry
	json.NewDecoder(rec2.Body).Decode(&result)
	if result.Owner != "carol" {
		t.Errorf("expected owner carol, got %q", result.Owner)
	}
}

func TestHandleOwners_GetNotFound(t *testing.T) {
	m := NewOwnerManager()
	req := httptest.NewRequest(http.MethodGet, "/owners?job=missing", nil)
	rec := httptest.NewRecorder()
	m.HandleOwners(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandleOwners_MethodNotAllowed(t *testing.T) {
	m := NewOwnerManager()
	req := httptest.NewRequest(http.MethodPatch, "/owners", nil)
	rec := httptest.NewRecorder()
	m.HandleOwners(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleOwners_PostMissingJob(t *testing.T) {
	m := NewOwnerManager()
	body, _ := json.Marshal(OwnerEntry{Owner: "dave"})
	req := httptest.NewRequest(http.MethodPost, "/owners", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	m.HandleOwners(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
