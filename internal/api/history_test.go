package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHistory_RecordAndSnapshot(t *testing.T) {
	h := NewHistory(3)

	h.Record(DriftEvent{JobName: "job1", OccurredAt: time.Now(), Actual: 5 * time.Second})
	h.Record(DriftEvent{JobName: "job2", OccurredAt: time.Now(), Actual: 10 * time.Second})

	snap := h.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 events, got %d", len(snap))
	}
}

func TestHistory_EvictsOldestWhenFull(t *testing.T) {
	h := NewHistory(2)

	h.Record(DriftEvent{JobName: "first"})
	h.Record(DriftEvent{JobName: "second"})
	h.Record(DriftEvent{JobName: "third"})

	snap := h.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 events after eviction, got %d", len(snap))
	}
	if snap[0].JobName != "second" {
		t.Errorf("expected oldest evicted; got %q as first", snap[0].JobName)
	}
	if snap[1].JobName != "third" {
		t.Errorf("expected 'third' as last; got %q", snap[1].JobName)
	}
}

func TestHandleHistory_EmptyReturnsArray(t *testing.T) {
	h := NewHistory(10)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/history", nil)

	handleHistory(h)(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var events []DriftEvent
	if err := json.NewDecoder(w.Body).Decode(&events); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected empty array, got %d events", len(events))
	}
}

func TestHandleHistory_ReturnsRecordedEvents(t *testing.T) {
	h := NewHistory(10)
	h.Record(DriftEvent{JobName: "backup", Actual: 3 * time.Minute})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/history", nil)
	handleHistory(h)(w, r)

	var events []DriftEvent
	json.NewDecoder(w.Body).Decode(&events)
	if len(events) != 1 || events[0].JobName != "backup" {
		t.Errorf("unexpected events: %+v", events)
	}
}

func TestHandleHistory_MethodNotAllowed(t *testing.T) {
	h := NewHistory(10)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/history", nil)

	handleHistory(h)(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
