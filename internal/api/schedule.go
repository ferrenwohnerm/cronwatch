package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

// ScheduleEntry holds the cron expression and description for a job.
type ScheduleEntry struct {
	JobName    string `json:"job_name"`
	Expression string `json:"expression"`
	Description string `json:"description,omitempty"`
}

// ScheduleManager stores expected cron schedules for registered jobs.
type ScheduleManager struct {
	mu      sync.RWMutex
	entries map[string]ScheduleEntry
}

// NewScheduleManager creates an empty ScheduleManager.
func NewScheduleManager() *ScheduleManager {
	return &ScheduleManager{
		entries: make(map[string]ScheduleEntry),
	}
}

// Set registers or updates the schedule for a job.
func (m *ScheduleManager) Set(entry ScheduleEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[entry.JobName] = entry
}

// Get returns the schedule for a job, and whether it was found.
func (m *ScheduleManager) Get(jobName string) (ScheduleEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.entries[jobName]
	return e, ok
}

// Delete removes the schedule for a job.
func (m *ScheduleManager) Delete(jobName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.entries, jobName)
}

// Snapshot returns a copy of all schedule entries.
func (m *ScheduleManager) Snapshot() []ScheduleEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]ScheduleEntry, 0, len(m.entries))
	for _, e := range m.entries {
		out = append(out, e)
	}
	return out
}

// HandleSchedules handles GET (list all) and POST (set) schedule requests.
// Route: /schedules?job=<name> for GET single; /schedules for GET all / POST.
func (m *ScheduleManager) HandleSchedules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		job := r.URL.Query().Get("job")
		if job != "" {
			entry, ok := m.Get(job)
			if !ok {
				http.Error(w, "job not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(entry)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m.Snapshot())

	case http.MethodPost:
		var entry ScheduleEntry
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil || entry.JobName == "" {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		m.Set(entry)
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		job := r.URL.Query().Get("job")
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
