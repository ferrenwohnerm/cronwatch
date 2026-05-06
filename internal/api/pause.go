package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// PauseManager tracks which jobs have monitoring paused.
type PauseManager struct {
	mu     sync.RWMutex
	paused map[string]time.Time // job name -> expiry (zero = indefinite)
}

// NewPauseManager creates a new PauseManager.
func NewPauseManager() *PauseManager {
	return &PauseManager{
		paused: make(map[string]time.Time),
	}
}

// Pause pauses monitoring for a job. If duration is 0, it pauses indefinitely.
func (pm *PauseManager) Pause(job string, duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if duration == 0 {
		pm.paused[job] = time.Time{}
	} else {
		pm.paused[job] = time.Now().Add(duration)
	}
}

// Resume removes a job from the paused set.
func (pm *PauseManager) Resume(job string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.paused, job)
}

// IsPaused returns true if the job is currently paused.
func (pm *PauseManager) IsPaused(job string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	expiry, ok := pm.paused[job]
	if !ok {
		return false
	}
	if expiry.IsZero() {
		return true
	}
	return time.Now().Before(expiry)
}

type pauseRequest struct {
	Job      string `json:"job"`
	Duration string `json:"duration,omitempty"`
}

func (pm *PauseManager) handlePause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req pauseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	var dur time.Duration
	if req.Duration != "" {
		var err error
		dur, err = time.ParseDuration(req.Duration)
		if err != nil {
			http.Error(w, "invalid duration", http.StatusBadRequest)
			return
		}
	}
	pm.Pause(req.Job, dur)
	w.WriteHeader(http.StatusNoContent)
}

func (pm *PauseManager) handleResume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Job string `json:"job"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	pm.Resume(req.Job)
	w.WriteHeader(http.StatusNoContent)
}
