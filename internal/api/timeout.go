package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// TimeoutEntry holds the configured timeout for a single job.
type TimeoutEntry struct {
	JobName   string        `json:"job_name"`
	Timeout   time.Duration `json:"timeout_seconds"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// TimeoutManager stores per-job timeout thresholds.
type TimeoutManager struct {
	mu      sync.RWMutex
	entries map[string]TimeoutEntry
}

// NewTimeoutManager creates an empty TimeoutManager.
func NewTimeoutManager() *TimeoutManager {
	return &TimeoutManager{
		entries: make(map[string]TimeoutEntry),
	}
}

// Set registers or updates the timeout for a job.
func (m *TimeoutManager) Set(jobName string, d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[jobName] = TimeoutEntry{
		JobName:   jobName,
		Timeout:   d,
		UpdatedAt: time.Now().UTC(),
	}
}

// Get returns the timeout for a job and whether it was found.
func (m *TimeoutManager) Get(jobName string) (TimeoutEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.entries[jobName]
	return e, ok
}

// Delete removes the timeout entry for a job.
func (m *TimeoutManager) Delete(jobName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.entries, jobName)
}

// Snapshot returns a copy of all timeout entries.
func (m *TimeoutManager) Snapshot() []TimeoutEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]TimeoutEntry, 0, len(m.entries))
	for _, e := range m.entries {
		out = append(out, e)
	}
	return out
}

// handleTimeouts serves GET /timeouts and POST /timeouts?job=<name>&seconds=<n>.
func handleTimeouts(m *TimeoutManager) http.HandlerFunc {
	type postBody struct {
		JobName string  `json:"job_name"`
		Seconds float64 `json:"timeout_seconds"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(m.Snapshot())
		case http.MethodPost:
			var body postBody
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.JobName == "" {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}
			m.Set(body.JobName, time.Duration(body.Seconds*float64(time.Second)))
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
