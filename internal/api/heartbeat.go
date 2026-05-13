package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// HeartbeatEntry records the last time a job sent a heartbeat ping.
type HeartbeatEntry struct {
	Job       string    `json:"job"`
	LastSeen  time.Time `json:"last_seen"`
	MissedBy  string    `json:"missed_by,omitempty"`
}

// HeartbeatManager tracks per-job heartbeat timestamps.
type HeartbeatManager struct {
	mu      sync.RWMutex
	beats   map[string]time.Time
	ttl     time.Duration
}

// NewHeartbeatManager creates a HeartbeatManager. Jobs that have not sent a
// heartbeat within ttl are considered stale.
func NewHeartbeatManager(ttl time.Duration) *HeartbeatManager {
	return &HeartbeatManager{
		beats: make(map[string]time.Time),
		ttl:   ttl,
	}
}

// Record stores the current time as the latest heartbeat for job.
func (h *HeartbeatManager) Record(job string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.beats[job] = time.Now()
}

// IsStale returns true if the job has not been seen within the configured TTL.
func (h *HeartbeatManager) IsStale(job string) (bool, time.Duration) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	last, ok := h.beats[job]
	if !ok {
		return true, 0
	}
	age := time.Since(last)
	if age > h.ttl {
		return true, age - h.ttl
	}
	return false, 0
}

// Snapshot returns a copy of all heartbeat entries.
func (h *HeartbeatManager) Snapshot() []HeartbeatEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]HeartbeatEntry, 0, len(h.beats))
	for job, ts := range h.beats {
		e := HeartbeatEntry{Job: job, LastSeen: ts}
		age := time.Since(ts)
		if age > h.ttl {
			e.MissedBy = (age - h.ttl).Round(time.Second).String()
		}
		out = append(out, e)
	}
	return out
}

// handleHeartbeat handles POST /heartbeat?job=<name> and GET /heartbeat.
func (h *HeartbeatManager) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		job := r.URL.Query().Get("job")
		if job == "" {
			http.Error(w, `{"error":"job parameter required"}`, http.StatusBadRequest)
			return
		}
		h.Record(job)
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(h.Snapshot())
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
