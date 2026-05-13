package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// JobStatusEntry holds the current known status of a job.
type JobStatusEntry struct {
	Job       string    `json:"job"`
	Status    string    `json:"status"` // "ok", "failing", "unknown"
	Message   string    `json:"message,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// JobStatusManager tracks per-job status strings.
type JobStatusManager struct {
	mu      sync.RWMutex
	entries map[string]JobStatusEntry
}

// NewJobStatusManager returns an initialised JobStatusManager.
func NewJobStatusManager() *JobStatusManager {
	return &JobStatusManager{
		entries: make(map[string]JobStatusEntry),
	}
}

// Set stores a status for the given job.
func (m *JobStatusManager) Set(job, status, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[job] = JobStatusEntry{
		Job:       job,
		Status:    status,
		Message:   message,
		UpdatedAt: time.Now().UTC(),
	}
}

// Get returns the status entry for a job and whether it exists.
func (m *JobStatusManager) Get(job string) (JobStatusEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.entries[job]
	return e, ok
}

// Snapshot returns a copy of all current status entries.
func (m *JobStatusManager) Snapshot() []JobStatusEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]JobStatusEntry, 0, len(m.entries))
	for _, e := range m.entries {
		out = append(out, e)
	}
	return out
}

// HandleJobStatus serves GET/POST /api/jobstatus?job=<name>.
func (m *JobStatusManager) HandleJobStatus(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		job := r.URL.Query().Get("job")
		w.Header().Set("Content-Type", "application/json")
		if job == "" {
			json.NewEncoder(w).Encode(m.Snapshot())
			return
		}
		e, ok := m.Get(job)
		if !ok {
			http.Error(w, `{"error":"job not found"}`, http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(e)
	case http.MethodPost:
		var req struct {
			Job     string `json:"job"`
			Status  string `json:"status"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" || req.Status == "" {
			http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
			return
		}
		m.Set(req.Job, req.Status, req.Message)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
