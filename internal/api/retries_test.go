package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRetryManager_RecordAndCount(t *testing.T) {
	rm := NewRetryManager(10)

	rec := rm.Record("backup", "timeout")
	if rec.JobName != "backup" {
		t.Errorf("expected job_name 'backup', got %q", rec.JobName)
	}
	if rec.Attempt != 1 {
		t.Errorf("expected attempt 1, got %d", rec.Attempt)
	}
	if rec.Reason != "timeout" {
		t.Errorf("expected reason 'timeout', got %q", rec.Reason)
	}

	rm.Record("backup", "exit code 1")
	if rm.Count("backup") != 2 {
		t.Errorf("expected count 2, got %d", rm.Count("backup"))
	}
}

func TestRetryManager_CountUnknownJob(t *testing.T) {
	rm := NewRetryManager(10)
	if rm.Count("nonexistent") != 0 {
		t.Error("expected 0 for unknown job")
	}
}

func TestRetryManager_EvictsWhenFull(t *testing.T) {
	rm := NewRetryManager(3)
	rm.Record("job", "a")
	rm.Record("job", "b")
	rm.Record("job", "c")
	rm.Record("job", "d") // should evict "a"

	snap := rm.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 records, got %d", len(snap))
	}
	if snap[0].Reason != "b" {
		t.Errorf("expected oldest surviving reason 'b', got %q", snap[0].Reason)
	}
}

func TestRetryManager_AttemptIncrements(t *testing.T) {
	rm := NewRetryManager(10)

	for i := 1; i <= 5; i++ {
		rec := rm.Record("deploy", "timeout")
		if rec.Attempt != i {
			t.Errorf("expected attempt %d, got %d", i, rec.Attempt)
		}
	}
}

func TestHandleRetries_ReturnsJSON(t *testing.T) {
	rm := NewRetryManager(10)
	rm.Record("sync", "drift")

	req := httptest.NewRequest(http.MethodGet, "/retries", nil)
	w := httptest.NewRecorder()
	rm.handleRetries(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var records []RetryRecord
	if err := json.NewDecoder(w.Body).Decode(&records); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].JobName != "sync" {
		t.Errorf("expected job_name 'sync', got %q", records[0].JobName)
	}
}

func TestHandleRetries_EmptyReturnsArray(t *testing.T) {
	rm := NewRetryManager(10)

	req := httptest.NewRequest(http.MethodGet, "/retries", nil)
	w := httptest.NewRecorder()
	rm.handleRetries(w, req)

	var records []RetryRecord
	if err := json.NewDecoder(w.Body).Decode(&records); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected empty array, got %d records", len(records))
	}
}

func TestHandleRetries_MethodNotAllowed(t *testing.T) {
	rm := NewRetryManager(10)

	req := httptest.NewRequest(http.MethodPost, "/retries", nil)
	w := httptest.NewRecorder()
	rm.handleRetries(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
