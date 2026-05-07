package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// RetryRecord captures a single retry event for a job.
type RetryRecord struct {
	JobName   string    `json:"job_name"`
	Attempt   int       `json:"attempt"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// RetryManager tracks retry attempts per job.
type RetryManager struct {
	mu      sync.RWMutex
	counts  map[string]int
	records []RetryRecord
	maxRecords int
}

// NewRetryManager creates a RetryManager with the given history capacity.
func NewRetryManager(maxRecords int) *RetryManager {
	if maxRecords <= 0 {
		maxRecords = 200
	}
	return &RetryManager{
		counts:     make(map[string]int),
		maxRecords: maxRecords,
	}
}

// Record increments the retry count for a job and appends a record.
func (r *RetryManager) Record(jobName, reason string) RetryRecord {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.counts[jobName]++
	rec := RetryRecord{
		JobName:   jobName,
		Attempt:   r.counts[jobName],
		Reason:    reason,
		Timestamp: time.Now().UTC(),
	}
	if len(r.records) >= r.maxRecords {
		r.records = r.records[1:]
	}
	r.records = append(r.records, rec)
	return rec
}

// Count returns the total retry count for a job.
func (r *RetryManager) Count(jobName string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.counts[jobName]
}

// Snapshot returns a copy of all retry records.
func (r *RetryManager) Snapshot() []RetryRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]RetryRecord, len(r.records))
	copy(out, r.records)
	return out
}

// handleRetries serves GET /retries — returns all recorded retry events.
func (r *RetryManager) handleRetries(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	snap := r.Snapshot()
	if snap == nil {
		snap = []RetryRecord{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snap)
}
