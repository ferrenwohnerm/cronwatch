package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

// TagManager stores arbitrary string tags per job for grouping/filtering.
type TagManager struct {
	mu   sync.RWMutex
	tags map[string][]string
}

// NewTagManager returns an initialised TagManager.
func NewTagManager() *TagManager {
	return &TagManager{tags: make(map[string][]string)}
}

// Set replaces the tag list for a job.
func (tm *TagManager) Set(job string, tags []string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	cp := make([]string, len(tags))
	copy(cp, tags)
	tm.tags[job] = cp
}

// Get returns the tags for a job (nil if unknown).
func (tm *TagManager) Get(job string) []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.tags[job]
}

// Snapshot returns a copy of all job→tags mappings.
func (tm *TagManager) Snapshot() map[string][]string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	out := make(map[string][]string, len(tm.tags))
	for k, v := range tm.tags {
		cp := make([]string, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}

// handleTags handles GET /tags and POST /tags.
func (tm *TagManager) handleTags(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tm.Snapshot())

	case http.MethodPost:
		var req struct {
			Job  string   `json:"job"`
			Tags []string `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		tm.Set(req.Job, req.Tags)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
