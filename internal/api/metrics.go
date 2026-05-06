package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// MetricsSnapshot holds aggregated runtime metrics for all tracked jobs.
type MetricsSnapshot struct {
	Jobs map[string]*JobMetrics `json:"jobs"`
}

// JobMetrics holds per-job execution statistics.
type JobMetrics struct {
	TotalRuns   int           `json:"total_runs"`
	DriftCount  int           `json:"drift_count"`
	LastRunAt   *time.Time    `json:"last_run_at,omitempty"`
	AvgDuration time.Duration `json:"avg_duration_ms"`
}

// Metrics records execution events and exposes an HTTP handler.
type Metrics struct {
	mu   sync.RWMutex
	jobs map[string]*jobStats
}

type jobStats struct {
	totalRuns   int
	driftCount  int
	lastRunAt   *time.Time
	totalMillis int64
}

// NewMetrics creates a new Metrics recorder.
func NewMetrics() *Metrics {
	return &Metrics{jobs: make(map[string]*jobStats)}
}

// Record registers a completed job run. drifted indicates the run was outside its window.
func (m *Metrics) Record(name string, duration time.Duration, drifted bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.jobs[name]
	if !ok {
		s = &jobStats{}
		m.jobs[name] = s
	}
	s.totalRuns++
	if drifted {
		s.driftCount++
	}
	now := time.Now().UTC()
	s.lastRunAt = &now
	s.totalMillis += duration.Milliseconds()
}

// Snapshot returns a point-in-time copy of all metrics.
func (m *Metrics) Snapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	snap := MetricsSnapshot{Jobs: make(map[string]*JobMetrics, len(m.jobs))}
	for name, s := range m.jobs {
		avg := time.Duration(0)
		if s.totalRuns > 0 {
			avg = time.Duration(s.totalMillis/int64(s.totalRuns)) * time.Millisecond
		}
		snap.Jobs[name] = &JobMetrics{
			TotalRuns:   s.totalRuns,
			DriftCount:  s.driftCount,
			LastRunAt:   s.lastRunAt,
			AvgDuration: avg,
		}
	}
	return snap
}

// handleMetrics serves the metrics snapshot as JSON.
func (m *Metrics) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.Snapshot()) //nolint:errcheck
}
