package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMetrics_RecordAndSnapshot(t *testing.T) {
	m := NewMetrics()
	m.Record("backup", 2*time.Second, false)
	m.Record("backup", 3*time.Second, true)

	snap := m.Snapshot()
	job, ok := snap.Jobs["backup"]
	if !ok {
		t.Fatal("expected 'backup' in snapshot")
	}
	if job.TotalRuns != 2 {
		t.Errorf("expected TotalRuns=2, got %d", job.TotalRuns)
	}
	if job.DriftCount != 1 {
		t.Errorf("expected DriftCount=1, got %d", job.DriftCount)
	}
	if job.LastRunAt == nil {
		t.Error("expected LastRunAt to be set")
	}
	expectedAvg := time.Duration(2500) * time.Millisecond
	if job.AvgDuration != expectedAvg {
		t.Errorf("expected AvgDuration=%v, got %v", expectedAvg, job.AvgDuration)
	}
}

func TestMetrics_MultipleJobs(t *testing.T) {
	m := NewMetrics()
	m.Record("jobA", time.Second, false)
	m.Record("jobB", 2*time.Second, true)

	snap := m.Snapshot()
	if len(snap.Jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(snap.Jobs))
	}
}

func TestHandleMetrics_ReturnsJSON(t *testing.T) {
	m := NewMetrics()
	m.Record("cleanup", 500*time.Millisecond, false)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	m.handleMetrics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var snap MetricsSnapshot
	if err := json.NewDecoder(rr.Body).Decode(&snap); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := snap.Jobs["cleanup"]; !ok {
		t.Error("expected 'cleanup' in response")
	}
}

func TestHandleMetrics_EmptyReturnsEmptyJobs(t *testing.T) {
	m := NewMetrics()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	m.handleMetrics(rr, req)

	var snap MetricsSnapshot
	if err := json.NewDecoder(rr.Body).Decode(&snap); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(snap.Jobs) != 0 {
		t.Errorf("expected empty jobs map, got %d entries", len(snap.Jobs))
	}
}

func TestHandleMetrics_MethodNotAllowed(t *testing.T) {
	m := NewMetrics()
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	rr := httptest.NewRecorder()
	m.handleMetrics(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}
