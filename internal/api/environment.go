package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

// EnvironmentManager stores arbitrary key-value environment metadata per job.
type EnvironmentManager struct {
	mu   sync.RWMutex
	data map[string]map[string]string
}

// NewEnvironmentManager creates an initialised EnvironmentManager.
func NewEnvironmentManager() *EnvironmentManager {
	return &EnvironmentManager{
		data: make(map[string]map[string]string),
	}
}

// Set stores a key-value pair for the given job.
func (m *EnvironmentManager) Set(job, key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[job]; !ok {
		m.data[job] = make(map[string]string)
	}
	m.data[job][key] = value
}

// Get returns the environment map for a job and whether it exists.
func (m *EnvironmentManager) Get(job string) (map[string]string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	env, ok := m.data[job]
	if !ok {
		return nil, false
	}
	copy := make(map[string]string, len(env))
	for k, v := range env {
		copy[k] = v
	}
	return copy, true
}

// Delete removes all environment metadata for a job.
func (m *EnvironmentManager) Delete(job string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, job)
}

// Snapshot returns a deep copy of all stored environment data.
func (m *EnvironmentManager) Snapshot() map[string]map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]map[string]string, len(m.data))
	for job, env := range m.data {
		copy := make(map[string]string, len(env))
		for k, v := range env {
			copy[k] = v
		}
		out[job] = copy
	}
	return out
}

// HandleEnvironment serves GET and POST requests for job environment metadata.
func (m *EnvironmentManager) HandleEnvironment(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		env, ok := m.Get(job)
		if !ok {
			env = map[string]string{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(env)

	case http.MethodPost:
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		for k, v := range payload {
			m.Set(job, k, v)
		}
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		m.Delete(job)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
