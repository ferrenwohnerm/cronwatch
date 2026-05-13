package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// LastSeenManager tracks the last time each job was observed running.
type LastSeenManager struct {
	mu      sync.RWMutex
	entries map[string]time.Time
}

// NewLastSeenManager creates an empty LastSeenManager.
func NewLastSeenManager() *LastSeenManager {
	return &LastSeenManager{
		entries: make(map[string]time.Time),
	}
}

// Record updates the last-seen timestamp for a job to now.
func (m *LastSeenManager) Record(job string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[job] = time.Now().UTC()
}

// Get returns the last-seen time for a job and whether it exists.
func (m *LastSeenManager) Get(job string) (time.Time, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.entries[job]
	return t, ok
}

// Snapshot returns a copy of all last-seen entries.
func (m *LastSeenManager) Snapshot() map[string]time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]time.Time, len(m.entries))
	for k, v := range m.entries {
		out[k] = v
	}
	return out
}

type lastSeenResponse struct {
	Job      string    `json:"job"`
	LastSeen time.Time `json:"last_seen"`
}

// handleLastSeen handles GET /last-seen?job=<name> and GET /last-seen (all jobs).
func (m *LastSeenManager) handleLastSeen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if job := r.URL.Query().Get("job"); job != "" {
		t, ok := m.Get(job)
		if !ok {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(lastSeenResponse{Job: job, LastSeen: t})
		return
	}

	snap := m.Snapshot()
	results := make([]lastSeenResponse, 0, len(snap))
	for job, t := range snap {
		results = append(results, lastSeenResponse{Job: job, LastSeen: t})
	}
	json.NewEncoder(w).Encode(results)
}
