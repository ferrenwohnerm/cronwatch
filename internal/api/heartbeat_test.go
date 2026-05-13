package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHeartbeatManager_RecordAndNotStale(t *testing.T) {
	hm := NewHeartbeatManager(5 * time.Second)
	hm.Record("backup")
	stale, _ := hm.IsStale("backup")
	if stale {
		t.Fatal("expected job to be fresh after record")
	}
}

func TestHeartbeatManager_UnknownJobIsStale(t *testing.T) {
	hm := NewHeartbeatManager(5 * time.Second)
	stale, _ := hm.IsStale("unknown")
	if !stale {
		t.Fatal("expected unknown job to be stale")
	}
}

func TestHeartbeatManager_StaleAfterTTL(t *testing.T) {
	hm := NewHeartbeatManager(1 * time.Millisecond)
	hm.Record("cleanup")
	time.Sleep(5 * time.Millisecond)
	stale, missed := hm.IsStale("cleanup")
	if !stale {
		t.Fatal("expected job to be stale after TTL")
	}
	if missed <= 0 {
		t.Fatalf("expected positive missed duration, got %v", missed)
	}
}

func TestHeartbeatManager_Snapshot(t *testing.T) {
	hm := NewHeartbeatManager(10 * time.Second)
	hm.Record("job-a")
	hm.Record("job-b")
	snap := hm.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
}

func TestHandleHeartbeat_Post(t *testing.T) {
	hm := NewHeartbeatManager(10 * time.Second)
	req := httptest.NewRequest(http.MethodPost, "/heartbeat?job=myjob", nil)
	w := httptest.NewRecorder()
	hm.handleHeartbeat(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	stale, _ := hm.IsStale("myjob")
	if stale {
		t.Fatal("expected job to be recorded after POST")
	}
}

func TestHandleHeartbeat_PostMissingJob(t *testing.T) {
	hm := NewHeartbeatManager(10 * time.Second)
	req := httptest.NewRequest(http.MethodPost, "/heartbeat", nil)
	w := httptest.NewRecorder()
	hm.handleHeartbeat(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleHeartbeat_Get(t *testing.T) {
	hm := NewHeartbeatManager(10 * time.Second)
	hm.Record("report")
	req := httptest.NewRequest(http.MethodGet, "/heartbeat", nil)
	w := httptest.NewRecorder()
	hm.handleHeartbeat(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var entries []HeartbeatEntry
	if err := json.NewDecoder(w.Body).Decode(&entries); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(entries) != 1 || entries[0].Job != "report" {
		t.Fatalf("unexpected entries: %+v", entries)
	}
}

func TestHandleHeartbeat_MethodNotAllowed(t *testing.T) {
	hm := NewHeartbeatManager(10 * time.Second)
	req := httptest.NewRequest(http.MethodDelete, "/heartbeat", nil)
	w := httptest.NewRecorder()
	hm.handleHeartbeat(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
