package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// RunLogEntry records a single completed job execution.
type RunLogEntry struct {
	JobName   string        `json:"job_name"`
	StartedAt time.Time     `json:"started_at"`
	FinishedAt time.Time    `json:"finished_at"`
	Duration  time.Duration `json:"duration_ms"`
	Drifted   bool          `json:"drifted"`
}

// RunLogManager stores a bounded list of job run log entries.
type RunLogManager struct {
	mu      sync.RWMutex
	entries []RunLogEntry
	maxSize int
}

// NewRunLogManager creates a RunLogManager with the given capacity.
func NewRunLogManager(maxSize int) *RunLogManager {
	if maxSize <= 0 {
		maxSize = 200
	}
	return &RunLogManager{
		entries: make([]RunLogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Record appends a new entry, evicting the oldest if at capacity.
func (m *RunLogManager) Record(entry RunLogEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.entries) >= m.maxSize {
		m.entries = m.entries[1:]
	}
	m.entries = append(m.entries, entry)
}

// Snapshot returns a copy of all current log entries.
func (m *RunLogManager) Snapshot() []RunLogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copy := make([]RunLogEntry, len(m.entries))
	for i, e := range m.entries {
		copy[i] = e
	}
	return copy
}

// handleRunLog serves GET /runlog, returning all recorded run entries as JSON.
func (m *RunLogManager) handleRunLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	entries := m.Snapshot()
	if entries == nil {
		entries = []RunLogEntry{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
