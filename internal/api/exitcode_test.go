package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExitCodeManager_RecordAndGet(t *testing.T) {
	m := NewExitCodeManager()
	m.Record("backup", 0)

	e, ok := m.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Job != "backup" || e.Code != 0 {
		t.Errorf("unexpected entry: %+v", e)
	}
}

func TestExitCodeManager_GetUnknown(t *testing.T) {
	m := NewExitCodeManager()
	_, ok := m.Get("ghost")
	if ok {
		t.Error("expected no entry for unknown job")
	}
}

func TestExitCodeManager_OverwritesPrevious(t *testing.T) {
	m := NewExitCodeManager()
	m.Record("sync", 1)
	m.Record("sync", 0)

	e, _ := m.Get("sync")
	if e.Code != 0 {
		t.Errorf("expected code 0, got %d", e.Code)
	}
}

func TestExitCodeManager_Snapshot(t *testing.T) {
	m := NewExitCodeManager()
	m.Record("jobA", 0)
	m.Record("jobB", 2)

	snap := m.Snapshot()
	if len(snap) != 2 {
		t.Errorf("expected 2 entries, got %d", len(snap))
	}
}

func TestHandleExitCodes_PostAndGet(t *testing.T) {
	m := NewExitCodeManager()

	body := bytes.NewBufferString(`{"job":"cleanup","code":1}`)
	req := httptest.NewRequest(http.MethodPost, "/exitcodes", body)
	w := httptest.NewRecorder()
	m.handleExitCodes(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/exitcodes", nil)
	w2 := httptest.NewRecorder()
	m.handleExitCodes(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w2.Code)
	}

	var entries []ExitCodeEntry
	if err := json.NewDecoder(w2.Body).Decode(&entries); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(entries) != 1 || entries[0].Job != "cleanup" || entries[0].Code != 1 {
		t.Errorf("unexpected entries: %+v", entries)
	}
}

func TestHandleExitCodes_BadBody(t *testing.T) {
	m := NewExitCodeManager()
	body := bytes.NewBufferString(`not-json`)
	req := httptest.NewRequest(http.MethodPost, "/exitcodes", body)
	w := httptest.NewRecorder()
	m.handleExitCodes(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleExitCodes_MethodNotAllowed(t *testing.T) {
	m := NewExitCodeManager()
	req := httptest.NewRequest(http.MethodDelete, "/exitcodes", nil)
	w := httptest.NewRecorder()
	m.handleExitCodes(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
