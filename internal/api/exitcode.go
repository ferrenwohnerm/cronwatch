package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// ExitCodeEntry records the exit code of a completed job run.
type ExitCodeEntry struct {
	Job      string    `json:"job"`
	Code     int       `json:"code"`
	RecordedAt time.Time `json:"recorded_at"`
}

// ExitCodeManager stores the most recent exit code per job.
type ExitCodeManager struct {
	mu      sync.RWMutex
	entries map[string]ExitCodeEntry
}

// NewExitCodeManager creates a new ExitCodeManager.
func NewExitCodeManager() *ExitCodeManager {
	return &ExitCodeManager{
		entries: make(map[string]ExitCodeEntry),
	}
}

// Record stores the exit code for the given job.
func (m *ExitCodeManager) Record(job string, code int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[job] = ExitCodeEntry{
		Job:        job,
		Code:       code,
		RecordedAt: time.Now().UTC(),
	}
}

// Get returns the most recent exit code entry for a job and whether it exists.
func (m *ExitCodeManager) Get(job string) (ExitCodeEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.entries[job]
	return e, ok
}

// Snapshot returns a copy of all recorded exit codes.
func (m *ExitCodeManager) Snapshot() []ExitCodeEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]ExitCodeEntry, 0, len(m.entries))
	for _, e := range m.entries {
		out = append(out, e)
	}
	return out
}

// handleExitCodes serves GET /exitcodes and POST /exitcodes.
func (m *ExitCodeManager) handleExitCodes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(m.Snapshot())
	case http.MethodPost:
		var req struct {
			Job  string `json:"job"`
			Code int    `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		m.Record(req.Job, req.Code)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
