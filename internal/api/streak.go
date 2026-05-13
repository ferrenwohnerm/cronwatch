package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

// StreakEntry holds consecutive success/failure counts for a job.
type StreakEntry struct {
	JobName        string `json:"job_name"`
	SuccessStreak  int    `json:"success_streak"`
	FailureStreak  int    `json:"failure_streak"`
	LastOutcome    string `json:"last_outcome"` // "success" or "failure"
}

// StreakManager tracks consecutive run outcomes per job.
type StreakManager struct {
	mu      sync.RWMutex
	streaks map[string]*StreakEntry
}

// NewStreakManager returns an initialised StreakManager.
func NewStreakManager() *StreakManager {
	return &StreakManager{streaks: make(map[string]*StreakEntry)}
}

// Record updates the streak counters for the given job.
func (sm *StreakManager) Record(jobName, outcome string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	e, ok := sm.streaks[jobName]
	if !ok {
		e = &StreakEntry{JobName: jobName}
		sm.streaks[jobName] = e
	}
	e.LastOutcome = outcome
	switch outcome {
	case "success":
		e.SuccessStreak++
		e.FailureStreak = 0
	case "failure":
		e.FailureStreak++
		e.SuccessStreak = 0
	}
}

// Get returns the StreakEntry for a job, and whether it exists.
func (sm *StreakManager) Get(jobName string) (StreakEntry, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	e, ok := sm.streaks[jobName]
	if !ok {
		return StreakEntry{}, false
	}
	return *e, true
}

// Snapshot returns a copy of all streak entries.
func (sm *StreakManager) Snapshot() []StreakEntry {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	out := make([]StreakEntry, 0, len(sm.streaks))
	for _, e := range sm.streaks {
		out = append(out, *e)
	}
	return out
}

// handleStreaks serves GET /streaks?job=<name> or GET /streaks for all.
func (sm *StreakManager) handleStreaks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if job := r.URL.Query().Get("job"); job != "" {
		e, ok := sm.Get(job)
		if !ok {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(e)
		return
	}
	json.NewEncoder(w).Encode(sm.Snapshot())
}
