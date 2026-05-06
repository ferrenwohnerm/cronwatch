package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

// LabelManager stores arbitrary key-value label pairs per job.
type LabelManager struct {
	mu     sync.RWMutex
	labels map[string]map[string]string
}

// NewLabelManager creates an empty LabelManager.
func NewLabelManager() *LabelManager {
	return &LabelManager{
		labels: make(map[string]map[string]string),
	}
}

// Set stores a label key/value for the given job.
func (lm *LabelManager) Set(job, key, value string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	if _, ok := lm.labels[job]; !ok {
		lm.labels[job] = make(map[string]string)
	}
	lm.labels[job][key] = value
}

// Get returns all labels for a job and whether the job exists.
func (lm *LabelManager) Get(job string) (map[string]string, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	labels, ok := lm.labels[job]
	if !ok {
		return nil, false
	}
	copy := make(map[string]string, len(labels))
	for k, v := range labels {
		copy[k] = v
	}
	return copy, true
}

// Delete removes all labels for a job.
func (lm *LabelManager) Delete(job string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	delete(lm.labels, job)
}

// Snapshot returns a copy of all labels across all jobs.
func (lm *LabelManager) Snapshot() map[string]map[string]string {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	out := make(map[string]map[string]string, len(lm.labels))
	for job, kv := range lm.labels {
		copy := make(map[string]string, len(kv))
		for k, v := range kv {
			copy[k] = v
		}
		out[job] = copy
	}
	return out
}

func handleLabels(lm *LabelManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job := r.URL.Query().Get("job")
		if job == "" {
			http.Error(w, "missing job query parameter", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			labels, _ := lm.Get(job)
			if labels == nil {
				labels = map[string]string{}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(labels)
		case http.MethodPost:
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			for k, v := range body {
				lm.Set(job, k, v)
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			lm.Delete(job)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
