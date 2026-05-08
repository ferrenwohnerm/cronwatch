package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

// OwnerEntry holds ownership metadata for a cron job.
type OwnerEntry struct {
	Job   string `json:"job"`
	Owner string `json:"owner"`
	Email string `json:"email,omitempty"`
	Team  string `json:"team,omitempty"`
}

// OwnerManager stores and retrieves job ownership information.
type OwnerManager struct {
	mu     sync.RWMutex
	owners map[string]OwnerEntry
}

// NewOwnerManager creates a new OwnerManager.
func NewOwnerManager() *OwnerManager {
	return &OwnerManager{
		owners: make(map[string]OwnerEntry),
	}
}

// Set stores or replaces the owner entry for a job.
func (m *OwnerManager) Set(job string, entry OwnerEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry.Job = job
	m.owners[job] = entry
}

// Get retrieves the owner entry for a job.
func (m *OwnerManager) Get(job string) (OwnerEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.owners[job]
	return e, ok
}

// Delete removes the owner entry for a job.
func (m *OwnerManager) Delete(job string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.owners, job)
}

// Snapshot returns a copy of all owner entries.
func (m *OwnerManager) Snapshot() []OwnerEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]OwnerEntry, 0, len(m.owners))
	for _, e := range m.owners {
		out = append(out, e)
	}
	return out
}

// HandleOwners handles GET and POST requests for job ownership.
func (m *OwnerManager) HandleOwners(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")

	switch r.Method {
	case http.MethodGet:
		if job == "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(m.Snapshot())
			return
		}
		entry, ok := m.Get(job)
		if !ok {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entry)

	case http.MethodPost:
		if job == "" {
			http.Error(w, "missing job parameter", http.StatusBadRequest)
			return
		}
		var entry OwnerEntry
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		m.Set(job, entry)
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
