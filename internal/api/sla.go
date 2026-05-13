package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// SLAEntry holds the SLA configuration for a single job.
type SLAEntry struct {
	JobName     string        `json:"job_name"`
	MaxDuration time.Duration `json:"max_duration_ns"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// SLAManager stores per-job SLA (maximum allowed duration) settings.
type SLAManager struct {
	mu      sync.RWMutex
	entries map[string]SLAEntry
}

// NewSLAManager returns an initialised SLAManager.
func NewSLAManager() *SLAManager {
	return &SLAManager{entries: make(map[string]SLAEntry)}
}

// Set stores or overwrites the SLA for a job.
func (m *SLAManager) Set(jobName string, maxDuration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[jobName] = SLAEntry{
		JobName:     jobName,
		MaxDuration: maxDuration,
		UpdatedAt:   time.Now().UTC(),
	}
}

// Get returns the SLA entry for a job and whether it was found.
func (m *SLAManager) Get(jobName string) (SLAEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.entries[jobName]
	return e, ok
}

// Delete removes the SLA entry for a job.
func (m *SLAManager) Delete(jobName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.entries, jobName)
}

// Snapshot returns a copy of all SLA entries.
func (m *SLAManager) Snapshot() []SLAEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]SLAEntry, 0, len(m.entries))
	for _, e := range m.entries {
		out = append(out, e)
	}
	return out
}

// HandleSLA handles GET / POST / DELETE for /sla?job=<name>.
func (m *SLAManager) HandleSLA(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")

	switch r.Method {
	case http.MethodGet:
		if job == "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(m.Snapshot())
			return
		}
		e, ok := m.Get(job)
		if !ok {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(e)

	case http.MethodPost:
		if job == "" {
			http.Error(w, "missing job parameter", http.StatusBadRequest)
			return
		}
		var body struct {
			MaxDurationNs int64 `json:"max_duration_ns"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.MaxDurationNs <= 0 {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		m.Set(job, time.Duration(body.MaxDurationNs))
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		if job == "" {
			http.Error(w, "missing job parameter", http.StatusBadRequest)
			return
		}
		m.Delete(job)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
