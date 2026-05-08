package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScheduleManager_SetAndGet(t *testing.T) {
	m := NewScheduleManager()
	entry := ScheduleEntry{JobName: "backup", Expression: "0 2 * * *", Description: "nightly backup"}
	m.Set(entry)

	got, ok := m.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if got.Expression != "0 2 * * *" {
		t.Errorf("expected expression %q, got %q", "0 2 * * *", got.Expression)
	}
}

func TestScheduleManager_GetUnknown(t *testing.T) {
	m := NewScheduleManager()
	_, ok := m.Get("nonexistent")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestScheduleManager_Delete(t *testing.T) {
	m := NewScheduleManager()
	m.Set(ScheduleEntry{JobName: "cleanup", Expression: "@daily"})
	m.Delete("cleanup")
	_, ok := m.Get("cleanup")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestScheduleManager_Snapshot(t *testing.T) {
	m := NewScheduleManager()
	m.Set(ScheduleEntry{JobName: "job1", Expression: "@hourly"})
	m.Set(ScheduleEntry{JobName: "job2", Expression: "@daily"})

	snap := m.Snapshot()
	if len(snap) != 2 {
		t.Errorf("expected 2 entries, got %d", len(snap))
	}
}

func TestHandleSchedules_PostAndGet(t *testing.T) {
	m := NewScheduleManager()

	body, _ := json.Marshal(ScheduleEntry{JobName: "report", Expression: "0 8 * * 1"})
	req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	m.HandleSchedules(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/schedules?job=report", nil)
	rec2 := httptest.NewRecorder()
	m.HandleSchedules(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var got ScheduleEntry
	json.NewDecoder(rec2.Body).Decode(&got)
	if got.Expression != "0 8 * * 1" {
		t.Errorf("unexpected expression: %q", got.Expression)
	}
}

func TestHandleSchedules_GetNotFound(t *testing.T) {
	m := NewScheduleManager()
	req := httptest.NewRequest(http.MethodGet, "/schedules?job=missing", nil)
	rec := httptest.NewRecorder()
	m.HandleSchedules(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandleSchedules_MethodNotAllowed(t *testing.T) {
	m := NewScheduleManager()
	req := httptest.NewRequest(http.MethodPut, "/schedules", nil)
	rec := httptest.NewRecorder()
	m.HandleSchedules(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleSchedules_PostBadBody(t *testing.T) {
	m := NewScheduleManager()
	req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewBufferString("not-json"))
	rec := httptest.NewRecorder()
	m.HandleSchedules(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
