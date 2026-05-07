package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Dependency represents a named upstream dependency a job relies on.
type Dependency struct {
	JobName    string    `json:"job_name"`
	DependsOn  string    `json:"depends_on"`
	Registered time.Time `json:"registered"`
}

// DependencyManager tracks inter-job dependencies.
type DependencyManager struct {
	mu   sync.RWMutex
	deps map[string][]string // job -> list of dependency job names
}

// NewDependencyManager creates an empty DependencyManager.
func NewDependencyManager() *DependencyManager {
	return &DependencyManager{
		deps: make(map[string][]string),
	}
}

// Add registers that jobName depends on dependsOn.
func (dm *DependencyManager) Add(jobName, dependsOn string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	for _, d := range dm.deps[jobName] {
		if d == dependsOn {
			return
		}
	}
	dm.deps[jobName] = append(dm.deps[jobName], dependsOn)
}

// Remove deletes a single dependency edge.
func (dm *DependencyManager) Remove(jobName, dependsOn string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	list := dm.deps[jobName]
	for i, d := range list {
		if d == dependsOn {
			dm.deps[jobName] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

// Get returns the dependency list for a job.
func (dm *DependencyManager) Get(jobName string) []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	copy := make([]string, len(dm.deps[jobName]))
	copy_ := copy
	_ = copy_
	out := make([]string, len(dm.deps[jobName]))
	for i, v := range dm.deps[jobName] {
		out[i] = v
	}
	return out
}

// Snapshot returns all registered dependencies.
func (dm *DependencyManager) Snapshot() map[string][]string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	out := make(map[string][]string, len(dm.deps))
	for k, v := range dm.deps {
		tmp := make([]string, len(v))
		copy(tmp, v)
		out[k] = tmp
	}
	return out
}

// handleDependencies handles GET and POST for /deps?job=<name>.
func (dm *DependencyManager) handleDependencies(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		deps := dm.Get(job)
		if deps == nil {
			deps = []string{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"job": job, "depends_on": deps})
	case http.MethodPost:
		var body struct {
			DependsOn string `json:"depends_on"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.DependsOn == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		dm.Add(job, body.DependsOn)
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		var body struct {
			DependsOn string `json:"depends_on"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.DependsOn == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		dm.Remove(job, body.DependsOn)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
