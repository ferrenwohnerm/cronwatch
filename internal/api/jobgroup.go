package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

// JobGroupManager tracks logical groupings of jobs.
type JobGroupManager struct {
	mu     sync.RWMutex
	groups map[string]string // job -> group name
}

// NewJobGroupManager creates a new JobGroupManager.
func NewJobGroupManager() *JobGroupManager {
	return &JobGroupManager{
		groups: make(map[string]string),
	}
}

// Set assigns a job to a group.
func (m *JobGroupManager) Set(job, group string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.groups[job] = group
}

// Get returns the group for a job and whether it exists.
func (m *JobGroupManager) Get(job string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	g, ok := m.groups[job]
	return g, ok
}

// Delete removes a job's group assignment.
func (m *JobGroupManager) Delete(job string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.groups, job)
}

// Snapshot returns a copy of all group assignments.
func (m *JobGroupManager) Snapshot() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]string, len(m.groups))
	for k, v := range m.groups {
		out[k] = v
	}
	return out
}

func (m *JobGroupManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	switch r.Method {
	case http.MethodGet:
		if job == "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(m.Snapshot())
			return
		}
		g, ok := m.Get(job)
		if !ok {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"job": job, "group": g})
	case http.MethodPost:
		var body struct {
			Job   string `json:"job"`
			Group string `json:"group"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Job == "" || body.Group == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		m.Set(body.Job, body.Group)
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		if job == "" {
			http.Error(w, "job param required", http.StatusBadRequest)
			return
		}
		m.Delete(job)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
