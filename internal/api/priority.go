package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

// PriorityLevel represents the urgency of a cron job.
type PriorityLevel string

const (
	PriorityLow      PriorityLevel = "low"
	PriorityNormal   PriorityLevel = "normal"
	PriorityHigh     PriorityLevel = "high"
	PriorityCritical PriorityLevel = "critical"
)

// PriorityManager stores per-job priority levels.
type PriorityManager struct {
	mu       sync.RWMutex
	priority map[string]PriorityLevel
}

// NewPriorityManager returns an initialised PriorityManager.
func NewPriorityManager() *PriorityManager {
	return &PriorityManager{priority: make(map[string]PriorityLevel)}
}

// Set assigns a priority level to a job.
func (m *PriorityManager) Set(job string, level PriorityLevel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.priority[job] = level
}

// Get returns the priority level for a job and whether it was found.
func (m *PriorityManager) Get(job string) (PriorityLevel, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	level, ok := m.priority[job]
	return level, ok
}

// Delete removes the priority entry for a job.
func (m *PriorityManager) Delete(job string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.priority, job)
}

// Snapshot returns a copy of all priority entries.
func (m *PriorityManager) Snapshot() map[string]PriorityLevel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]PriorityLevel, len(m.priority))
	for k, v := range m.priority {
		out[k] = v
	}
	return out
}

// HandlePriority serves GET and POST /priority?job=<name>.
func (m *PriorityManager) HandlePriority(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		level, ok := m.Get(job)
		if !ok {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"job": job, "priority": string(level)})

	case http.MethodPost:
		var body struct {
			Priority PriorityLevel `json:"priority"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Priority == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		m.Set(job, body.Priority)
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		m.Delete(job)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
